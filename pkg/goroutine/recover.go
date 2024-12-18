package goroutine

import (
	"errors"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/pkg/utils"
	"fmt"
	"github.com/getsentry/sentry-go"
)

func RecoverGO(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(fmt.Sprintf("recover from panic: %v", r))
				e, ok := r.(error)
				if !ok {
					e = errors.New("panic occurred: " + utils.ConvertToString(r) + "  " + version.GetBuildInfo())
				}
				sentry.CaptureException(e)
				logger.Sync()
			}
		}()
		f()
	}()
}
