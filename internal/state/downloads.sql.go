

package state

import (
	"context"
)

const createDownload = `-- name: CreateDownload :one
INSERT INTO downloads (queue_id, url, save_path, state, retries)
VALUES (?, ?, ?, ?, ?)
RETURNING id, queue_id, url, save_path, state, retries
`

type CreateDownloadParams struct {
	QueueID  int64
	Url      string
	SavePath string
	State    string
	Retries  int64
}

func (q *Queries) CreateDownload(ctx context.Context, arg CreateDownloadParams) (Download, error) {
	row := q.db.QueryRowContext(ctx, createDownload,
		arg.QueueID,
		arg.Url,
		arg.SavePath,
		arg.State,
		arg.Retries,
	)
	var i Download
	err := row.Scan(
		&i.ID,
		&i.QueueID,
		&i.Url,
		&i.SavePath,
		&i.State,
		&i.Retries,
	)
	return i, err
}

const deleteDownload = `-- name: DeleteDownload :exec
DELETE FROM downloads
WHERE id = ?
`

func (q *Queries) DeleteDownload(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deleteDownload, id)
	return err
}

const deleteDownloadChunk = `-- name: DeleteDownloadChunk :exec
DELETE FROM download_chunks
WHERE id = ?
`

func (q *Queries) DeleteDownloadChunk(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteDownloadChunk, id)
	return err
}

const getDownload = `-- name: GetDownload :one
SELECT id, queue_id, url, save_path, state, retries FROM downloads
WHERE id = ?
`

func (q *Queries) GetDownload(ctx context.Context, id int64) (Download, error) {
	row := q.db.QueryRowContext(ctx, getDownload, id)
	var i Download
	err := row.Scan(
		&i.ID,
		&i.QueueID,
		&i.Url,
		&i.SavePath,
		&i.State,
		&i.Retries,
	)
	return i, err
}

const getDownloadChunk = `-- name: GetDownloadChunk :one
SELECT id, range_start, range_end, current_pointer, download_id, single_part FROM download_chunks
WHERE id = ?
`

func (q *Queries) GetDownloadChunk(ctx context.Context, id string) (DownloadChunk, error) {
	row := q.db.QueryRowContext(ctx, getDownloadChunk, id)
	var i DownloadChunk
	err := row.Scan(
		&i.ID,
		&i.RangeStart,
		&i.RangeEnd,
		&i.CurrentPointer,
		&i.DownloadID,
		&i.SinglePart,
	)
	return i, err
}

const getDownloadChunksByDownloadID = `-- name: GetDownloadChunksByDownloadID :many
SELECT id, range_start, range_end, current_pointer, download_id, single_part FROM download_chunks
WHERE download_id = ?
`

