package events

type DownloadFailedEvent struct {
	ID    int64
	URL   string
	Error error
}
