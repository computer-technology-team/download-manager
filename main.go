package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/computer-technology-team/download-manager.git/cmd"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
)

var testOneHandler bool = false
var testDownload bool = true

func main() {

	if testDownload {
		downloader := downloads.NewDownloader(downloads.DownloaderConfig{
			URL:                   "http://calculus.math.sharif.ir/gm2_2025/Exercises/GM1-01.pdf",
			SavePath:              "test.txt",
			BandwidthLimitBytesPS: 100000,
		}, nil)
		downloader.Start()
		time.Sleep(3 * 1000 * time.Millisecond)
		downloader.GetTicker().Quite()
		time.Sleep(3 * 1000 * time.Millisecond)
		downloader.GetTicker().Start()
		time.Sleep(10 * 1000 * time.Millisecond) // WTF
	} else if testOneHandler {
		a := make([][]int, 0)
		a = append(a, []int{1, 1})
		fmt.Println(a)
		// handler:= downloads.NewDownloadChunkHandler()
		req, err := http.NewRequest("HEAD", "https://www.w3schools.com/sql/sql_insert.asp", nil)
		// req.Header.Set("Test-Header", "testtsett")
		if err != nil {
			log.Fatal(err)
		}
		req.Write(os.Stdout)
		fmt.Println("==========")
		fmt.Println("Host:", req.Host)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("\n\n===== RESPONSE")
		fmt.Println("status code = ", resp.StatusCode)
		for k, vs := range resp.Header {
			fmt.Printf("%s: %d, %+v\n", k, len(vs), vs)
		}
		ticker := downloads.NewTicker()
		writer := downloads.NewSynchronizedFileWriter("test.txt")
		handler := downloads.NewDownloadChunkHandler(downloads.DownloaderConfig{
			URL:                   "ocw.sharif.ir/uploads/57328ad75bf66befcaebf8c44946647a.mp4",
			SavePath:              "test.txt", // ?
			BandwidthLimitBytesPS: 1000000,    // ?
		},
			1e3, 2e3,
			&ticker,
			writer)
		ticker.Start()
		handler.Start()
		time.Sleep(10 * 1000 * time.Millisecond)
		// writer.Close()
		// fmt.Println("=========================")
		// sc := bufio.NewScanner(resp.Body)
		// for sc.Scan() {
		// 	fmt.Println(sc.Text())
		// }
	} else {
		err := cmd.NewRootCmd().Execute()
		if err != nil {
			panic(err)
		}
	}
}
