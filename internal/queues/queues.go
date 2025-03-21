package queues

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"sync"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

// Common errors
var (
	ErrEmptyFileName = errors.New("empty file name: URL does not contain a valid file name")
)

type QueueManager interface {
	PauseDownload(ctx context.Context, id int64) error
	ResumeDownload(ctx context.Context, id int64) error
	RetryDownload(ctx context.Context, id int64) error

	ListDownloadsWithQueueName(ctx context.Context) ([]state.ListDownloadsWithQueueNameRow, error)
	CreateDownload(ctx context.Context, url, fileName string, queueID int64) error
	DeleteDownload(ctx context.Context, id int64) error

	CreateQueue(ctx context.Context, createQueueParams state.CreateQueueParams) error
	DeleteQueue(ctx context.Context, id int64) error
	ListQueue(ctx context.Context) ([]state.Queue, error)
	EditQueue(ctx context.Context, arg state.UpdateQueueParams) error
}

type queueManager struct {
	queries            *state.Queries
	inProgressHandlers map[int64]downloads.DownloadHandler
	queueLimiters      map[int64]*bandwidthlimit.Limiter
	mu                 sync.RWMutex // Mutex to protect inProgressHandlers
}

// Helper function to set download state
func (q *queueManager) setDownloadState(ctx context.Context, id int64, downloadState string) error {
	param := state.SetDownloadStateParams{State: downloadState, ID: id}
	_, err := q.queries.SetDownloadState(ctx, param)
	if err != nil {
		slog.Error("failed to set download state", "state", downloadState, "downloadID", id, "error", err)
		return err
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.DownloadStateChanged,
		Payload:   param,
	}

	return nil
}

// Helper function to start the next download of a queue if it has capacity
func (q *queueManager) startNextDownloadIfPossible(ctx context.Context, queueID int64) error {
	var activeDownloads int64 = 0

	// Count the number of active downloads for the given queueID
	q.mu.RLock()
	for id := range q.inProgressHandlers {
		download, err := q.queries.GetDownload(ctx, id)
		if err != nil {
			q.mu.RUnlock()
			slog.Error("failed to get download details", "downloadID", id, "error", err)
			return fmt.Errorf("failed to get download details: %w", err)
		}
		if download.QueueID == queueID {
			activeDownloads++
		}
	}
	q.mu.RUnlock()

	// Get the queue's MaxConcurrent limit from the database
	queue, err := q.queries.GetQueue(ctx, queueID)
	if err != nil {
		slog.Error("failed to get queue details", "queueID", queueID, "error", err)
		return fmt.Errorf("failed to get queue details: %w", err)
	}

	// If the queue is full, do nothing and return nil
	if activeDownloads >= queue.MaxConcurrent {
		slog.Info("queue is full, cannot start next download", "queueID", queueID, "activeDownloads", activeDownloads, "maxConcurrent", queue.MaxConcurrent)
		return nil
	}

	// Get the next pending download for the queue
	nextDownload, err := q.queries.GetPendingDownloadByQueueID(ctx, queueID)
	if err != nil {
		slog.Error("failed to get pending download by queue ID", "queueID", queueID, "error", err)
		return err
	}

	// Resume the next download using the existing ResumeDownload method
	if err := q.ResumeDownload(ctx, nextDownload.ID); err != nil {
		slog.Error("failed to resume download", "downloadID", nextDownload.ID, "error", err)
		return fmt.Errorf("failed to resume download: %w", err)
	}

	slog.Info("started next download", "downloadID", nextDownload.ID)
	return nil
}

