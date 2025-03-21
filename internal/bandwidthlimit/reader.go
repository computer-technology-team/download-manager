package bandwidthlimit

import (
	"context"
	"io"
)

// LimitedReader wraps an io.Reader with bandwidth limiting capabilities
type LimitedReader struct {
	reader  io.Reader
	limiter *Limiter
	ctx     context.Context
}

// NewLimitedReader creates a new bandwidth-limited reader
func NewLimitedReader(ctx context.Context, reader io.Reader, limiter *Limiter) *LimitedReader {
	return &LimitedReader{
		reader:  reader,
		limiter: limiter,
		ctx:     ctx,
	}
}

// Read implements the io.Reader interface with bandwidth limiting
func (r *LimitedReader) Read(p []byte) (n int, err error) {
	// First read from the underlying reader
	n, err = r.reader.Read(p)

	// If we read some data, wait according to our rate limit
	if n > 0 {
		// Wait until we're allowed to transfer n bytes
		waitErr := r.limiter.Wait(r.ctx, n)
		if waitErr != nil {
			// Context canceled or deadline exceeded
			return n, waitErr
		}
	}

	return n, err
}
