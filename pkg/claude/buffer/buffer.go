package buffer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// Config holds buffer configuration options
type Config struct {
	// MaxStdoutSize is the maximum size for stdout buffer in bytes
	MaxStdoutSize int64
	// MaxStderrSize is the maximum size for stderr buffer in bytes
	MaxStderrSize int64
	// BufferTimeout is the maximum time to wait for buffer operations
	BufferTimeout time.Duration
	// EnableTruncation allows buffers to be truncated when they exceed limits
	EnableTruncation bool
	// TruncationSuffix is added when content is truncated
	TruncationSuffix string
}

// DefaultConfig returns sensible default buffer configuration
func DefaultConfig() *Config {
	return &Config{
		MaxStdoutSize:     10 * 1024 * 1024, // 10MB
		MaxStderrSize:     1 * 1024 * 1024,  // 1MB
		BufferTimeout:     30 * time.Second,
		EnableTruncation:  true,
		TruncationSuffix:  "\n[... output truncated due to size limit ...]",
	}
}

// LimitedBuffer is a bytes.Buffer with size and timeout limits
type LimitedBuffer struct {
	buf      bytes.Buffer
	maxSize  int64
	truncSuf string
	mu       sync.RWMutex
	written  int64
	truncated bool
}

// NewLimitedBuffer creates a new limited buffer
func NewLimitedBuffer(maxSize int64, truncationSuffix string) *LimitedBuffer {
	return &LimitedBuffer{
		maxSize:  maxSize,
		truncSuf: truncationSuffix,
	}
}

// Write implements io.Writer with size limits
func (lb *LimitedBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	originalLen := len(p)
	
	if lb.written >= lb.maxSize {
		lb.truncated = true
		return originalLen, nil // Silently discard additional writes
	}

	remainingSpace := lb.maxSize - lb.written
	if int64(len(p)) > remainingSpace {
		// Truncate the write
		p = p[:remainingSpace]
		lb.truncated = true
	}

	n, err = lb.buf.Write(p)
	lb.written += int64(n)

	return originalLen, err // Return original length to avoid write errors
}

// Bytes returns the buffer contents, adding truncation suffix if needed
func (lb *LimitedBuffer) Bytes() []byte {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	data := lb.buf.Bytes()
	if lb.truncated && lb.truncSuf != "" {
		result := make([]byte, 0, len(data)+len(lb.truncSuf))
		result = append(result, data...)
		result = append(result, []byte(lb.truncSuf)...)
		return result
	}
	return data
}

// String returns the buffer contents as string
func (lb *LimitedBuffer) String() string {
	return string(lb.Bytes())
}

// Size returns the current buffer size
func (lb *LimitedBuffer) Size() int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.written
}

// Truncated returns whether the buffer was truncated
func (lb *LimitedBuffer) Truncated() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.truncated
}

// Reset resets the buffer
func (lb *LimitedBuffer) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.buf.Reset()
	lb.written = 0
	lb.truncated = false
}

// BufferManager manages multiple buffers with configuration
type BufferManager struct {
	config *Config
}

// NewBufferManager creates a new buffer manager
func NewBufferManager(config *Config) *BufferManager {
	if config == nil {
		config = DefaultConfig()
	}
	return &BufferManager{config: config}
}

// NewStdoutBuffer creates a new stdout buffer with configured limits
func (bm *BufferManager) NewStdoutBuffer() *LimitedBuffer {
	return NewLimitedBuffer(bm.config.MaxStdoutSize, bm.config.TruncationSuffix)
}

// NewStderrBuffer creates a new stderr buffer with configured limits
func (bm *BufferManager) NewStderrBuffer() *LimitedBuffer {
	return NewLimitedBuffer(bm.config.MaxStderrSize, bm.config.TruncationSuffix)
}

// CopyWithTimeout copies from src to dst with timeout and context
func (bm *BufferManager) CopyWithTimeout(ctx context.Context, dst io.Writer, src io.Reader) error {
	ctx, cancel := context.WithTimeout(ctx, bm.config.BufferTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(dst, src)
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("buffer copy timeout: %w", ctx.Err())
	}
}

// SafeReader wraps a reader with size limits
type SafeReader struct {
	reader   io.Reader
	maxBytes int64
	read     int64
}

// NewSafeReader creates a reader that stops after maxBytes
func NewSafeReader(reader io.Reader, maxBytes int64) *SafeReader {
	return &SafeReader{
		reader:   reader,
		maxBytes: maxBytes,
	}
}

// Read implements io.Reader with size limits
func (sr *SafeReader) Read(p []byte) (n int, err error) {
	if sr.read >= sr.maxBytes {
		return 0, io.EOF
	}

	remaining := sr.maxBytes - sr.read
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err = sr.reader.Read(p)
	sr.read += int64(n)
	return n, err
}