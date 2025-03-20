package downloads

type DownloadState string

const (
	StateInProgress DownloadState = "IN_PROGRESS"
	StatePaused     DownloadState = "PAUSED"
	StateCompleted  DownloadState = "COMPLETED"
	StateFailed     DownloadState = "FAILED"
)

type DownloadHandler interface {
	Start() error
	Pause() error
	Cancel() error
}

type DownloadStatus struct {
	ID                 int64
	ProgressPercentage float32
	Speed              float32
	State              DownloadState
}
