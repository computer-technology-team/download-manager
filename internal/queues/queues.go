package queues

import "github.com/computer-technology-team/download-manager.git/internal/state"

type QueueManager interface {
	PauseDownload(id int64) error
	ResumeDownload(id int64) error
	RetryDownload(id int64) error

	CreateDownload() error
	ListDownloads() ([]state.Download, error)
	DeleteDownload() error

	CreateQueue() error
	DeleteQueue() error
	ListQueue() ([]state.Queue, error)
	EditQueue(id int64) error
}

func New() QueueManager {
	//do your init
	return nil
}
