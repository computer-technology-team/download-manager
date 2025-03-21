package bandwidthlimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const DefaultBandwidth int64 = 1_000_000

type Limiter struct {
	mu        sync.Mutex
	limiter   *rate.Limiter
	unlimited bool
}

func NewLimiter(bandwidthBytesPS *int64) *Limiter {
	l := &Limiter{}

	if bandwidthBytesPS == nil || *bandwidthBytesPS <= 0 {
		l.SetUnlimited()
	} else {
		l.SetBandwidth(*bandwidthBytesPS)
	}

	return l
}

func (l *Limiter) SetBandwidth(bytesPerSecond int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if bytesPerSecond <= 0 {
		l.setUnlimitedLocked()
		return
	}

	l.limiter = rate.NewLimiter(rate.Limit(bytesPerSecond), max(int(bytesPerSecond), 1<<16))
	l.unlimited = false
}

func (l *Limiter) SetUnlimited() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.setUnlimitedLocked()
}

func (l *Limiter) setUnlimitedLocked() {
	l.limiter = rate.NewLimiter(rate.Inf, 1)
	l.unlimited = true
}

func (l *Limiter) Allow(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.unlimited {
		return true
	}

	return l.limiter.AllowN(time.Now(), n)
}

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

func (l *Limiter) Reserve(n int) *rate.Reservation {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.unlimited {
		return l.limiter.ReserveN(time.Now(), 0) 
	}

	return l.limiter.ReserveN(time.Now(), n)
}

func (l *Limiter) IsUnlimited() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.unlimited
}
