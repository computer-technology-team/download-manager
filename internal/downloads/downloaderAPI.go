package downloads

import (
	"database/sql"

	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type DownloadState string

const (
	StateInitialized DownloadState = "initialized"
	StateDownloading DownloadState = "downloading"
	StatePaused      DownloadState = "paused"
	StateCompleted   DownloadState = "completed"
	StateCanceled    DownloadState = "canceled"
	StateFailed      DownloadState = "failed"
)

type DownloadHandler interface {
	Start() error

	Pause() error

	Resume() error

	Cancel() error

	Retry() error

	Status() DownloadStatus
	
	GetTicker() *Ticker
}

type DownloadStatus struct {
	ProgressRate float32 // percent
	Speed        float32 // bytes per second
	State        DownloadState
}

type DownloaderConfig struct {
	URL                   string
	SavePath              string
	BandwidthLimitBytesPS int64 // (-1 for unlimited)
}

func NewDownloader(cfg DownloaderConfig, db *sql.DB) DownloadHandler {
	d := defaultDownloader{
		url:            cfg.URL,
		savePath:       cfg.SavePath,
		bandwidthLimit: cfg.BandwidthLimitBytesPS,
		state:          StateInitialized,
		queries:        state.New(db),
		ticker:         NewTicker(),
	}
	d.ticker.SetBandwidth(d.bandwidthLimit)
	return &d
}
