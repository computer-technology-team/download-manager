package bandwidthlimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// DefaultBandwidth is the default bandwidth limit in bytes per second (1 MB/s)
const DefaultBandwidth int64 = 1_000_000

// Limiter provides bandwidth limiting functionality for downloads
type Limiter struct {
	mu      sync.Mutex
	limiter *rate.Limiter
	// unlimited is true when there's no bandwidth limit
	unlimited bool
}

// NewLimiter creates a new bandwidth limiter with the specified limit in bytes per second
// If bandwidthBytesPS is nil or <= 0, it creates an unlimited limiter
func NewLimiter(bandwidthBytesPS *int64) *Limiter {
	l := &Limiter{}

	if bandwidthBytesPS == nil || *bandwidthBytesPS <= 0 {
		l.SetUnlimited()
	} else {
		l.SetBandwidth(*bandwidthBytesPS)
	}

	return l
}

// SetBandwidth updates the bandwidth limit in bytes per second
func (l *Limiter) SetBandwidth(bytesPerSecond int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if bytesPerSecond <= 0 {
		l.setUnlimitedLocked()
		return
	}

	// Convert bytes per second to tokens per second
	// Each token represents 1 byte
	l.limiter = rate.NewLimiter(rate.Limit(bytesPerSecond), int(bytesPerSecond))
	l.unlimited = false
}

// SetUnlimited removes any bandwidth limit
func (l *Limiter) SetUnlimited() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.setUnlimitedLocked()
}

// setUnlimitedLocked sets the limiter to unlimited mode (must be called with lock held)
func (l *Limiter) setUnlimitedLocked() {
	// Use a very high limit to effectively make it unlimited
	l.limiter = rate.NewLimiter(rate.Inf, 1)
	l.unlimited = true
}

// Allow checks if n bytes can be transferred without exceeding the bandwidth limit
// It returns immediately with true if n bytes can be transferred, or false if the limit would be exceeded
func (l *Limiter) Allow(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.unlimited {
		return true
	}

	return l.limiter.AllowN(time.Now(), n)
}

// Wait blocks until n bytes can be transferred without exceeding the bandwidth limit
// It returns an error if the context is canceled
func (l *Limiter) Wait(ctx context.Context, n int) error {
	l.mu.Lock()

	if l.unlimited {
		l.mu.Unlock()
		return nil
	}

	limiter := l.limiter
	l.mu.Unlock()

	return limiter.WaitN(ctx, n)
}

// Reserve returns a Reservation that indicates how long the caller must wait before n bytes can be transferred
func (l *Limiter) Reserve(n int) *rate.Reservation {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.unlimited {
		return l.limiter.ReserveN(time.Now(), 0) // Zero delay reservation
	}

	return l.limiter.ReserveN(time.Now(), n)
}

// IsUnlimited returns true if the limiter has no bandwidth limit
func (l *Limiter) IsUnlimited() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.unlimited
}
