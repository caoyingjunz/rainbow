package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/caoyingjunz/rainbow/pkg/pixiuctl/cmd"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cmd := cmd.NewDefaultPixiuCtlCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
