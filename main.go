// package main
//
// import (
//
//	"github.com/computer-technology-team/download-manager.git/cmd"
//	"github.com/computer-technology-team/download-manager.git/logging"
//
// )
//
//	func main() {
//		// TODO: do this in a better way and use env to decide to log or not
//		onExit, err := logging.InitializeLogger()
//		if err != nil {
//			panic(err)
//		}
//
//		defer func() { _ = onExit() }()
//
//		err = cmd.NewRootCmd().Execute()
//		if err != nil {
//			panic(err)
//		}
//	}
package main

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/bandwidthlimit"
	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/state"
)

// "github.com/computer-technology-team/download-manager.git/cmd"
// "github.com/computer-technology-team/download-manager.git/logging"

func main() {
	// TODO: do this in a better way and use env to decide to log or not

	// onExit, err := logging.InitializeLogger()
	// if err != nil {
	//  panic(err)
	// }

	// defer func() { _ = onExit() }()

	// err = cmd.NewRootCmd().Execute()
	// if err != nil {
	//  panic(err)
	// }
	downloadHandler := downloads.NewDownloadHandler(state.Download{
		ID:       0,
		QueueID:  0,
		Url:      "https://www.gutenberg.org/files/1342/1342-0.txt",
		SavePath: `1342-0.txt`,
		State:    string(downloads.StateInProgress),
		Retries:  0,
	},
		nil,
		bandwidthlimit.NewLimiter(lo.ToPtr(bandwidthlimit.DefaultBandwidth/10)))

	downloadHandler.Start()

	go func() {
		for event := range events.GetEventChannel() {
			if event.EventType == events.DownloadProgressed {
				progress := event.Payload.(downloads.DownloadStatus)
				fmt.Printf("State: %s, Progress: %.6f %%, Speed: %.6f MB/s",
					progress.State, progress.ProgressPercentage, progress.Speed/1024/1024)
				fmt.Println()
			}
		}
	}()

	time.Sleep(time.Second * 3)

	downloadHandler.Pause()

	time.Sleep(time.Second * 2)

	downloadHandler.Start()

	time.Sleep(time.Hour)

}
