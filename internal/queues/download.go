package queues

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadFile downloads a file in chunks and writes it to the specified path.
func DownloadFile(url, filePath string, tokenChan <-chan struct{}) error {
	// Create or open the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Send an HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Download the file in chunks
	buf := make([]byte, 1024) // 1 KB chunk size
	for {
		// Wait for a token to download the next chunk
		<-tokenChan

		// Read a chunk of data
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading chunk: %v", err)
		}
		if n == 0 {
			break
		}

		// Write the chunk to the file
		if _, err := file.Write(buf[:n]); err != nil {
			return fmt.Errorf("error writing chunk: %v", err)
		}
	}

	fmt.Printf("Download completed: %s -> %s\n", url, filePath)
	return nil
}
