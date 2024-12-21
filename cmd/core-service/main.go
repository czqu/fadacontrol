package main

import (
	"fadacontrol/internal/base/application"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Fatal(err)
			logger.Sync()
		}
	}()
	utils.NetworkChangeCallbackInit()
	application.Execute()
}
