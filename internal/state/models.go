

package state

import (
	"database/sql"
)

type Download struct {
	ID       int64
	QueueID  int64
	Url      string
	SavePath string
	State    string
	Retries  int64
}

type DownloadChunk struct {
	ID             string
	RangeStart     int64
	RangeEnd       int64
	CurrentPointer int64
	DownloadID     int64
	SinglePart     bool
}

type Queue struct {
	ID            int64
	Name          string
	Directory     string
	MaxBandwidth  sql.NullInt64
	StartDownload TimeValue
	EndDownload   TimeValue
	RetryLimit    int64
	ScheduleMode  bool
	MaxConcurrent int64
}
