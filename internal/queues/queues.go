package queues

import (
	"context"
	"database/sql"

	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type QueueManager interface {
	PauseDownload(ctx context.Context, id int64) error
	ResumeDownload(ctx context.Context, id int64) error
	RetryDownload(ctx context.Context, id int64) error

	CreateDownload(ctx context.Context) error
	ListDownloads(ctx context.Context) ([]state.Download, error)
	DeleteDownload(ctx context.Context) error

	CreateQueue(ctx context.Context) error
	DeleteQueue(ctx context.Context) error
	ListQueue(ctx context.Context) ([]state.Queue, error)
	EditQueue(ctx context.Context, id int64) error
}

type queueManager struct {
	queries *state.Queries
}

// CreateDownload implements QueueManager.
func (q queueManager) CreateDownload(ctx context.Context) error {
	panic("unimplemented")
}

// CreateQueue implements QueueManager.
func (q queueManager) CreateQueue(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteDownload implements QueueManager.
func (q queueManager) DeleteDownload(ctx context.Context) error {
	panic("unimplemented")
}

// DeleteQueue implements QueueManager.
func (q queueManager) DeleteQueue(ctx context.Context) error {
	panic("unimplemented")
}

// EditQueue implements QueueManager.
func (q queueManager) EditQueue(ctx context.Context, id int64) error {
	panic("unimplemented")
}

// ListDownloads implements QueueManager.
func (q queueManager) ListDownloads(ctx context.Context) ([]state.Download, error) {
	panic("unimplemented")
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
