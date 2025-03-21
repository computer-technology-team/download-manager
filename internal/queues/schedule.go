package queues

import (
	"context"
	"log/slog"
	"time"
)

func (q *queueManager) scheduler(ctx context.Context) error {
	for {
		// Fetch all queues
		queues, err := q.ListQueue(ctx)
		if err != nil {
			slog.Error("failed to fetch queues", "error", err)
			time.Sleep(1 * time.Minute) // Wait before retrying
			continue
		}

		// Get the current time
		now := time.Now()

		// Iterate through each queue
		for _, queue := range queues {
			if queue.ScheduleMode {
				// Extract start and end times directly from the queue
				startTime := queue.StartDownload
				endTime := queue.EndDownload

				// Normalize the times to today's date
				start_time := time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour, startTime.Minute, 0, 0, now.Location())
				end_time := time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour, endTime.Minute, 0, 0, now.Location())

				// Handle cross-midnight time windows
				if end_time.Before(start_time) {
					// If the end time is before the start time, it means the time window spans across midnight
					// Adjust the end time to the next day
					end_time = end_time.Add(24 * time.Hour)
				}

				// Check if the current time is within the queue's active time
				if now.After(start_time) && now.Before(end_time) {
					for i := 0; i < int(queue.MaxConcurrent); i++ {
						q.startNextDownloadIfPossible(ctx, queue.ID)
					}
				} else {
					q.queries.UpdateInProgressToPendingInQueue(ctx, queue.ID)

					// Remove in-progress handlers for this queue
					q.mu.Lock()
					for downloadID, handler := range q.inProgressHandlers {
						download, err := q.queries.GetDownload(ctx, downloadID)
						if err != nil {
							slog.Error("failed to get download details", "downloadID", downloadID, "error", err)
							continue
						}

						if download.QueueID == queue.ID {
							// Stop the handler if it's running
							if err := handler.Pause(); err != nil {
								slog.Error("failed to pause download handler", "downloadID", downloadID, "error", err)
							}

							// Remove the handler from the map
							delete(q.inProgressHandlers, downloadID)
							slog.Debug("removed in-progress handler", "downloadID", downloadID)
						}
					}
					q.mu.Unlock()
				}
			}
		}

		// Sleep for 1 minute before checking again
		time.Sleep(1 * time.Minute)
	}
}
