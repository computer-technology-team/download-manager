package queues

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"

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
}

// CreateDownload implements QueueManager.
func (q queueManager) CreateDownload(ctx context.Context, downloadURL, fileName string, queueID int64) error {
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return err
	}

	if fileName == "" {
		lastPathSegment := path.Base(parsedURL.Path)

		if lastPathSegment == "" || lastPathSegment == "." || lastPathSegment == "/" {
			return ErrEmptyFileName
		}

		fileName = lastPathSegment
	}

	queue, err := q.queries.GetQueue(ctx, queueID)
	if err != nil {
		return fmt.Errorf("failed to get queue from db: %w", err)
	}

	createDownloadParams := state.CreateDownloadParams{
		QueueID:  queueID,
		Url:      downloadURL,
		SavePath: path.Join(queue.Directory, fileName),
		State:    "",
		Retries:  0,
	}

	download, err := q.queries.CreateDownload(ctx, createDownloadParams)
	if err != nil {
		return err
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.DownloadCreated,
		Payload:   download,
	}

	return nil
}

// CreateQueue implements QueueManager.
func (q queueManager) CreateQueue(ctx context.Context, createQueueParams state.CreateQueueParams) error {
	queue, err := q.queries.CreateQueue(ctx, createQueueParams)
	if err != nil {
		return err
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueCreated,
		Payload:   queue,
	}

	return nil
}

func (q queueManager) DeleteDownload(ctx context.Context, id int64) error {
	err := q.inProgressHandlers[id].Pause()
	if err != nil {
		return err
	}

	current_download, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		return err
	}

	queue_id := current_download.QueueID
	nextDownload, err := q.queries.GetPausedDownloadByQueueID(ctx, queue_id)
	if err == nil {
		_, err = q.queries.SetDownloadState(ctx, state.SetDownloadStateParams{State: "IN_PROGRESS", ID: nextDownload.ID})
		return err
	}

	return errors.Join(err, q.queries.DeleteDownload(ctx, id))
}

// DeleteQueue implements QueueManager.
func (q queueManager) DeleteQueue(ctx context.Context, id int64) error {
	err := q.queries.DeleteQueue(ctx, id)
	if err != nil {
		return err
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueDeleted,
		Payload:   id,
	}

	return nil
}

func (q queueManager) EditQueue(ctx context.Context, arg state.UpdateQueueParams) error {
	queue, err := q.queries.UpdateQueue(ctx, arg)
	if err != nil {
		return err
	}

	events.GetUIEventChannel() <- events.Event{
		EventType: events.QueueEdited,
		Payload:   queue,
	}

	return nil
}

func (q queueManager) ListDownloadsWithQueueName(ctx context.Context) ([]state.ListDownloadsWithQueueNameRow, error) {
	return q.queries.ListDownloadsWithQueueName(ctx)
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
