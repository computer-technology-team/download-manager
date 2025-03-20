package downloads

import (
	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

func NewDownloadHandler(downloadConfig state.Download, downloadChuncks []state.DownloadChunk, ticker bandwidthlimit.Ticker) DownloadHandler {
	pausedChan := make(chan int, 1)

	defDow := defaultDownloader{
		id:            downloadConfig.ID,
		queueID:       downloadConfig.QueueID,
		url:           downloadConfig.Url,
		savePath:      downloadConfig.SavePath,
		state:         DownloadState(downloadConfig.State),
		ticker:        ticker,
		chunkHandlers: nil,
		progress:      0,
		progressRate:  0,
		size:          0,
		pausedChan:    nil,
		ctx:           nil,
		ctxCancel:     nil,
		writer:        NewSynchronizedFileWriter(downloadConfig.SavePath),
	}

	defDow.pausedChan = &pausedChan

	if len(downloadChuncks) == numberOfChuncks {
		chunkhandlersList := make([]*DownloadChunkHandler, numberOfChuncks)

		for i, chunk := range downloadChuncks {

			handler := NewDownloadChunkHandler(state.DownloadChunk{
				ID:             chunk.ID,
				RangeStart:     chunk.RangeStart,
				RangeEnd:       chunk.RangeEnd,
				CurrentPointer: chunk.CurrentPointer,
				DownloadID:     chunk.DownloadID,
			}, defDow.pausedChan)

			chunkhandlersList[i] = &handler
		}
		defDow.chunkHandlers = chunkhandlersList
	}

	return &defDow
}
