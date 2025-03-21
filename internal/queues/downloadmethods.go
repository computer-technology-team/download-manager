package queues

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

func (q *queueManager) PauseDownload(ctx context.Context, id int64) error {
	if err := q.setDownloadState(ctx, id, string(downloads.StatePaused)); err != nil {
		return err
	}

	currentDownload, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}

	if currentDownload.State != string(downloads.StateInProgress) {
		slog.Error("invalid state", "current_download_state", currentDownload.State)
		return errors.New("can not pause download that is not in progress")
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

	createDownloadParams.State = string(downloads.StateInProgress)

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

func (q *queueManager) DeleteDownload(ctx context.Context, id int64) error {
	q.mu.RLock()
	handler, ok := q.inProgressHandlers[id]
	q.mu.RUnlock()

	if ok {
		if err := handler.Pause(); err != nil {
			slog.Error("failed to pause download handler", "downloadID", id, "error", err)
			return fmt.Errorf("failed to pause download handler: %w", err)
		}
	}

	currentDownload, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}
	queueID := currentDownload.QueueID 

	if err := q.queries.DeleteDownload(ctx, id); err != nil {
		slog.Error("failed to delete download", "downloadID", id, "error", err)
		return fmt.Errorf("failed to delete download: %w", err)
	}

	if ok {
		q.mu.Lock()
		delete(q.inProgressHandlers, id)
		q.mu.Unlock()

		if err := q.startNextDownloadIfPossible(ctx, queueID); err != nil {
			return err
		}
	}

	slog.Info("download deleted successfully", "downloadID", id)

	return nil
}
