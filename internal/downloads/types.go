package downloads

import "github.com/computer-technology-team/download-manager.git/internal/state"

type DownloadState string

const (
	StateInProgress DownloadState = "IN_PROGRESS"
	StatePaused     DownloadState = "PAUSED"
	StateCompleted  DownloadState = "COMPLETED"
	StateFailed     DownloadState = "FAILED"
	StatePending    DownloadState = "PENDING"
)

type DownloadHandler interface {
	Start() error
	Pause() error
	Cancel() error
}

type DownloadStatus struct {
	ID                 int64
	ProgressPercentage float64
	Speed              float64
	State              DownloadState
	DownloadChuncks    []state.DownloadChunk
}
