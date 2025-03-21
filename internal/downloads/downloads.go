package downloads

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

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
	wg            sync.WaitGroup
	failedChannel chan error
}

func (d *defaultDownloader) keepTrackOfProgress() {
	d.reportProgress()
	for {
		select {
		case <-d.ctx.Done():
			slog.Info("context canceled in keep track")
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
		if d.progress == d.size {
			d.state = StateCompleted
			events.GetEventChannel() <- events.Event{
				EventType: events.DownloadCompleted,
				Payload:   d.status(),
			}
			d.ctxCancel()
		} else {
			events.GetEventChannel() <- events.Event{
				EventType: events.DownloadProgressed,
				Payload:   d.status(),
			}
		}
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
		return fmt.Errorf("could not get headers from url %s: %w", d.url, err)
	}

	d.writer = NewSynchronizedFileWriter(d.savePath)

	d.size, err = getContentSize(resp.Header)
	if err != nil {
		slog.Error("could not get content size", "error", err)
		return fmt.Errorf("could not get content size from url %s: %w", d.url, err)
	}

	var segmentsList [][]int64
	var acceptsRanges bool

	if doesAccpetRanges(resp) {
		acceptsRanges = true
		segmentsList = d.getChunkSegments()
	} else {
		segmentsList = [][]int64{{0, d.size}}
	}

	if d.chunkHandlers == nil {
		chunkhandlersList := make([]*DownloadChunkHandler, 0)

		for _, segment := range segmentsList {
			l, r := segment[0], segment[1]

			handler := NewDownloadChunkHandler(state.DownloadChunk{
				ID:             uuid.NewString(),
				RangeStart:     l,
				RangeEnd:       r,
				CurrentPointer: l,
				DownloadID:     d.id,
				SinglePart:     !acceptsRanges,
			}, d.pausedChan, d.failedChannel, &d.wg)

			chunkhandlersList = append(chunkhandlersList, handler)
		}

		d.chunkHandlers = chunkhandlersList
	}

	for _, handler := range d.chunkHandlers {
		handler.Start(d.ctx, d.url, d.limiter, d.writer)
	}

	d.reportProgress()
	go d.keepTrackOfProgress()
	go d.listenForFailiure()
	return nil
}

func (d *defaultDownloader) getChunkSegments() [][]int64 {

	chunkSize := int64(math.Ceil(float64(d.size) / numberOfChuncks))

	segmentsList := make([][]int64, 0)

	var i int64
	for ; chunkSize*i < d.size; i++ {
		segmentsList = append(segmentsList, []int64{i * chunkSize, min((i+1)*chunkSize, d.size)})
	}

	return segmentsList
}

func getContentSize(header http.Header) (int64, error) {
	contentLength := header.Get("Content-Length")
	if contentLength == "" {
		return 0, errors.New("response does not have Content-Length")
	}
	return strconv.ParseInt(contentLength, 10, 64)
}

func (d *defaultDownloader) Pause() error {
	if d.ctxCancel != nil {
		d.ctxCancel()
		slog.Info("context canceled")
	}
	close(*d.pausedChan)

	d.wg.Wait()
	d.writer.Close()

	slog.Info("paused")
	return nil
}

func (d *defaultDownloader) Cancel() error {
	err := d.Pause()
	if err != nil {
		return fmt.Errorf("could not pause download: %w", err)
	}

	if d.ctxCancel != nil {
		d.ctxCancel()
	}

	if d.savePath != "" {
		if err := os.Remove(d.savePath); err != nil {
			return fmt.Errorf("could not delete file: %w", err)
		}
	}

	return nil
}

func (d *defaultDownloader) status() DownloadStatus {
	status := DownloadStatus{
		ID:                 d.id,
		ProgressPercentage: (float64(d.progress) / float64(d.size)) * 100,
		Speed:              float64(d.progressRate),
		State:              d.state,
		DownloadChuncks:    nil,
	}

	chunkList := make([]state.DownloadChunk, numberOfChuncks)

	for i, chunkHandler := range d.chunkHandlers {
		downloadChunk := state.DownloadChunk{
			ID:             chunkHandler.chunckID,
			RangeStart:     chunkHandler.rangeStart,
			RangeEnd:       chunkHandler.rangeEnd,
			CurrentPointer: chunkHandler.currentPointer,
			DownloadID:     d.id,
		}
		chunkList[i] = downloadChunk
	}

	status.DownloadChuncks = chunkList

	return status
}

func (d *defaultDownloader) listenForFailiure() {
	select {
	case err := <-d.failedChannel:
		slog.Info("failed channel")
		_ = d.Pause()

		events.GetEventChannel() <- events.Event{
			EventType: events.DownloadFailed,
			Payload:   events.DownloadFailedEvent{Error: err},
		}
		return
	case <-d.ctx.Done():
		return
	}

}

func doesAccpetRanges(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	return resp.Header.Get("Accept-Ranges") == "bytes"
}