func (q *queueManager) CreateDownload(ctx context.Context, downloadURL, fileName string, queueID int64) error {
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		slog.Error("failed to parse download URL", "url", downloadURL, "error", err)
		return fmt.Errorf("failed to parse download URL: %w", err)
	}

	if fileName == "" {
		lastPathSegment := path.Base(parsedURL.Path)

		if lastPathSegment == "" || lastPathSegment == "." || lastPathSegment == "/" {
			slog.Error("empty file name in URL", "url", downloadURL)
			return ErrEmptyFileName
		}

		fileName = lastPathSegment
	}

	queue, err := q.queries.GetQueue(ctx, queueID)
	if err != nil {
		slog.Error("failed to get queue from database", "queueID", queueID, "error", err)
		return fmt.Errorf("failed to get queue: %w", err)
	}

	createDownloadParams := state.CreateDownloadParams{
		QueueID:  queueID,
		Url:      downloadURL,
		SavePath: path.Join(queue.Directory, fileName),
		State:    string(downloads.StatePending),
		Retries:  0,
	}

	download, err := q.queries.CreateDownload(ctx, createDownloadParams)
	if err != nil {
		slog.Error("failed to create download", "params", createDownloadParams, "error", err)
		return fmt.Errorf("failed to create download: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.DownloadCreated,
		Payload: state.ListDownloadsWithQueueNameRow{
			ID:        download.ID,
			QueueID:   download.QueueID,
			Url:       download.Url,
			SavePath:  download.SavePath,
			State:     download.State,
			Retries:   download.Retries,
			QueueName: queue.Name,
		},
	}

	slog.Info("download created successfully", "downloadID", download.ID)

	if err := q.startNextDownloadIfPossible(ctx, queueID); err != nil {
		return err
	}

	return nil
}

func (q *queueManager) CreateQueue(ctx context.Context, createQueueParams state.CreateQueueParams) error {
	queue, err := q.queries.CreateQueue(ctx, createQueueParams)
	if err != nil {
		slog.Error("failed to create queue", "params", createQueueParams, "error", err)
		return fmt.Errorf("failed to create queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueCreated,
		Payload:   queue,
	}

	slog.Info("queue created successfully", "queueID", queue.ID)
	return nil
}

func (q *queueManager) DeleteDownload(ctx context.Context, id int64) error {
	q.mu.RLock()
	handler, ok := q.inProgressHandlers[id]
	q.mu.RUnlock()

	if !ok {
		return fmt.Errorf("download handler not found for ID %d", id)
	}

	if err := handler.Pause(); err != nil {
		slog.Error("failed to pause download handler", "downloadID", id, "error", err)
		return fmt.Errorf("failed to pause download handler: %w", err)
	}

	currentDownload, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}
	queueID := currentDownload.QueueID // we need the queueID to for starting a new download

	if err := q.queries.DeleteDownload(ctx, id); err != nil {
		slog.Error("failed to delete download", "downloadID", id, "error", err)
		return fmt.Errorf("failed to delete download: %w", err)
	}

	q.mu.Lock()
	delete(q.inProgressHandlers, id)
	q.mu.Unlock()

	slog.Info("download deleted successfully", "downloadID", id)

	if err := q.startNextDownloadIfPossible(ctx, queueID); err != nil {
		return err
	}

	return nil
}

func (q *queueManager) DeleteQueue(ctx context.Context, id int64) error {
	err := q.queries.DeleteQueue(ctx, id)
	if err != nil {
		slog.Error("failed to delete queue", "queueID", id, "error", err)
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueDeleted,
		Payload:   id,
	}

	slog.Info("queue deleted successfully", "queueID", id)
	return nil
}

func (q *queueManager) EditQueue(ctx context.Context, arg state.UpdateQueueParams) error {
	queue, err := q.queries.UpdateQueue(ctx, arg)
	if err != nil {
		slog.Error("failed to update queue", "params", arg, "error", err)
		return fmt.Errorf("failed to update queue: %w", err)
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueEdited,
		Payload:   queue,
	}

	slog.Info("queue updated successfully", "queueID", queue.ID)
	return nil
}

func (q *queueManager) ListDownloadsWithQueueName(ctx context.Context) ([]state.ListDownloadsWithQueueNameRow, error) {
	downloads, err := q.queries.ListDownloadsWithQueueName(ctx)
	if err != nil {
		slog.Error("failed to list downloads with queue name", "error", err)
		return nil, fmt.Errorf("failed to list downloads: %w", err)
	}

	slog.Info("listed downloads with queue name", "count", len(downloads))
	return downloads, nil
}

