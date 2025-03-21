

package state

import (
	"context"
	"database/sql"
)

const countInProgressDownloadsInQueue = `-- name: CountInProgressDownloadsInQueue :one
SELECT COUNT(*) 
FROM downloads
WHERE queue_id = ? AND status = 'IN_PROGRESS'
`

func (q *Queries) CountInProgressDownloadsInQueue(ctx context.Context, queueID int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, countInProgressDownloadsInQueue, queueID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createQueue = `-- name: CreateQueue :one
INSERT INTO queues (name, directory, max_bandwidth, start_download, end_download, retry_limit, max_concurrent, schedule_mode)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, directory, max_bandwidth, start_download, end_download, retry_limit, schedule_mode, max_concurrent
`

type CreateQueueParams struct {
	Name          string
	Directory     string
	MaxBandwidth  sql.NullInt64
	StartDownload TimeValue
	EndDownload   TimeValue
	RetryLimit    int64
	MaxConcurrent int64
	ScheduleMode  bool
}

func (q *Queries) CreateQueue(ctx context.Context, arg CreateQueueParams) (Queue, error) {
	row := q.db.QueryRowContext(ctx, createQueue,
		arg.Name,
		arg.Directory,
		arg.MaxBandwidth,
		arg.StartDownload,
		arg.EndDownload,
		arg.RetryLimit,
		arg.MaxConcurrent,
		arg.ScheduleMode,
	)
	var i Queue
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Directory,
		&i.MaxBandwidth,
		&i.StartDownload,
		&i.EndDownload,
		&i.RetryLimit,
		&i.ScheduleMode,
		&i.MaxConcurrent,
	)
	return i, err
}

const deleteQueue = `-- name: DeleteQueue :exec
DELETE FROM queues
WHERE id = ?
`

func (q *Queries) DeleteQueue(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteQueue, id)
	return err
}

const getQueue = `-- name: GetQueue :one
SELECT id, name, directory, max_bandwidth, start_download, end_download, retry_limit, schedule_mode, max_concurrent FROM queues
WHERE id = ?
`

func (q *Queries) GetQueue(ctx context.Context, id int64) (Queue, error) {
	row := q.db.QueryRowContext(ctx, getQueue, id)
	var i Queue
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Directory,
		&i.MaxBandwidth,
		&i.StartDownload,
		&i.EndDownload,
		&i.RetryLimit,
		&i.ScheduleMode,
		&i.MaxConcurrent,
	)
	return i, err
}

const listQueues = `-- name: ListQueues :many
SELECT id, name, directory, max_bandwidth, start_download, end_download, retry_limit, schedule_mode, max_concurrent FROM queues
`

func (q *Queries) ListQueues(ctx context.Context) ([]Queue, error) {
	rows, err := q.db.QueryContext(ctx, listQueues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Queue
	for rows.Next() {
		var i Queue
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Directory,
			&i.MaxBandwidth,
			&i.StartDownload,
			&i.EndDownload,
			&i.RetryLimit,
			&i.ScheduleMode,
			&i.MaxConcurrent,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInProgressToPendingInQueue = `-- name: UpdateInProgressToPendingInQueue :exec
UPDATE downloads
SET status = 'PENDING'
WHERE status = 'IN_PROGRESS' AND queue_id = ?
`

func (q *Queries) UpdateInProgressToPendingInQueue(ctx context.Context, queueID int64) error {
	_, err := q.db.ExecContext(ctx, updateInProgressToPendingInQueue, queueID)
	return err
}

const updateQueue = `-- name: UpdateQueue :one
UPDATE queues
SET name = ?, max_bandwidth = ?, start_download = ?, end_download = ?,
retry_limit = ?, max_concurrent = ?, schedule_mode = ?, directory = ?
WHERE id = ?
RETURNING id, name, directory, max_bandwidth, start_download, end_download, retry_limit, schedule_mode, max_concurrent
`

type UpdateQueueParams struct {
	Name          string
	MaxBandwidth  sql.NullInt64
	StartDownload TimeValue
	EndDownload   TimeValue
	RetryLimit    int64
	MaxConcurrent int64
	ScheduleMode  bool
	Directory     string
	ID            int64
}

func (q *Queries) UpdateQueue(ctx context.Context, arg UpdateQueueParams) (Queue, error) {
	row := q.db.QueryRowContext(ctx, updateQueue,
		arg.Name,
		arg.MaxBandwidth,
		arg.StartDownload,
		arg.EndDownload,
		arg.RetryLimit,
		arg.MaxConcurrent,
		arg.ScheduleMode,
		arg.Directory,
		arg.ID,
	)
	var i Queue
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Directory,
		&i.MaxBandwidth,
		&i.StartDownload,
		&i.EndDownload,
		&i.RetryLimit,
		&i.ScheduleMode,
		&i.MaxConcurrent,
	)
	return i, err
}
