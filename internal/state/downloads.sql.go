// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: downloads.sql

package state

import (
	"context"
)

const getAllDownloads = `-- name: GetAllDownloads :many
SELECT id, queue_id, url, save_path, state, retries FROM downloads
`

func (q *Queries) GetAllDownloads(ctx context.Context) ([]Download, error) {
	rows, err := q.db.QueryContext(ctx, getAllDownloads)
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