func (q *queueManager) PauseDownload(ctx context.Context, id int64) error {
	if err := q.setDownloadState(ctx, id, string(downloads.StatePaused)); err != nil {
		return err
	}

	currentDownload, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}

	q.mu.Lock()
	handler, ok := q.inProgressHandlers[id]
	if ok {
		if err := handler.Pause(); err != nil {
			q.mu.Unlock()
			slog.Error("failed to pause download handler", "downloadID", id, "error", err)
			return fmt.Errorf("failed to pause download handler: %w", err)
		}
		delete(q.inProgressHandlers, id)
	}
	q.mu.Unlock()

	if err := q.startNextDownloadIfPossible(ctx, currentDownload.QueueID); err != nil {
		return err
	}

	slog.Info("download paused successfully", "downloadID", id)
	return nil
}

func (q *queueManager) ResumeDownload(ctx context.Context, id int64) error {
	downloadConfig, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download configuration", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download configuration: %w", err)
	}

	downloadChunks, err := q.queries.GetDownloadChunksByDownloadID(ctx, id)
	if err != nil {
		slog.Error("failed to get download chunks", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download chunks: %w", err)
	}

	q.mu.Lock()
	limiter, ok := q.queueLimiters[downloadConfig.QueueID]
	if !ok {
		queue, err := q.queries.GetQueue(ctx, downloadConfig.QueueID)
		if err != nil {
			q.mu.Unlock()
			slog.Error("failed to get queue details", "queueID", downloadConfig.QueueID, "error", err)
			return fmt.Errorf("failed to get queue details: %w", err)
		}
		if queue.MaxBandwidth.Valid {
			limiter = bandwidthlimit.NewLimiter(&queue.MaxBandwidth.Int64)
		} else {
			limiter = bandwidthlimit.NewLimiter(nil)
		}
		q.queueLimiters[downloadConfig.QueueID] = limiter
	}
	q.mu.Unlock()

	handler, err := downloads.NewDownloadHandler(downloadConfig, downloadChunks, limiter)
	if err != nil {
		return err
	}

	if err := q.setDownloadState(ctx, id, string(downloads.StateInProgress)); err != nil {
		return err
	}

	q.mu.RLock()
	q.inProgressHandlers[id] = handler
	q.mu.RUnlock()

	if err := handler.Start(); err != nil {
		slog.Error("failed to start download handler", "downloadID", id, "error", err)
		return fmt.Errorf("failed to start download handler: %w", err)
	}

	slog.Info("download resumed successfully", "downloadID", id)
	return nil
}

func (q *queueManager) RetryDownload(ctx context.Context, id int64) error {
	_, err := q.queries.SetDownloadRetry(ctx, state.SetDownloadRetryParams{Retries: 0, ID: id})
	if err != nil {
		slog.Error("failed to set download retry count", "downloadID", id, "error", err)
		return fmt.Errorf("failed to set download retry count: %w", err)
	}

	slog.Info("retrying download", "downloadID", id)
	return q.ResumeDownload(ctx, id)
}

func (q *queueManager) ListQueue(ctx context.Context) ([]state.Queue, error) {
	queues, err := q.queries.ListQueues(ctx)
	if err != nil {
		slog.Error("failed to list queues", "error", err)
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	slog.Info("listed queues", "count", len(queues))
	return queues, nil
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
	// List queues
	queues, err := q.queries.ListQueues(ctx)
	if err != nil {
		slog.Error("failed to list queues during initialization", "error", err)
		return fmt.Errorf("failed to list queues during initialization: %w", err)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Initialize limiters
	for _, queue := range queues {
		if queue.MaxBandwidth.Valid {
			q.queueLimiters[queue.ID] = bandwidthlimit.NewLimiter(&queue.MaxBandwidth.Int64)
		} else {
			q.queueLimiters[queue.ID] = bandwidthlimit.NewLimiter(nil)
		}
	}

	// Initialize IN_PROGRESS downloads
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
