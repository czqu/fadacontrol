package main

import (
	"fadacontrol/internal/base/application"
	"fadacontrol/internal/base/logger"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Sync()
		}
	}()
	application.Execute()
}
