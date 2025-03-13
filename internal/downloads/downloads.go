package downloads

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type defaultDownloader struct {
	url            string
	savePath       string
	bandwidthLimit int64
	state          DownloadState
	queries        *state.Queries
	ticker         Ticker
	hasStarted     bool
}

func (d *defaultDownloader) GetTicker() *Ticker {
	return &d.ticker
}

func (d *defaultDownloader) Start() error {
	// err := d.queries.CreateDownload(context.Background(), state.CreateDownloadParams{
	// 	Url:                   d.url,
	// 	SavePath:              d.savePath,
	// 	BandwidthLimitBytesPS: float64(d.bandwidthLimit),
	// })
	// if err != nil {
	// 	// TODO return err
	// }
	d.hasStarted = true
	req, err := http.NewRequest("HEAD", d.url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error in requesting header: ", err)
		// TODO log.Fatal(err)
	}
	for k, vs := range resp.Header { // TODO test if
		fmt.Printf("%s: %d, %+v\n", k, len(vs), vs)
	}

	writer := NewSynchronizedFileWriter(d.savePath)
	segmentsList := d.getChunkSegments(resp.Header)

	chunkhandlersList := make([]*DownloadChunkHandler, 0)

	for _, segment := range segmentsList {
		l, r := segment[0], segment[1]
		fmt.Println(l, r, "range")
		handler := NewDownloadChunkHandler(DownloaderConfig{
			URL:                   d.url,
			SavePath:              d.savePath,
			BandwidthLimitBytesPS: d.bandwidthLimit,
		}, l, r, &d.ticker, writer)
		chunkhandlersList = append(chunkhandlersList, &handler)
	}

	for _, handler := range chunkhandlersList {
		handler.Start()
	}

	d.ticker.Start()

	return nil
}

func (d *defaultDownloader) getChunkSegments(header http.Header) [][]int64 {
	//TODO this is a prototype

	size, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil {
		fmt.Println("no content length header", err) // TODO
	}

	if header.Get("Accept-Ranges") != "bytes" {
		fmt.Errorf("server does not accept range requests") // TODO need an if on test mode
		return [][]int64{{0, int64(size)}}
	}

	chunkSize := int(size / 10)

	segmentsList := make([][]int64, 0)

	for i := 0; chunkSize*i < size; i++ {
		segmentsList = append(segmentsList, []int64{int64(i * chunkSize), int64(min((i+1)*chunkSize, size))})
	}

	return segmentsList

}

func getFileSize(url string) (int64, error) {
	panic("unimplemented")
}

func (d *defaultDownloader) Pause() error {
	// err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
	// 	State: string(StatePaused),
	// 	ID:    d.url, // Use the URL or ID as the key
	// })
	// if err != nil {
	// 	return err
	// }
	d.ticker.Quite()
	return nil
}

func (d *defaultDownloader) Resume() error {
	// err := d.queries.UpdateDownloadState(context.Background(), state.UpdateDownloadStateParams{
	// 	State: string(StateDownloading),
	// 	ID:    d.url, // Use the URL or ID as the key
	// })
	// if err != nil {
	// 	return err
	// }
	if d.hasStarted{
		d.ticker.Start()
	} else{
		d.Start()
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
