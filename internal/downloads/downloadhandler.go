package downloads

import (
	"fmt"
	"os"
	"sync"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

func NewDownloadHandler(downloadConfig state.Download, downloadChuncks []state.DownloadChunk, limiter *bandwidthlimit.Limiter) (DownloadHandler, error) {

	savePath := downloadConfig.SavePath

	if _, err := os.Stat(savePath); err == nil {
		return nil, fmt.Errorf("file already exists at %s", savePath)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking file at %s: %w", savePath, err)
	}
	pausedChan := make(chan int, 1)

	defDow := defaultDownloader{
		id:            downloadConfig.ID,
		queueID:       downloadConfig.QueueID,
		url:           downloadConfig.Url,
		savePath:      downloadConfig.SavePath,
		state:         DownloadState(downloadConfig.State),
		limiter:       limiter,
		chunkHandlers: nil,
		progress:      0,
		progressRate:  0,
		size:          0,
		pausedChan:    nil,
		ctx:           nil,
		ctxCancel:     nil,
		writer:        NewSynchronizedFileWriter(downloadConfig.SavePath),
		failedChannel: make(chan error, numberOfChuncks),
		wg:            sync.WaitGroup{},
	}

	defDow.pausedChan = &pausedChan

	if len(downloadChuncks) == numberOfChuncks {
		chunkhandlersList := make([]*DownloadChunkHandler, numberOfChuncks)

		for i, chunk := range downloadChuncks {

			handler := NewDownloadChunkHandler(chunk, defDow.pausedChan, defDow.failedChannel, &defDow.wg)

			chunkhandlersList[i] = handler
		}
		defDow.chunkHandlers = chunkhandlersList
	} else if len(downloadChuncks) == 1 && downloadChuncks[0].SinglePart {
		defDow.chunkHandlers = []*DownloadChunkHandler{
			NewDownloadChunkHandler(downloadChuncks[0], defDow.pausedChan, defDow.failedChannel, &defDow.wg),
		}
	}

	return &defDow, nil
}
