package downloads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

type DownloadChunkHandler struct {
	mainDownloadID int64
	chunckID       string
	rangeStart     int64
	rangeEnd       int64
	currentPointer int64
	pausedChan     *chan int
	failedChan     chan error
	wg             *sync.WaitGroup
	singlePart     bool
}

func NewDownloadChunkHandler(cfg state.DownloadChunk,
	pausedChan *chan int, failedChan chan error, wg *sync.WaitGroup) *DownloadChunkHandler {
	downChunk := DownloadChunkHandler{
		mainDownloadID: cfg.DownloadID,
		chunckID:       cfg.ID,
		rangeStart:     cfg.RangeStart,
		rangeEnd:       cfg.RangeEnd,
		currentPointer: cfg.CurrentPointer,
		singlePart:     cfg.SinglePart,
		wg:             wg,
		failedChan:     failedChan,
		pausedChan:     pausedChan,
	}
	return &downChunk
}

func (chunkHandler *DownloadChunkHandler) Start(ctx context.Context, url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {
	chunkHandler.wg.Add(1)
	go chunkHandler.start(ctx, url, limiter, syncWriter)
}

func (chunkHandler *DownloadChunkHandler) start(ctx context.Context, url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {
	defer chunkHandler.wg.Done()

	writer := io.NewOffsetWriter(syncWriter, chunkHandler.currentPointer)

	resp, err := chunkHandler.sendRequest(ctx, url, chunkHandler.currentPointer, chunkHandler.rangeEnd)
	if err != nil {
		slog.Error("error sending request", "error", err)

		chunkHandler.failedChan <- err

		return
	}
	defer resp.Body.Close()

	reader := bandwidthlimit.NewLimitedReader(ctx, resp.Body, limiter)

	for {
		select {
		case <-ctx.Done():
			return
		case <-*chunkHandler.pausedChan:
			return
		default:
			n, err := io.CopyN(writer, reader, 1<<14)
			chunkHandler.currentPointer += int64(n)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				if errors.Is(err, context.Canceled) {
					return
				}

				slog.Error("error reading from response", "error", err)
				chunkHandler.failedChan <- err
				return
			}

			if chunkHandler.currentPointer >= chunkHandler.rangeEnd {
				return
			}
		}

	}
}

func (chunkHandler *DownloadChunkHandler) sendRequest(ctx context.Context, requestURL string, rangeStart, rangeEnd int64) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if !chunkHandler.singlePart {

		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd-1))
	}

	transport := &http.Transport{
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned non-success status: %s", resp.Status)
	}

	return resp, nil
}

func (DownloadHandler *DownloadChunkHandler) getRemaining() int64 {
	return DownloadHandler.rangeEnd - DownloadHandler.currentPointer
}
