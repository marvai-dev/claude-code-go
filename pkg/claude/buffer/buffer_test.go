package buffer

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestLimitedBuffer_Write(t *testing.T) {
	tests := []struct {
		name       string
		maxSize    int64
		writeData  []string
		wantSize   int64
		wantTrunc  bool
		wantSuffix bool
	}{
		{
			name:       "within limit",
			maxSize:    100,
			writeData:  []string{"hello", "world"},
			wantSize:   10,
			wantTrunc:  false,
			wantSuffix: false,
		},
		{
			name:       "exceeds limit",
			maxSize:    5,
			writeData:  []string{"hello", "world"},
			wantSize:   5,
			wantTrunc:  true,
			wantSuffix: true,
		},
		{
			name:       "exactly at limit",
			maxSize:    10,
			writeData:  []string{"hello", "world"},
			wantSize:   10,
			wantTrunc:  false,
			wantSuffix: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := NewLimitedBuffer(tt.maxSize, "[TRUNCATED]")
			
			for _, data := range tt.writeData {
				lb.Write([]byte(data))
			}

			if lb.Size() != tt.wantSize {
				t.Errorf("Size() = %d, want %d", lb.Size(), tt.wantSize)
			}

			if lb.Truncated() != tt.wantTrunc {
				t.Errorf("Truncated() = %v, want %v", lb.Truncated(), tt.wantTrunc)
			}

			result := lb.String()
			hasSuffix := strings.Contains(result, "[TRUNCATED]")
			if hasSuffix != tt.wantSuffix {
				t.Errorf("Contains truncation suffix = %v, want %v", hasSuffix, tt.wantSuffix)
			}
		})
	}
}

func TestLimitedBuffer_Reset(t *testing.T) {
	lb := NewLimitedBuffer(5, "[TRUNCATED]")
	
	// Write data that exceeds limit
	lb.Write([]byte("hello world"))
	
	// Verify it's truncated
	if !lb.Truncated() {
		t.Error("Expected buffer to be truncated")
	}
	
	// Reset and verify
	lb.Reset()
	
	if lb.Size() != 0 {
		t.Errorf("Size after reset = %d, want 0", lb.Size())
	}
	
	if lb.Truncated() {
		t.Error("Buffer should not be truncated after reset")
	}
	
	if lb.String() != "" {
		t.Errorf("String after reset = %q, want empty", lb.String())
	}
}

func TestBufferManager_NewBuffers(t *testing.T) {
	config := &Config{
		MaxStdoutSize:    1024,
		MaxStderrSize:    512,
		TruncationSuffix: "[TRUNCATED]",
	}
	
	bm := NewBufferManager(config)
	
	stdout := bm.NewStdoutBuffer()
	stderr := bm.NewStderrBuffer()
	
	// Test stdout buffer has correct config
	stdout.Write(make([]byte, 2000)) // Exceed limit
	if !stdout.Truncated() {
		t.Error("Stdout buffer should be truncated")
	}
	if stdout.Size() != 1024 {
		t.Errorf("Stdout buffer size = %d, want 1024", stdout.Size())
	}
	
	// Test stderr buffer has correct config
	stderr.Write(make([]byte, 1000)) // Exceed limit
	if !stderr.Truncated() {
		t.Error("Stderr buffer should be truncated")
	}
	if stderr.Size() != 512 {
		t.Errorf("Stderr buffer size = %d, want 512", stderr.Size())
	}
}

func TestBufferManager_CopyWithTimeout(t *testing.T) {
	bm := NewBufferManager(&Config{BufferTimeout: 100 * time.Millisecond})
	
	tests := []struct {
		name    string
		src     io.Reader
		wantErr bool
	}{
		{
			name:    "normal copy",
			src:     strings.NewReader("hello world"),
			wantErr: false,
		},
		{
			name:    "slow reader with timeout", 
			src:     &slowReader{delay: 200 * time.Millisecond},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer
			
			err := bm.CopyWithTimeout(ctx, &buf, tt.src)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyWithTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSafeReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxBytes int64
		wantRead int
		wantErr  error
	}{
		{
			name:     "within limit",
			input:    "hello",
			maxBytes: 10,
			wantRead: 5,
			wantErr:  nil,
		},
		{
			name:     "exceeds limit",
			input:    "hello world",
			maxBytes: 5,
			wantRead: 5,
			wantErr:  io.ErrUnexpectedEOF,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := NewSafeReader(strings.NewReader(tt.input), tt.maxBytes)
			
			buf := make([]byte, len(tt.input))
			n, err := io.ReadFull(sr, buf)
			
			if n != tt.wantRead {
				t.Errorf("Read bytes = %d, want %d", n, tt.wantRead)
			}
			
			if err != tt.wantErr {
				t.Errorf("Error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.MaxStdoutSize <= 0 {
		t.Error("MaxStdoutSize should be positive")
	}
	
	if config.MaxStderrSize <= 0 {
		t.Error("MaxStderrSize should be positive")
	}
	
	if config.BufferTimeout <= 0 {
		t.Error("BufferTimeout should be positive")
	}
	
	if !config.EnableTruncation {
		t.Error("EnableTruncation should be true by default")
	}
	
	if config.TruncationSuffix == "" {
		t.Error("TruncationSuffix should not be empty")
	}
}

// Helper type for testing timeout functionality
type slowReader struct {
	delay time.Duration
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	time.Sleep(sr.delay)
	return 0, io.EOF
}

func BenchmarkLimitedBuffer_Write(b *testing.B) {
	lb := NewLimitedBuffer(1024*1024, "[TRUNCATED]") // 1MB
	data := make([]byte, 1024) // 1KB chunks
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.Write(data)
		if i%1000 == 0 {
			lb.Reset() // Reset periodically to avoid hitting limit
		}
	}
}

func BenchmarkLimitedBuffer_Concurrent(b *testing.B) {
	lb := NewLimitedBuffer(1024*1024, "[TRUNCATED]")
	data := make([]byte, 100)
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.Write(data)
		}
	})
}