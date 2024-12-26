package goroutine

import (
	"context"
	"fadacontrol/internal/base/logger"
	"fmt"
	"github.com/getsentry/sentry-go"
	"google.golang.org/grpc"
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
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); err != nil {
			logger.Fatal(fmt.Sprintf("recover from panic in rpc: %v", r))
			logger.Sync()
			sentry.CurrentHub().Recover(r)
			sentry.Flush(time.Second * 5)
		}
	}()

	resp, err = handler(ctx, req)
	return resp, err
}
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Fatal(fmt.Sprintf("recover from panic in rpc stream: %v", r))
			logger.Sync()
			sentry.CurrentHub().Recover(r)
			sentry.Flush(time.Second * 5)
		}
	}()

	// Call the stream handler
	err := handler(srv, ss)
	return err
}
