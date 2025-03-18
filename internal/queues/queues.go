package queues

import (
	"database/sql"
	"sync"
	"time"
)

type Queue struct {
	ID            int64
	Name          string
	Directory     string
	MaxBandwidth  sql.NullInt64
	DownloadStart sql.NullString
	DownloadEnd   sql.NullString
	RetryLimit    int64
	Downloads     []*Download
	Mutex         sync.Mutex
}

type Download struct {
	ID     int
	URL    string
	Status string
}

// NewQueue initializes a new queue.
func NewQueue(name, directory string, maxBandwidth, retryLimit int, start, end time.Time) *Queue {
	return &Queue{
		Name:          name,
		Directory:     directory,
		MaxBandwidth:  sql.NullInt64{Int64: int64(maxBandwidth), Valid: true},
		RetryLimit:    int64(retryLimit),
		DownloadStart: sql.NullString{String: start.Format(time.RFC3339), Valid: true},
		DownloadEnd:   sql.NullString{String: end.Format(time.RFC3339), Valid: true},
		Downloads:     []*Download{},
	}
}

// AddDownload adds a new download task to the queue.
func (q *Queue) AddDownload(url string) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	download := &Download{
		ID:     len(q.Downloads) + 1,
		URL:    url,
		Status: "pending",
	}
	q.Downloads = append(q.Downloads, download)
}
