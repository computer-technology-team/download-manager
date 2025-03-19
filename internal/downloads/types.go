package downloads

import "context"

type DownloadState string

const (
	StateInProgress DownloadState = "IN_PROGRESS"
	StatePaused     DownloadState = "PAUSED"
	StateCompleted  DownloadState = "COMPLETED"
	StateFailed     DownloadState = "FAILED"
)

type DownloadHandler interface {
	Start(ctx context.Context) error
	Pause() error
}

type DownloadStatus struct {
	ProgressPercentage float32
	Speed              float32 // bytes per second
	State              DownloadState
}