func (q *Queries) GetDownloadChunksByDownloadID(ctx context.Context, downloadID int64) ([]DownloadChunk, error) {
	rows, err := q.db.QueryContext(ctx, getDownloadChunksByDownloadID, downloadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DownloadChunk
	for rows.Next() {
		var i DownloadChunk
		if err := rows.Scan(
			&i.ID,
			&i.RangeStart,
			&i.RangeEnd,
			&i.CurrentPointer,
			&i.DownloadID,
			&i.SinglePart,
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

const getDownloadsByStatus = `-- name: GetDownloadsByStatus :many
SELECT id, queue_id, url, save_path, state, retries 
FROM downloads
WHERE state = ?
`

func (q *Queries) GetDownloadsByStatus(ctx context.Context, state string) ([]Download, error) {
	rows, err := q.db.QueryContext(ctx, getDownloadsByStatus, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Download
	for rows.Next() {
		var i Download
		if err := rows.Scan(
			&i.ID,
			&i.QueueID,
			&i.Url,
			&i.SavePath,
			&i.State,
			&i.Retries,
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

const getPendingDownloadByQueueID = `-- name: GetPendingDownloadByQueueID :one
SELECT id, queue_id, url, save_path, state, retries FROM downloads
WHERE queue_id = ? AND state = 'PENDING'
LIMIT 1
`

func (q *Queries) GetPendingDownloadByQueueID(ctx context.Context, queueID int64) (Download, error) {
	row := q.db.QueryRowContext(ctx, getPendingDownloadByQueueID, queueID)
	var i Download
	err := row.Scan(
		&i.ID,
		&i.QueueID,
		&i.Url,
		&i.SavePath,
		&i.State,
		&i.Retries,
	)
	return i, err
}

const listDownloadChunks = `-- name: ListDownloadChunks :many
SELECT id, range_start, range_end, current_pointer, download_id, single_part FROM download_chunks
`

func (q *Queries) ListDownloadChunks(ctx context.Context) ([]DownloadChunk, error) {
	rows, err := q.db.QueryContext(ctx, listDownloadChunks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DownloadChunk
	for rows.Next() {
		var i DownloadChunk
		if err := rows.Scan(
			&i.ID,
			&i.RangeStart,
			&i.RangeEnd,
			&i.CurrentPointer,
			&i.DownloadID,
			&i.SinglePart,
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

const listDownloads = `-- name: ListDownloads :many
SELECT id, queue_id, url, save_path, state, retries 
FROM downloads
`

func (q *Queries) ListDownloads(ctx context.Context) ([]Download, error) {
	rows, err := q.db.QueryContext(ctx, listDownloads)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Download
	for rows.Next() {
		var i Download
		if err := rows.Scan(
			&i.ID,
			&i.QueueID,
			&i.Url,
			&i.SavePath,
			&i.State,
			&i.Retries,
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

const listDownloadsWithQueueName = `-- name: ListDownloadsWithQueueName :many
SELECT downloads.id, downloads.queue_id, downloads.url, downloads.save_path, downloads.state, downloads.retries, queues.name as queue_name
FROM downloads JOIN queues on downloads.queue_id = queues.id
`

type ListDownloadsWithQueueNameRow struct {
	ID        int64
	QueueID   int64
	Url       string
	SavePath  string
	State     string
	Retries   int64
	QueueName string
}

func (q *Queries) ListDownloadsWithQueueName(ctx context.Context) ([]ListDownloadsWithQueueNameRow, error) {
	rows, err := q.db.QueryContext(ctx, listDownloadsWithQueueName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListDownloadsWithQueueNameRow
	for rows.Next() {
		var i ListDownloadsWithQueueNameRow
		if err := rows.Scan(
			&i.ID,
			&i.QueueID,
			&i.Url,
			&i.SavePath,
			&i.State,
			&i.Retries,
			&i.QueueName,
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

const setDownloadRetry = `-- name: SetDownloadRetry :one
UPDATE downloads
SET retries = ?
WHERE id = ?
RETURNING id, queue_id, url, save_path, state, retries
`

type SetDownloadRetryParams struct {
	Retries int64
	ID      int64
}

func (q *Queries) SetDownloadRetry(ctx context.Context, arg SetDownloadRetryParams) (Download, error) {
	row := q.db.QueryRowContext(ctx, setDownloadRetry, arg.Retries, arg.ID)
	var i Download
	err := row.Scan(
		&i.ID,
		&i.QueueID,
		&i.Url,
		&i.SavePath,
		&i.State,
		&i.Retries,
	)
	return i, err
}

const setDownloadState = `-- name: SetDownloadState :one
UPDATE downloads
SET state = ?
WHERE id = ?
RETURNING id, queue_id, url, save_path, state, retries
`

type SetDownloadStateParams struct {
	State string
	ID    int64
}

func (q *Queries) SetDownloadState(ctx context.Context, arg SetDownloadStateParams) (Download, error) {
	row := q.db.QueryRowContext(ctx, setDownloadState, arg.State, arg.ID)
	var i Download
	err := row.Scan(
		&i.ID,
		&i.QueueID,
		&i.Url,
		&i.SavePath,
		&i.State,
		&i.Retries,
	)
	return i, err
}

const upsertDownloadChunk = `-- name: UpsertDownloadChunk :one
INSERT INTO download_chunks (id, range_start, range_end, current_pointer, download_id, single_part)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE
SET current_pointer = EXCLUDED.current_pointer
RETURNING id, range_start, range_end, current_pointer, download_id, single_part
`

type UpsertDownloadChunkParams struct {
	ID             string
	RangeStart     int64
	RangeEnd       int64
	CurrentPointer int64
	DownloadID     int64
	SinglePart     bool
}

func (q *Queries) UpsertDownloadChunk(ctx context.Context, arg UpsertDownloadChunkParams) (DownloadChunk, error) {
	row := q.db.QueryRowContext(ctx, upsertDownloadChunk,
		arg.ID,
		arg.RangeStart,
		arg.RangeEnd,
		arg.CurrentPointer,
		arg.DownloadID,
		arg.SinglePart,
	)
	var i DownloadChunk
	err := row.Scan(
		&i.ID,
		&i.RangeStart,
		&i.RangeEnd,
		&i.CurrentPointer,
		&i.DownloadID,
		&i.SinglePart,
	)
	return i, err
}
