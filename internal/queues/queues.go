package queues

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

var (
	ErrEmptyFileName = errors.New("empty file name: URL does not contain a valid file name")
)

type QueueManager interface {

	PauseDownload(ctx context.Context, id int64) error
	ResumeDownload(ctx context.Context, id int64) error
	RetryDownload(ctx context.Context, id int64) error
	CreateDownload(ctx context.Context, url, fileName string, queueID int64) error
	DeleteDownload(ctx context.Context, id int64) error

	CreateQueue(ctx context.Context, createQueueParams state.CreateQueueParams) error
	DeleteQueue(ctx context.Context, id int64) error
	ListQueue(ctx context.Context) ([]state.Queue, error)
	EditQueue(ctx context.Context, arg state.UpdateQueueParams) error

	DownloadFailed(ctx context.Context, id int64) error
	DownloadCompleted(ctx context.Context, id int64) error
	UpsertChunks(ctx context.Context, status downloads.DownloadStatus) error
	ListDownloadsWithQueueName(ctx context.Context) ([]state.ListDownloadsWithQueueNameRow, error)
}

type queueManager struct {
	queries            *state.Queries
	inProgressHandlers map[int64]downloads.DownloadHandler
	queueLimiters      map[int64]*bandwidthlimit.Limiter
	mu                 sync.RWMutex
}

func New(db *sql.DB) (QueueManager, error) {
	qm := &queueManager{
		queries:            state.New(db),
		inProgressHandlers: make(map[int64]downloads.DownloadHandler),
		queueLimiters:      make(map[int64]*bandwidthlimit.Limiter),
	}

	if err := qm.init(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize QueueManager: %w", err)
	}

	return qm, nil
}

func (q *queueManager) init(ctx context.Context) error {

	queues, err := q.queries.ListQueues(ctx)
	if err != nil {
		slog.Error("failed to list queues during initialization", "error", err)
		return fmt.Errorf("failed to list queues during initialization: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for _, queue := range queues {
		if queue.MaxBandwidth.Valid {
			q.queueLimiters[queue.ID] = bandwidthlimit.NewLimiter(&queue.MaxBandwidth.Int64)
		} else {
			q.queueLimiters[queue.ID] = bandwidthlimit.NewLimiter(nil)
		}
	}

	inProgressDownloads, err := q.queries.GetDownloadsByStatus(ctx, string(downloads.StateInProgress))
	if err != nil {
		slog.Error("failed to get in-progress downloads during initialization", "error", err)
		return fmt.Errorf("failed to get in-progress downloads during initialization: %w", err)
	}

	for _, download := range inProgressDownloads {
		downloadChunks, err := q.queries.GetDownloadChunksByDownloadID(ctx, download.ID)
		if err != nil {
			slog.Error("failed to get download chunks for in-progress download", "downloadID", download.ID, "error", err)
			return fmt.Errorf("failed to get download chunks for in-progress download: %w", err)
		}

		limiter, ok := q.queueLimiters[download.QueueID]
		if !ok {
			slog.Error("limiter not found for queue during initialization", "queueID", download.QueueID)
			return fmt.Errorf("limiter not found for queue %d", download.QueueID)
		}

		handler, err := downloads.NewDownloadHandler(download, downloadChunks, limiter)
		if err != nil {
			slog.Error("failed to initilize download handler", "error", err)
			return err
		}
		q.inProgressHandlers[download.ID] = handler

		if err := handler.Start(); err != nil {
			slog.Error("failed to start in-progress download handler", "downloadID", download.ID, "error", err)
			return fmt.Errorf("failed to start in-progress download handler: %w", err)
		}

		slog.Info("resumed in-progress download during initialization", "downloadID", download.ID)
	}

	slog.Info("initialization completed successfully")
	return nil
}
