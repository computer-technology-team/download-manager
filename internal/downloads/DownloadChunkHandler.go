package downloads

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

type DownloadChunkHandler struct {
	url            string
	rangeStart     int64
	rangeEnd       int64
	savePath       string
	bandwidthLimit int64
	ticker         *Ticker
	currentPointer int64
	syncWriter     SynchronizedFileWriter
}

func NewDownloadChunkHandler(cfg DownloaderConfig, rangeL int64, rangeR int64, sharedTicker *Ticker, syncWriter SynchronizedFileWriter) DownloadChunkHandler {
	return DownloadChunkHandler{
		url:            cfg.URL,
		rangeStart:     rangeL,
		rangeEnd:       rangeR,
		savePath:       cfg.SavePath,
		bandwidthLimit: cfg.BandwidthLimitBytesPS,
		ticker:         sharedTicker,
		currentPointer: rangeL,
		syncWriter:     syncWriter,
	}
}
func (chunkHandler *DownloadChunkHandler) Start() {
	go chunkHandler.start()
}

func (chunkHandler *DownloadChunkHandler) start() {

	temp, err := getConn(chunkHandler.url) // connect to the proper host with the correct protocol

	if err != nil {
		fmt.Println("error in starting connection: ", err)
		//TODO handle error
	}

	conn := *temp                                                                                  // TODO would copying be a problem?
	reader := sendRequest(chunkHandler.url, &conn, chunkHandler.rangeStart, chunkHandler.rangeEnd) // send get request
	buffer := make([]byte, 4096)
	for {
		chunkHandler.ticker.GetToken()
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading:", err) //TODO
			return
		}
		n = int(min(int64(n), chunkHandler.rangeEnd-chunkHandler.currentPointer))

		chunkHandler.syncWriter.Write(buffer, chunkHandler.currentPointer, int64(n))
		chunkHandler.currentPointer += int64(n)

		if chunkHandler.currentPointer == chunkHandler.rangeEnd {
			break // TODO free wait list
		}
	}
}

func sendRequest(requestURL string, conn *tls.Conn, rangeStart, rangeEnd int64) io.Reader { // send get and skip header
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

func getConn(requestURL string) (*tls.Conn, error) {
	parsedURL, _ := url.Parse(requestURL) // TODO handle error
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", parsedURL.Host))
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: true, // Set to false in production for certificate verification
	})

	// Handshake and start the TLS connection
	if err := tlsConn.Handshake(); err != nil {
		fmt.Println("TLS handshake failed:", err) //TODO
	}
	fmt.Println("!")

	return tlsConn, nil
}
