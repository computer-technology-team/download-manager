package downloads

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	// "github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/google/uuid"
)

const progressUpdatePeriod int = 1000 // milliseconds
const movingAverageScale float64 = .1 // new average = old * (1 - alpha) + current * alpha
const numberOfChuncks = 10

type defaultDownloader struct {
	id            int64
	queueID       int64
	url           string
	savePath      string
	state         DownloadState
	ticker        bandwidthlimit.Ticker
	chunkHandlers []*DownloadChunkHandler
	progress      int64
	progressRate  float64
	size          int64
	pausedChan    *chan int
	isPaused      bool
}

func (d *defaultDownloader) GetTicker() *bandwidthlimit.Ticker {
	return &d.ticker
}

func (d *defaultDownloader) keepTrackOfProgress() {
	for {
		time.Sleep(time.Duration(progressUpdatePeriod))
		currentProgress := d.getTotalProgress()
		newRate := float64(currentProgress-d.progress) / float64(progressUpdatePeriod)
		d.progressRate = d.progressRate*(1-movingAverageScale) + newRate*movingAverageScale
		d.progress = currentProgress

		// queues.signal
	}
}

func (d *defaultDownloader) getTotalProgress() int64 {
	total := int64(0)
	for _, handler := range d.chunkHandlers {
		total += handler.getRemaining()
	}
	return d.size - total
}

func (d *defaultDownloader) Start(_ context.Context) error {
	if d.isPaused {
		*d.pausedChan = make(chan int)
	}

	req, err := http.NewRequest("HEAD", d.url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error in requesting header: ", err)
		// TODO log.Fatal(err)
	}
	for k, vs := range resp.Header { // TODO test if
		fmt.Printf("%s: %d, %+v\n", k, len(vs), vs)
	}

	d.url = resp.Request.URL.String()

	writer := NewSynchronizedFileWriter(d.savePath)
	segmentsList := d.getChunkSegments(resp.Header)

	if len(d.chunkHandlers) != numberOfChuncks {
		chunkhandlersList := make([]*DownloadChunkHandler, 0)

		for _, segment := range segmentsList {
			l, r := segment[0], segment[1]

			handler := NewDownloadChunkHandler(state.DownloadChunk{
				ID:             uuid.NewString(),
				RangeStart:     l,
				RangeEnd:       r,
				CurrentPointer: l,
				DownloadID:     d.id,
			}, d.pausedChan)

			chunkhandlersList = append(chunkhandlersList, &handler)
		}

		d.chunkHandlers = chunkhandlersList
	}

	for _, handler := range d.chunkHandlers {
		handler.Start(d.url, &d.ticker, writer)
	}

	d.ticker.Start()

	go d.keepTrackOfProgress()
	return nil
}

func (d *defaultDownloader) getChunkSegments(header http.Header) [][]int64 {
	//TODO this is a prototype

	size, err := strconv.Atoi(header.Get("Content-Length"))
	d.size = int64(size)
	if err != nil {
		fmt.Println("no content length header", err) // TODO
	}

	if header.Get("Accept-Ranges") != "bytes" {
		fmt.Errorf("server does not accept range requests") // TODO need an if on test mode
		return [][]int64{{0, int64(size)}}
	}

	chunkSize := int(size / numberOfChuncks)

	segmentsList := make([][]int64, 0)

	for i := 0; chunkSize*i < size; i++ {
		segmentsList = append(segmentsList, []int64{int64(i * chunkSize), int64(min((i+1)*chunkSize, size))})
	}

	return segmentsList
}

func (d *defaultDownloader) Pause() error {
	if !d.isPaused {
		close(*d.pausedChan)
		d.isPaused = true
	}
	return nil
}
