package downloads

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

const progressUpdatePeriod int = 1

const movingAverageScale float64 = .75 // new average = old * (1 - alpha) + current * alpha
const numberOfChuncks = 10

type defaultDownloader struct {
	id            int64
	queueID       int64
	url           string
	savePath      string
	state         DownloadState
	limiter       *bandwidthlimit.Limiter
	chunkHandlers []*DownloadChunkHandler
	progress      int64
	progressRate  float64
	size          int64
	pausedChan    *chan int
	ctx           context.Context
	ctxCancel     context.CancelFunc
	writer        *SynchronizedFileWriter
}

func (d *defaultDownloader) keepTrackOfProgress() {
	d.reportProgress()
	for {
		select {
		case <-d.ctx.Done():
			d.reportProgress()
			return
		case <-time.After(time.Second * time.Duration(progressUpdatePeriod)):
			d.reportProgress()
		}
	}
}

func (d *defaultDownloader) reportProgress() {
	if d.state == StateInProgress {
		currentProgress := d.getTotalProgress()
		newRate := float64(currentProgress-d.progress) / float64(progressUpdatePeriod)
		d.progressRate = d.progressRate*(1-movingAverageScale) + newRate*movingAverageScale
		d.progress = currentProgress
	}

	events.GetEventChannel() <- events.Event{
		EventType: events.DownloadProgressed,
		Payload:   d.status(),
	}
}

func (d *defaultDownloader) getTotalProgress() int64 {
	total := int64(0)
	for _, handler := range d.chunkHandlers {
		total += handler.getRemaining()
	}
	return d.size - total
}

func (d *defaultDownloader) Start() error {
	if d.state == StatePaused {
		close(*d.pausedChan)
	}

	d.state = StateInProgress
	d.ctx, d.ctxCancel = context.WithCancel(context.Background())

	req, err := http.NewRequest("HEAD", d.url, nil)
	if err != nil {
		return err
	}

	httpClient := http.Client{Transport: &http.Transport{
		DisableCompression: true,
	}}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error in requesting header: ", err)
		// TODO log.Fatal(err)
	}

	d.url = resp.Request.URL.String()

	d.writer = NewSynchronizedFileWriter(d.savePath)
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
		handler.Start(d.ctx, d.url, d.limiter, d.writer)
	}

	d.reportProgress()
	go d.keepTrackOfProgress()
	return nil
}

func (d *defaultDownloader) getChunkSegments(header http.Header) [][]int64 {
	//TODO this is a prototype

	size, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64)
	d.size = size
	if err != nil {
		fmt.Println("no content length header", err) // TODO
	}

	if header.Get("Accept-Ranges") != "bytes" {
		fmt.Println("server does not accept range requests") // TODO need an if on test mode
		return [][]int64{{0, int64(size)}}
	}

	chunkSize := int64(math.Ceil(float64(size) / numberOfChuncks))

	segmentsList := make([][]int64, 0)

	var i int64
	for ; chunkSize*i < size; i++ {
		segmentsList = append(segmentsList, []int64{i * chunkSize, min((i+1)*chunkSize, size)})
	}
	fmt.Println(len(segmentsList))

	return segmentsList
}

func (d *defaultDownloader) Pause() error {
	if d.state == StateInProgress {
		if d.ctxCancel != nil {
			d.ctxCancel()
		}

		d.pausedChan = lo.ToPtr(make(chan int))
		d.state = StatePaused
		d.writer.Close()
	}
	return nil
}

func (d *defaultDownloader) Cancel() error {
	d.Pause()

	if d.ctxCancel != nil {
		d.ctxCancel()
	}

	path, _ := filepath.Abs(d.savePath)

	if path != "" {
		if err := os.Remove(path); err != nil {
			fmt.Println("couldn't delete")
			return err
		}
	}

	return nil
}

func (d *defaultDownloader) status() DownloadStatus {
	return DownloadStatus{
		ID:                 d.id,
		ProgressPercentage: (float64(d.progress) / float64(d.size)) * 100,
		Speed:              float64(d.progressRate),
		State:              d.state,
	}
}
