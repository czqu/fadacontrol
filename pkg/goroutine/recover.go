package goroutine

import (
	"fadacontrol/internal/base/logger"
	"fmt"
	"github.com/getsentry/sentry-go"
	"time"
)

func RecoverGO(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Fatal(fmt.Sprintf("recover from panic: %v", r))
				logger.Sync()
				sentry.CurrentHub().Recover(r)
				sentry.Flush(time.Second * 5)
			}
		}()
		f()
	}()
}
