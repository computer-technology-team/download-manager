package queues

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"golang.org/x/exp/slog"
)

type QueueManager interface {
	PauseDownload(ctx context.Context, id int64) error
	ResumeDownload(ctx context.Context, id int64) error
	RetryDownload(ctx context.Context, id int64) error

	CreateDownload(ctx context.Context, arg state.CreateDownloadParams) error
	ListDownloads(ctx context.Context) ([]state.Download, error)
	DeleteDownload(ctx context.Context, id int64) error

	CreateQueue(ctx context.Context, arg state.CreateQueueParams) error
	DeleteQueue(ctx context.Context, id int64) error
	ListQueue(ctx context.Context) ([]state.Queue, error)
	EditQueue(ctx context.Context, id int64, arg state.UpdateQueueParams) error
}

type queueManager struct {
	queries            *state.Queries
	inProgressHandlers map[int64]downloads.DownloadHandler
	queueTickers       map[int64]bandwidthlimit.Ticker
}

func (q queueManager) CreateDownload(ctx context.Context, arg state.CreateDownloadParams) error {
	_, err := q.queries.CreateDownload(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (q queueManager) CreateQueue(ctx context.Context, arg state.CreateQueueParams) error {
	queue, err := q.queries.CreateQueue(ctx, arg)
	if err != nil {
		return err
	}
	q.queueTickers[queue.ID] = bandwidthlimit.NewTicker()
	return nil
}

func (q queueManager) DeleteDownload(ctx context.Context, id int64) error {
	q.inProgressHandlers[id].Pause()
	current_download, err := q.queries.GetDownload(ctx, id)
	if err != nil {
		return err
	}
	queue_id := current_download.QueueID
	nextDownload, err := q.queries.GetPausedDownloadByQueueID(ctx, queue_id)
	if err == nil {
		download_config, err := q.queries.GetDownload(ctx, nextDownload.ID)
		if err != nil {
			return err
		}
		next_download_chunks, _ := q.queries.GetDownloadChunksByDownloadID(ctx, nextDownload.ID)
		q.inProgressHandlers[nextDownload.ID] = downloads.NewDownloadHandler(download_config, next_download_chunks, bandwidthlimit.NewTicker())
		q.queries.SetDownloadState(ctx, state.SetDownloadStateParams{State: "IN_PROGRESS", ID: nextDownload.ID})
	}
	q.queries.DeleteDownload(ctx, id)
	return nil
}

func (q queueManager) DeleteQueue(ctx context.Context, id int64) error {
	if err := q.queries.DeleteQueue(ctx, id); err != nil {
		return err
	}
	if ticker, exists := q.queueTickers[id]; exists {
		ticker.Stop()
		delete(q.queueTickers, id)
	}
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

func (q queueManager) ListQueue(ctx context.Context) ([]state.Queue, error) {
	return q.queries.ListQueues(ctx)
}

func New(db *sql.DB) QueueManager {
	return queueManager{queries: state.New(db)}
}

func (q queueManager) scheduler() {
	for {
		// Fetch all queues
		queues, err := q.ListQueue(context.Background())
		if err != nil {
			log.Printf("Error fetching queues: %v", err)
			continue
		}

		// Get the current time
		now := time.Now()

		// Iterate through each queue
		for _, queue := range queues {
			// Parse the start and end times
			startTime, err := time.Parse("15:04", queue.StartDownload)
			if err != nil {
				log.Printf("Error parsing start time for queue %d: %v", queue.ID, err)
				continue
			}
			endTime, err := time.Parse("15:04", queue.EndDownload)
			if err != nil {
				log.Printf("Error parsing end time for queue %d: %v", queue.ID, err)
				continue
			}

			// Normalize the times to today's date
			startTime = time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
			endTime = time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 0, 0, now.Location())

			// Check if the current time is within the queue's active time
			if now.After(startTime) && now.Before(endTime) {
				// Start the ticker if it's not already running
				if _, exists := q.queueTickers[queue.ID]; !exists {
					ticker := bandwidthlimit.NewTicker()
					ticker.Start()
					q.queueTickers[queue.ID] = ticker
					slog.Debug("Started ticker for queue %d", queue.ID)
				}
			} else {
				// Stop the ticker if it's running
				if ticker, exists := q.queueTickers[queue.ID]; exists {
					ticker.Stop()
					delete(q.queueTickers, queue.ID)
					log.Printf("Stopped ticker for queue %d", queue.ID)
				}
			}
		}

		// Sleep for 1 minute before checking again
		time.Sleep(1 * time.Minute)
	}
}
