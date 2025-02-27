package main

import (
	"github.com/computer-technology-team/download-manager.git/cmd"
	"github.com/computer-technology-team/download-manager.git/logging"
)

func main() {
	// TODO: do this in a better way and use env to decide to log or not
	onExit, err := logging.InitializeLogger()
	if err != nil {
		panic(err)
	}

	defer func() { _ = onExit() }()

	err = cmd.NewRootCmd().Execute()
	if err != nil {
		panic(err)
	}
}
