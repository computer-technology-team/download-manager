package queues

type DownloadHandler interface {
	Start() error

	Pause() error

	Resume() error

	Cancel() error

	Retry() error

	Status() DownloadStatus
}
