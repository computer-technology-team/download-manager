package downloads

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

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
	singlePart     bool
}

func NewDownloadChunkHandler(cfg state.DownloadChunk, pausedChan *chan int) *DownloadChunkHandler {
	downChunk := DownloadChunkHandler{
		mainDownloadID: cfg.DownloadID,
		chunckID:       cfg.ID,
		rangeStart:     cfg.RangeStart,
		rangeEnd:       cfg.RangeEnd,
		currentPointer: cfg.CurrentPointer,
		singlePart:     cfg.SinglePart,
	}
	downChunk.pausedChan = pausedChan
	return &downChunk
}

func (chunkHandler *DownloadChunkHandler) Start(ctx context.Context, url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {
	go chunkHandler.start(ctx, url, limiter, syncWriter)
}

func (chunkHandler *DownloadChunkHandler) start(ctx context.Context, url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {

	slog.Info("chunk handler started", "range_start", chunkHandler.rangeStart,
		"range_end", chunkHandler.rangeEnd, "current_pointer", chunkHandler.currentPointer)

	writer := io.NewOffsetWriter(syncWriter, chunkHandler.currentPointer)

	// Use the new sendRequest that uses standard HTTP library
	resp, err := chunkHandler.sendRequest(ctx, url, chunkHandler.currentPointer, chunkHandler.rangeEnd)
	if err != nil {
		slog.Error("error sending request", "error", err)
		//TODO handle error properly
		return
	}
	defer resp.Body.Close()

	reader := bandwidthlimit.NewLimitedReader(ctx, resp.Body, limiter)

	for {
		<-*chunkHandler.pausedChan

		n, err := io.CopyN(writer, reader, 1<<14)
		if err != nil {
			if err == io.EOF {
				chunkHandler.currentPointer += int64(n)
				break
			}
			slog.Error("error reading from response", "error", err)
			return
		}

		chunkHandler.currentPointer += int64(n)
		if chunkHandler.currentPointer >= chunkHandler.rangeEnd {
			break // TODO free wait list
		}
	}
}

func (chunkHandler *DownloadChunkHandler) sendRequest(ctx context.Context, requestURL string, rangeStart, rangeEnd int64) (*http.Response, error) {
	// Create a new HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if !chunkHandler.singlePart {
		// Set the Range header for partial content
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd-1))
	}

	// Create a custom transport with reasonable timeouts
	transport := &http.Transport{
		DisableCompression: true,
	}

	// Create a client with the custom transport
	client := &http.Client{
		Transport: transport,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check if we got a successful response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned non-success status: %s", resp.Status)
	}

	return resp, nil
}

// This function is no longer needed as we're using the standard HTTP library
// which handles connections automatically

func (DownloadHandler *DownloadChunkHandler) getRemaining() int64 {
	return DownloadHandler.rangeEnd - DownloadHandler.currentPointer
}
