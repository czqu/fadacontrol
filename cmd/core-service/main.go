package main

import (
	"fadacontrol/internal/base/cmd"
	"fadacontrol/internal/base/logger"
	"time"
)

func main() {

	go func() {
		time.Sleep(5 * time.Second)
		i := 0
		for {
			time.Sleep(5000 * time.Millisecond)
			logger.Info("test: ", i)
			i++
		}

	}()
	cmd.Execute()
}
