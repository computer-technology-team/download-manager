// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: queues.sql

package state

import (
	"context"
	"database/sql"
)

const createQueue = `-- name: CreateQueue :one
INSERT INTO queues (name, directory, max_bandwidth, start_download, end_download, retry_limit)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id, name, directory, max_bandwidth, start_download, end_download, retry_limit
`

type CreateQueueParams struct {
	Name          string
	Directory     string
	MaxBandwidth  sql.NullInt64
	StartDownload string
	EndDownload   string
	RetryLimit    int64
}

func (q *Queries) CreateQueue(ctx context.Context, arg CreateQueueParams) (Queue, error) {
	row := q.db.QueryRowContext(ctx, createQueue,
		arg.Name,
		arg.Directory,
		arg.MaxBandwidth,
		arg.StartDownload,
		arg.EndDownload,
		arg.RetryLimit,
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
SELECT id, name, directory, max_bandwidth, start_download, end_download, retry_limit FROM queues
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
	)
	return i, err
}

const listQueues = `-- name: ListQueues :many
SELECT id, name, directory, max_bandwidth, start_download, end_download, retry_limit FROM queues
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

const updateQueue = `-- name: UpdateQueue :one
UPDATE queues
SET name = ?, max_bandwidth = ?, start_download = ?, end_download = ?, retry_limit = ?
WHERE id = ?
RETURNING id, name, directory, max_bandwidth, start_download, end_download, retry_limit
`

type UpdateQueueParams struct {
	Name          string
	MaxBandwidth  sql.NullInt64
	StartDownload string
	EndDownload   string
	RetryLimit    int64
	ID            int64
}

func (q *Queries) UpdateQueue(ctx context.Context, arg UpdateQueueParams) (Queue, error) {
	row := q.db.QueryRowContext(ctx, updateQueue,
		arg.Name,
		arg.MaxBandwidth,
		arg.StartDownload,
		arg.EndDownload,
		arg.RetryLimit,
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
	)
	return i, err
}
