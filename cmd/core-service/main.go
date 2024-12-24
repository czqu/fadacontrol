package main

import (
	"fadacontrol/internal/base/application"
	"fadacontrol/internal/base/logger"
	"github.com/getsentry/sentry-go"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Fatal(err)
			logger.Sync()
			sentry.CurrentHub().Recover(err)
			sentry.Flush(time.Second * 5)
		}
	}()

	application.Execute()
}
