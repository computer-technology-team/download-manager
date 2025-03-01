package downloads

type DownloadState string

const (
	StateInitialized DownloadState = "initialized"
	StateDownloading DownloadState = "downloading"
	StatePaused      DownloadState = "paused"
	StateCompleted   DownloadState = "completed"
	StateCanceled    DownloadState = "canceled"
	StateFailed      DownloadState = "failed"
)

type Downloader interface {
	Start() error

	Pause() error

	Resume() error

	Cancel() error

	Retry() error

	Status() DownloadStatus
}

type DownloadStatus struct {
	ProgressRate float32 // percent
	Speed        float32 // bytes per second
	State        DownloadState
}

type Download struct {
	URL            string
	SavePath       string
	BandwidthLimit int64 // bytes per second (-1 means unlimited)
}

func NewDownloader(cfg Download) Downloader {
	return &defaultDownloader{
		url:            cfg.URL,
		savePath:       cfg.SavePath,
		bandwidthLimit: cfg.BandwidthLimit,
		state:          StateInitialized,
	}
}
