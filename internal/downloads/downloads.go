package downloads

import (
	"context"

	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type defaultDownloader struct {
	url            string
	savePath       string
	bandwidthLimit int64
	state          DownloadState
	queries        *state.Queries
	ticker			Ticker
}

func (d *defaultDownloader) Start() error {
	err := d.queries.CreateDownload(context.Background(), state.CreateDownloadParams{
		Url:                   d.url,
		SavePath:              d.savePath,
		BandwidthLimitBytesPS: float64(d.bandwidthLimit),
	})
	if err != nil {
		return err
	}


	

	return nil
}

func (d *defaultDownloader) Pause() error {
	err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
		State: string(StatePaused),
		ID:    d.url, // Use the URL or ID as the key
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *defaultDownloader) Resume() error {
	err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
		State: string(StateDownloading),
		ID:    d.url, // Use the URL or ID as the key
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *defaultDownloader) Cancel() error {
	err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
		State: string(StateCanceled),
		ID:    d.url, // Use the URL or ID as the key
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *defaultDownloader) Retry() error {
	err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
		State: string(StateDownloading),
		ID:    d.url, // Use the URL or ID as the key
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *defaultDownloader) Status() DownloadStatus {
	status, err := d.queries.GetDownloadStatus(context.Background(), d.url) // Use the URL or ID as the key
	if err != nil {
		return DownloadStatus{State: StateFailed}
	}

	return DownloadStatus{
		ProgressRate: float32(status.ProgressPersent),
		Speed:        float32(status.SpeedBytesPS),
		State:        DownloadState(status.State),
	}
}
