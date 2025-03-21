package queues

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

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

func (q *queueManager) startNextDownloadIfPossible(ctx context.Context, queueID int64) error {
	var activeDownloads int64 = 0

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

	queue, err := q.queries.GetQueue(ctx, queueID)
	if err != nil {
		slog.Error("failed to get queue details", "queueID", queueID, "error", err)
		return fmt.Errorf("failed to get queue details: %w", err)
	}

	if activeDownloads >= queue.MaxConcurrent {
		slog.Info("queue is full, cannot start next download", "queueID", queueID, "activeDownloads", activeDownloads, "maxConcurrent", queue.MaxConcurrent)
		return nil
	}

	nextDownload, err := q.queries.GetPendingDownloadByQueueID(ctx, queueID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		slog.Error("failed to get pending download by queue ID", "queueID", queueID, "error", err)
		return err
	}

	if err := q.ResumeDownload(ctx, nextDownload.ID); err != nil {
		slog.Error("failed to resume download", "downloadID", nextDownload.ID, "error", err)
		return fmt.Errorf("failed to resume download: %w", err)
	}

	slog.Info("started next download", "downloadID", nextDownload.ID)
	return nil
}

func (q *queueManager) startNextDownloadIfPossibleByDownloadID(ctx context.Context, downloadID int64) error {

	download, err := q.queries.GetDownload(ctx, downloadID)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", downloadID, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}

	queueID := download.QueueID

	return q.startNextDownloadIfPossible(ctx, queueID)
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

func (q *queueManager) UpsertChunks(ctx context.Context, status downloads.DownloadStatus) error {
	var errs []error

	for _, chunk := range status.DownloadChuncks {
		_, err := q.queries.UpsertDownloadChunk(ctx,
			state.UpsertDownloadChunkParams(chunk))
		if err != nil {
			slog.Error("Could not upsert download chunk", "chunkID", chunk.ID, "downloadID", chunk.DownloadID, "error", err)
			errs = append(errs, err)
		} else {
			slog.Debug("Download chunk upserted successfully", "chunkID", chunk.ID, "downloadID", chunk.DownloadID)
		}
	}

	return errors.Join(errs...)
}

func (q *queueManager) DownloadFailed(ctx context.Context, id int64) error {

	download, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		slog.Error("failed to get download details", "downloadID", id, "error", err)
		return fmt.Errorf("failed to get download details: %w", err)
	}

	queue, err := q.queries.GetQueue(ctx, download.QueueID)
	if err != nil {
		slog.Error("failed to get queue details", "queueID", download.QueueID, "error", err)
		return fmt.Errorf("failed to get queue details: %w", err)
	}

	if download.Retries < queue.RetryLimit {

		if _, err := q.queries.SetDownloadRetry(ctx, state.SetDownloadRetryParams{
			Retries: download.Retries + 1,
			ID:      id,
		}); err != nil {
			slog.Error("failed to set download retry count", "downloadID", id, "error", err)
			return fmt.Errorf("failed to set download retry count: %w", err)
		}

		if err := q.ResumeDownload(ctx, id); err != nil {
			slog.Error("failed to retry download", "downloadID", id, "error", err)
			return fmt.Errorf("failed to retry download: %w", err)
		}

		slog.Info("retrying download", "downloadID", id, "retryCount", download.Retries+1)
		return nil
	}

	if err := q.setDownloadState(ctx, id, string(downloads.StateFailed)); err != nil {
		slog.Error("failed to set download state to failed", "downloadID", id, "error", err)
		return fmt.Errorf("failed to set download state to failed: %w", err)
	}

	q.mu.Lock()
	delete(q.inProgressHandlers, id)
	q.mu.Unlock()

	slog.Info("download marked as failed", "downloadID", id)

	if err := q.startNextDownloadIfPossibleByDownloadID(ctx, id); err != nil {
		slog.Error("failed to start next download in queue", "downloadID", id, "error", err)
		return fmt.Errorf("failed to start next download in queue: %w", err)
	}

	return nil
}

func (q *queueManager) DownloadCompleted(ctx context.Context, id int64) error {

	if err := q.setDownloadState(ctx, id, string(downloads.StateCompleted)); err != nil {
		slog.Error("failed to set download state to completed", "downloadID", id, "error", err)
		return fmt.Errorf("failed to set download state to completed: %w", err)
	}

	slog.Info("download marked as completed", "downloadID", id)

	if err := q.startNextDownloadIfPossibleByDownloadID(ctx, id); err != nil {
		slog.Error("failed to start next download in queue", "downloadID", id, "error", err)
		return fmt.Errorf("failed to start next download in queue: %w", err)
	}

	return nil
}
