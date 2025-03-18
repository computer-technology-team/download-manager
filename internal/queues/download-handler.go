package queues

type DownloadHandler interface {
	Start() error

	Pause() error

	Resume() error

	Cancel() error

	Retry() error

	Status() DownloadStatus
}

type DownloadStatus struct {
	ProgressRate float32 // percent
	Speed        float32 // bytes per second
	State        DownloadState
}

type DownloadState string

const (
	StateInitialized DownloadState = "initialized"
	StateDownloading DownloadState = "downloading"
	StatePaused      DownloadState = "paused"
	StateCompleted   DownloadState = "completed"
	StateCanceled    DownloadState = "canceled"
	StateFailed      DownloadState = "failed"
)
