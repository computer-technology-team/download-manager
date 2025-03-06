package downloads

//"errors"

type defaultDownloader struct {
	url            string
	savePath       string
	bandwidthLimit int64
	state          DownloadState
}

func (d *defaultDownloader) Start() error {
	return nil
}

func (d *defaultDownloader) Pause() error {
	return nil
}

func (d *defaultDownloader) Resume() error {
	return nil
}

func (d *defaultDownloader) Cancel() error {
	return nil
}

func (d *defaultDownloader) Retry() error {
	return nil
}

func (d *defaultDownloader) Status() DownloadStatus {
	return DownloadStatus{
		ProgressRate: 0.0,
		Speed:        0.0,
		State:        d.state,
	}
}
