package main

import (
	"fadacontrol/internal/base/cmd"
	"fadacontrol/internal/base/logger"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Sync()
		}
	}()
	cmd.Execute()
}
