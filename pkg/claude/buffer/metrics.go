package buffer

import (
	"sync"
	"time"
)

// Metrics tracks buffer usage statistics
type Metrics struct {
	mu                  sync.RWMutex
	TotalBytesWritten   int64
	TotalTruncations    int64
	TotalTimeouts       int64
	MaxBufferSizeSeen   int64
	AverageWriteSize    float64
	LastOperationTime   time.Time
	OperationCount      int64
	ErrorCount          int64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		LastOperationTime: time.Now(),
	}
}

// RecordWrite records a write operation
func (m *Metrics) RecordWrite(bytesWritten int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.TotalBytesWritten += bytesWritten
	m.OperationCount++
	m.LastOperationTime = time.Now()
	
	// Update average
	m.AverageWriteSize = float64(m.TotalBytesWritten) / float64(m.OperationCount)
	
	// Track max buffer size
	if bytesWritten > m.MaxBufferSizeSeen {
		m.MaxBufferSizeSeen = bytesWritten
	}
}

// RecordTruncation records a buffer truncation
func (m *Metrics) RecordTruncation() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalTruncations++
}

// RecordTimeout records a timeout event
func (m *Metrics) RecordTimeout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalTimeouts++
}

// RecordError records an error event
func (m *Metrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorCount++
}

// GetStats returns current statistics
func (m *Metrics) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return Stats{
		TotalBytesWritten: m.TotalBytesWritten,
		TotalTruncations:  m.TotalTruncations,
		TotalTimeouts:     m.TotalTimeouts,
		MaxBufferSizeSeen: m.MaxBufferSizeSeen,
		AverageWriteSize:  m.AverageWriteSize,
		LastOperationTime: m.LastOperationTime,
		OperationCount:    m.OperationCount,
		ErrorCount:        m.ErrorCount,
	}
}

// Stats represents buffer usage statistics
type Stats struct {
	TotalBytesWritten int64
	TotalTruncations  int64
	TotalTimeouts     int64
	MaxBufferSizeSeen int64
	AverageWriteSize  float64
	LastOperationTime time.Time
	OperationCount    int64
	ErrorCount        int64
}

// MonitoredBuffer wraps LimitedBuffer with metrics
type MonitoredBuffer struct {
	*LimitedBuffer
	metrics *Metrics
}

// NewMonitoredBuffer creates a buffer with metrics tracking
func NewMonitoredBuffer(maxSize int64, truncationSuffix string) *MonitoredBuffer {
	return &MonitoredBuffer{
		LimitedBuffer: NewLimitedBuffer(maxSize, truncationSuffix),
		metrics:       NewMetrics(),
	}
}

// Write implements io.Writer with metrics tracking
func (mb *MonitoredBuffer) Write(p []byte) (n int, err error) {
	wasTruncated := mb.Truncated()
	n, err = mb.LimitedBuffer.Write(p)
	
	mb.metrics.RecordWrite(int64(len(p)))
	
	// Check if this write caused truncation
	if !wasTruncated && mb.Truncated() {
		mb.metrics.RecordTruncation()
	}
	
	if err != nil {
		mb.metrics.RecordError()
	}
	
	return n, err
}

// GetMetrics returns the buffer's metrics
func (mb *MonitoredBuffer) GetMetrics() Stats {
	return mb.metrics.GetStats()
}

// Enhanced BufferManager with monitoring and recovery
type EnhancedBufferManager struct {
	*BufferManager
	globalMetrics *Metrics
	healthCheck   *HealthChecker
}

// NewEnhancedBufferManager creates a buffer manager with monitoring
func NewEnhancedBufferManager(config *Config) *EnhancedBufferManager {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &EnhancedBufferManager{
		BufferManager: NewBufferManager(config),
		globalMetrics: NewMetrics(),
		healthCheck:   NewHealthChecker(),
	}
}

// NewMonitoredStdoutBuffer creates a monitored stdout buffer
func (ebm *EnhancedBufferManager) NewMonitoredStdoutBuffer() *MonitoredBuffer {
	return NewMonitoredBuffer(ebm.config.MaxStdoutSize, ebm.config.TruncationSuffix)
}

// NewMonitoredStderrBuffer creates a monitored stderr buffer
func (ebm *EnhancedBufferManager) NewMonitoredStderrBuffer() *MonitoredBuffer {
	return NewMonitoredBuffer(ebm.config.MaxStderrSize, ebm.config.TruncationSuffix)
}

// GetGlobalMetrics returns aggregated metrics across all buffers
func (ebm *EnhancedBufferManager) GetGlobalMetrics() Stats {
	return ebm.globalMetrics.GetStats()
}

// HealthChecker monitors buffer system health
type HealthChecker struct {
	mu               sync.RWMutex
	lastHealthCheck  time.Time
	consecutiveErrors int
	isHealthy        bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		lastHealthCheck: time.Now(),
		isHealthy:       true,
	}
}

// CheckHealth performs a health check
func (hc *HealthChecker) CheckHealth(metrics Stats) HealthStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	hc.lastHealthCheck = time.Now()
	
	// Check various health indicators
	status := HealthStatus{
		IsHealthy:    true,
		LastCheck:    hc.lastHealthCheck,
		Issues:       []string{},
		Suggestions:  []string{},
	}
	
	// Check error rate
	if metrics.OperationCount > 0 {
		errorRate := float64(metrics.ErrorCount) / float64(metrics.OperationCount)
		if errorRate > 0.1 { // More than 10% error rate
			status.IsHealthy = false
			status.Issues = append(status.Issues, "High error rate detected")
			status.Suggestions = append(status.Suggestions, "Consider increasing buffer sizes or timeout values")
		}
	}
	
	// Check truncation rate
	if metrics.OperationCount > 0 {
		truncationRate := float64(metrics.TotalTruncations) / float64(metrics.OperationCount)
		if truncationRate > 0.2 { // More than 20% truncation rate
			status.Issues = append(status.Issues, "High truncation rate detected")
			status.Suggestions = append(status.Suggestions, "Consider increasing buffer limits or using streaming")
		}
	}
	
	// Check timeout frequency
	if metrics.TotalTimeouts > 0 {
		status.Issues = append(status.Issues, "Timeouts detected")
		status.Suggestions = append(status.Suggestions, "Consider increasing timeout duration")
	}
	
	hc.isHealthy = status.IsHealthy
	return status
}

// IsHealthy returns current health status
func (hc *HealthChecker) IsHealthy() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.isHealthy
}

// HealthStatus represents the health of the buffer system
type HealthStatus struct {
	IsHealthy    bool
	LastCheck    time.Time
	Issues       []string
	Suggestions  []string
}