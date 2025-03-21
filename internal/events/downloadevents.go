package events

type DownloadFailedEvent struct {
	ID    int64
	Error error
}
