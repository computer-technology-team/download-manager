package downloads

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

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
}

func NewDownloadChunkHandler(cfg state.DownloadChunk, pausedChan *chan int) DownloadChunkHandler {
	downChunk := DownloadChunkHandler{
		mainDownloadID: cfg.DownloadID,
		chunckID:       cfg.ID,
		rangeStart:     cfg.RangeStart,
		rangeEnd:       cfg.RangeEnd,
		currentPointer: cfg.CurrentPointer,
	}
	downChunk.pausedChan = pausedChan
	return downChunk
}

func (chunkHandler *DownloadChunkHandler) Start(url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {
	go chunkHandler.start(url, limiter, syncWriter)
}

func (chunkHandler *DownloadChunkHandler) start(url string, limiter *bandwidthlimit.Limiter, syncWriter *SynchronizedFileWriter) {

	conn, err := getConn(url) // connect to the proper host with the correct protocol
	if err != nil {
		fmt.Println("error in starting connection: ", err)
		//TODO handle error
	}

	writer := io.NewOffsetWriter(syncWriter, chunkHandler.currentPointer+chunkHandler.rangeStart)

	reader := bandwidthlimit.NewLimitedReader(context.Background(),
		sendRequest(url, conn, chunkHandler.rangeStart, chunkHandler.rangeEnd), limiter)
	for {
		<-*chunkHandler.pausedChan

		n, err := io.Copy(writer, reader)

		if err != nil {
			fmt.Println("Error reading:", err) //TODO
			return
		}

		chunkHandler.currentPointer += int64(n)

		if chunkHandler.currentPointer == chunkHandler.rangeEnd {
			break // TODO free wait list
		}
	}
}

func sendRequest(requestURL string, conn net.Conn, rangeStart, rangeEnd int64) io.ReadCloser { // send get and skip header
	parsedURL, _ := url.Parse(requestURL)

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Range: bytes=%d-%d\r\n"+
		"Connection: close\r\n\r\n", parsedURL.Path, parsedURL.Host, rangeStart, rangeEnd)
	_, _ = conn.Write([]byte(request))

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		fmt.Println("Error reading HTTP response:", err)
	}
	// Skip the headers and read the response body directly

	return resp.Body
}

func getConn(requestURL string) (net.Conn, error) {
	parsedURL, _ := url.Parse(requestURL) // TODO handle error

	// conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", parsedURL.Host))
	// if err != nil {
	// 	return nil, err
	// }

	// tlsConn := tls.Client(conn, &tls.Config{
	// 	InsecureSkipVerify: true, // Set to false in production for certificate verification
	// })

	// // Handshake and start the TLS connection
	// if err := tlsConn.Handshake(); err != nil {
	// 	fmt.Println("TLS handshake failed:", err) //
	// }

	tlsConn, err := tls.Dial("tcp", parsedURL.Host+":443", &tls.Config{
		InsecureSkipVerify: true,
	})

	if err == nil {
		return tlsConn, nil
	}

	conn, err := net.Dial("tcp", parsedURL.Host+":80")
	if err == nil {
		return conn, nil
	}

	return nil, err // failed to connect to either 443 or 80
}

func (DownloadHandler *DownloadChunkHandler) getRemaining() int64 {
	return DownloadHandler.rangeEnd - DownloadHandler.currentPointer
}
