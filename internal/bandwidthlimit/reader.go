package bandwidthlimit

import (
	"context"
	"errors"
	"io"
)

type LimitedReader struct {
	reader  io.Reader
	limiter *Limiter
	ctx     context.Context
}

func NewLimitedReader(ctx context.Context, reader io.Reader, limiter *Limiter) *LimitedReader {
	return &LimitedReader{
		reader:  reader,
		limiter: limiter,
		ctx:     ctx,
	}
}

func (r *LimitedReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)

	if n > 0 {
		waitErr := r.limiter.Wait(r.ctx, n)
		if waitErr != nil {
			return n, errors.Join(waitErr, err)
		}
	}

	return n, err
}
