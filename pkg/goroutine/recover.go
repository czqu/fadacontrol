package goroutine

import (
	"fadacontrol/internal/base/logger"
	"fmt"
)

func RecoverGO(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf(fmt.Sprintf("recover from panic: %v", r))
			}
		}()
		f()
	}()
}
