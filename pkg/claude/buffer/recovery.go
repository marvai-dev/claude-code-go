package buffer

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// RecoveryConfig defines recovery behavior for buffer operations
type RecoveryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration
	// FallbackBufferSize is the reduced buffer size to try on failure
	FallbackBufferSize int64
	// EnableGracefulDegradation allows falling back to streaming
	EnableGracefulDegradation bool
	// CrashOnPanic determines if panics should crash or be recovered
	CrashOnPanic bool
}

// DefaultRecoveryConfig returns sensible defaults for recovery
func DefaultRecoveryConfig() *RecoveryConfig {
	return &RecoveryConfig{
		MaxRetries:                3,
		RetryDelay:                time.Second,
		FallbackBufferSize:        1024 * 1024, // 1MB fallback
		EnableGracefulDegradation: true,
		CrashOnPanic:              false,
	}
}

// ResilientBuffer provides automatic recovery and fallback mechanisms
type ResilientBuffer struct {
	mu             sync.RWMutex
	primaryBuffer  *LimitedBuffer
	fallbackBuffer *LimitedBuffer
	config         *RecoveryConfig
	metrics        *Metrics
	retryCount     int
	usingFallback  bool
	lastError      error
}

// NewResilientBuffer creates a buffer with recovery capabilities
func NewResilientBuffer(maxSize int64, truncationSuffix string, recoveryConfig *RecoveryConfig) *ResilientBuffer {
	if recoveryConfig == nil {
		recoveryConfig = DefaultRecoveryConfig()
	}
	
	return &ResilientBuffer{
		primaryBuffer:  NewLimitedBuffer(maxSize, truncationSuffix),
		fallbackBuffer: NewLimitedBuffer(recoveryConfig.FallbackBufferSize, truncationSuffix+" [FALLBACK]"),
		config:         recoveryConfig,
		metrics:        NewMetrics(),
	}
}

// Write implements io.Writer with automatic recovery
func (rb *ResilientBuffer) Write(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	// Recover from panics if configured
	if !rb.config.CrashOnPanic {
		defer func() {
			if r := recover(); r != nil {
				rb.metrics.RecordError()
				rb.lastError = fmt.Errorf("panic during write: %v", r)
				err = rb.lastError
				n = 0
			}
		}()
	}
	
	// Try primary buffer first
	if !rb.usingFallback {
		n, err = rb.primaryBuffer.Write(p)
		if err == nil {
			rb.metrics.RecordWrite(int64(n))
			return n, nil
		}
		
		// Primary buffer failed, consider fallback
		rb.lastError = err
		rb.metrics.RecordError()
		
		if rb.shouldUseFallback() {
			rb.switchToFallback()
		}
	}
	
	// Use fallback buffer
	if rb.usingFallback {
		n, err = rb.fallbackBuffer.Write(p)
		if err != nil {
			rb.metrics.RecordError()
			rb.lastError = err
		} else {
			rb.metrics.RecordWrite(int64(n))
		}
		return n, err
	}
	
	return 0, rb.lastError
}

// shouldUseFallback determines if we should switch to fallback buffer
func (rb *ResilientBuffer) shouldUseFallback() bool {
	if !rb.config.EnableGracefulDegradation {
		return false
	}
	
	rb.retryCount++
	return rb.retryCount >= rb.config.MaxRetries
}

// switchToFallback switches to using the fallback buffer
func (rb *ResilientBuffer) switchToFallback() {
	rb.usingFallback = true
	// Copy existing data from primary to fallback if possible
	if primaryData := rb.primaryBuffer.Bytes(); len(primaryData) > 0 {
		rb.fallbackBuffer.Write(primaryData)
	}
}

// Bytes returns the current buffer contents
func (rb *ResilientBuffer) Bytes() []byte {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	if rb.usingFallback {
		return rb.fallbackBuffer.Bytes()
	}
	return rb.primaryBuffer.Bytes()
}

// String returns the buffer contents as string
func (rb *ResilientBuffer) String() string {
	return string(rb.Bytes())
}

// Truncated returns whether the buffer was truncated
func (rb *ResilientBuffer) Truncated() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	
	if rb.usingFallback {
		return rb.fallbackBuffer.Truncated()
	}
	return rb.primaryBuffer.Truncated()
}

// IsUsingFallback returns whether the fallback buffer is active
func (rb *ResilientBuffer) IsUsingFallback() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.usingFallback
}

// GetLastError returns the last error encountered
func (rb *ResilientBuffer) GetLastError() error {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.lastError
}

// Reset resets both buffers and recovery state
func (rb *ResilientBuffer) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	rb.primaryBuffer.Reset()
	rb.fallbackBuffer.Reset()
	rb.usingFallback = false
	rb.retryCount = 0
	rb.lastError = nil
}

// GetMetrics returns buffer metrics
func (rb *ResilientBuffer) GetMetrics() Stats {
	return rb.metrics.GetStats()
}

// ResilientCopy performs io.Copy with automatic retry and recovery
type ResilientCopy struct {
	config  *RecoveryConfig
	metrics *Metrics
}

// NewResilientCopy creates a new resilient copy operation
func NewResilientCopy(config *RecoveryConfig) *ResilientCopy {
	if config == nil {
		config = DefaultRecoveryConfig()
	}
	
	return &ResilientCopy{
		config:  config,
		metrics: NewMetrics(),
	}
}

// Copy performs resilient copy with retries and fallbacks
func (rc *ResilientCopy) Copy(ctx context.Context, dst io.Writer, src io.Reader) error {
	var lastErr error
	
	for attempt := 0; attempt <= rc.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(rc.config.RetryDelay):
			}
		}
		
		// Attempt the copy
		_, err := rc.attemptCopy(ctx, dst, src)
		if err == nil {
			return nil // Success
		}
		
		lastErr = err
		rc.metrics.RecordError()
		
		// If context is cancelled, don't retry
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	
	return fmt.Errorf("copy failed after %d attempts: %w", rc.config.MaxRetries+1, lastErr)
}

// attemptCopy performs a single copy attempt
func (rc *ResilientCopy) attemptCopy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	// Use a channel to handle the copy operation
	done := make(chan struct{})
	var n int64
	var err error
	
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil && !rc.config.CrashOnPanic {
				err = fmt.Errorf("panic during copy: %v", r)
			}
		}()
		
		n, err = io.Copy(dst, src)
	}()
	
	select {
	case <-done:
		if err == nil {
			rc.metrics.RecordWrite(n)
		}
		return n, err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// GetMetrics returns copy operation metrics
func (rc *ResilientCopy) GetMetrics() Stats {
	return rc.metrics.GetStats()
}

// CircuitBreaker prevents cascade failures in buffer operations
type CircuitBreaker struct {
	mu                sync.RWMutex
	failureThreshold  int
	resetTimeout      time.Duration
	consecutiveFailures int
	lastFailureTime   time.Time
	state             CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// StateClosed allows operations to proceed
	StateClosed CircuitState = iota
	// StateOpen blocks operations
	StateOpen
	// StateHalfOpen allows limited operations to test recovery
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
	}
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(operation func() error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}
	
	err := operation()
	cb.recordResult(err)
	return err
}

// canExecute determines if an operation can be executed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if err != nil {
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()
		
		if cb.consecutiveFailures >= cb.failureThreshold {
			cb.state = StateOpen
		}
	} else {
		cb.consecutiveFailures = 0
		cb.state = StateClosed
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}