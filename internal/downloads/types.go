package downloads

import "context"

type DownloadHandler interface {
	Start(ctx context.Context) error
	Pause() error
}