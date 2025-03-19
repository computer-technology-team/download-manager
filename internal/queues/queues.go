package queues

import (
	"context"
	"database/sql"

	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type QueueManager interface {
	PauseDownload(ctx context.Context, id int64) error
	ResumeDownload(ctx context.Context, id int64) error
	RetryDownload(ctx context.Context, id int64) error

	CreateDownload(ctx context.Context) error
	ListDownloads(ctx context.Context) ([]state.Download, error)
	DeleteDownload(ctx context.Context, id int64) error

	CreateQueue(ctx context.Context) error
	DeleteQueue(ctx context.Context, id int64) error
	ListQueue(ctx context.Context) ([]state.Queue, error)
	EditQueue(ctx context.Context, id int64, arg state.UpdateQueueParams) error
}

type queueManager struct {
	queries            *state.Queries
	inProgressHandlers map[int64]downloads.DownloadHandler
}

// CreateDownload implements QueueManager.
func (q queueManager) CreateDownload(ctx context.Context) error {
	panic("unimplemented")
}

// CreateQueue implements QueueManager.
func (q queueManager) CreateQueue(ctx context.Context) error {
	panic("unimplemented")
}

func (q queueManager) DeleteDownload(ctx context.Context, id int64) error {
	q.inProgressHandlers[id].Pause()
	current_download, e := q.queries.GetDownload(ctx, id)
	if e != nil {
		return e
	}
	queue_id := current_download.QueueID
	nextDownload, err := q.queries.GetPausedDownloadByQueueID(ctx, queue_id)
	if err == nil {
		q.queries.SetDownloadState(ctx, state.SetDownloadStateParams{State: "IN_PROGRESS", ID: nextDownload.ID})
	}
	q.queries.DeleteDownload(ctx, id)
	return nil
}

// DeleteQueue implements QueueManager.
func (q queueManager) DeleteQueue(ctx context.Context, id int64) error {
	q.queries.DeleteQueue(ctx, id)

	return nil
}

func (q queueManager) EditQueue(ctx context.Context, id int64, arg state.UpdateQueueParams) error {
	q.queries.UpdateQueue(ctx, arg)
	return nil
}

func (q queueManager) ListDownloads(ctx context.Context) ([]state.Download, error) {
	return q.queries.ListDownloads(ctx)
}

// PauseDownload implements QueueManager.
func (q queueManager) PauseDownload(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// ResumeDownload implements QueueManager.
func (q queueManager) ResumeDownload(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// RetryDownload implements QueueManager.
func (q queueManager) RetryDownload(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// ListQueue implements QueueManager.
func (q queueManager) ListQueue(ctx context.Context) ([]state.Queue, error) {
	return q.queries.ListQueues(ctx)
}

func New(db *sql.DB) QueueManager {
	return queueManager{queries: state.New(db)}
}
