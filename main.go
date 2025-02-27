package main

import "github.com/computer-technology-team/download-manager.git/cmd"

func main() {
	err := cmd.NewRootCmd().Execute()
	if err != nil {
		panic(err)
	}
}
