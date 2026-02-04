package main

import (
	"os"

	"github.com/rodrigolobo/st/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
