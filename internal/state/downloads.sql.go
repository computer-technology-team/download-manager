// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: downloads.sql

package state

import (
	"context"
)

const getAllDownloads = `-- name: GetAllDownloads :many
SELECT id
FROM
    downloads
`

func (q *Queries) GetAllDownloads(ctx context.Context) ([]interface{}, error) {
	rows, err := q.db.QueryContext(ctx, getAllDownloads)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []interface{}
	for rows.Next() {
		var id interface{}
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
