package main

import (
	"errors"
	"fadacontrol/internal/base/application"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/pkg/utils"
	"github.com/getsentry/sentry-go"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Sync()
			e, ok := err.(error)
			if !ok {
				e = errors.New("panic occurred: " + utils.ConvertToString(err) + "  " + version.GetBuildInfo())
			}
			sentry.CaptureException(e)
		}
	}()
	err := sentry.Init(sentry.ClientOptions{
		Dsn:   "https://82431285059e21675920c08d0e172643@o4508488989605888.ingest.us.sentry.io/4508489034825728",
		Debug: true,
	})
	sentry.ConfigureScope(func(scope *sentry.Scope) {

		scope.SetTag("app_info", version.GetBuildInfo())

	})
	if err != nil {
		logger.Error(err)
	}
	defer sentry.Flush(5 * time.Second)
	application.Execute()
}
