package downloads

import (
	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

func NewDownloadHandler(downloadConfig state.Download, downloadChuncks []state.DownloadChunk, ticker bandwidthlimit.Ticker) DownloadHandler {
	return nil
}

