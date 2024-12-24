package internal_service

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/log"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema/internal_command"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/custom_command_service"
	"fadacontrol/pkg/goroutine"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"
	"os/user"

	"google.golang.org/grpc"
	"os"
	"strconv"
	"time"
)

type InternalSlaveService struct {
	_done chan bool

	ctx         context.Context
	cu          *custom_command_service.CustomCommandService
	co          *control_pc.ControlPCService
	_exitSignal *conf.ExitChanStruct
}

func NewInternalSlaveService(_exitSignal *conf.ExitChanStruct, cu *custom_command_service.CustomCommandService, co *control_pc.ControlPCService, ctx context.Context) *InternalSlaveService {
	return &InternalSlaveService{_exitSignal: _exitSignal, cu: cu, co: co, ctx: ctx, _done: make(chan bool)}
}
func (s *InternalSlaveService) Start() {
	port := 2095
	host := "127.0.0.1"
	addr := host + ":" + strconv.Itoa(port)
	goroutine.RecoverGO(func() {
		s.connectToServer(addr)
		os.Exit(-1)
	})

}
func (s *InternalSlaveService) Stop() {
	s._done <- true
}

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 8 * time.Second
)

func (s *InternalSlaveService) connectToServer(addr string) {
	defer func() {
		logger.Info("slave will exit")
		os.Exit(-1)
	}()
	logger.Info("slave connecting.")
	backoff := initialBackoff

	for {

		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		logger.Info("slave connecting..")
		select {
		case <-s._done:
			return
		default:
			break
		}
		logger.Info("slave connecting...")
		if err != nil || conn == nil {
			logger.Infof("Error connecting to server: %v\n", err)

			logger.Infof("will sleep %v\n", backoff)
			// Wait for the backoff time and try again
			time.Sleep(backoff)

			// Increase the backoff time until maxBackoff is reached
			if backoff < maxBackoff {
				backoff *= 2

			}
			if backoff >= maxBackoff {

				logger.Warn("max back off,will exit")
				logger.Info("slave exit")

				return

			}
			continue
		}

		logger.Info("slave connected")

		backoff = maxBackoff / 2
		client := internal_command.NewBaseClient(conn)
		_user, err := user.Current()
		if err != nil {
			logger.Error("Error getting current user: %v\n", err)
			return
		}
		registerResp, err := client.RegisterClient(context.Background(), &internal_command.ClientInfo{
			Username: _user.Username,
		})

		if err != nil || registerResp.Code != int32(exception.ErrSuccess.Code) {
			logger.Warnf("Error registering client: %v", err)
			return
		}
		var registerClientResponse internal_command.RegisterClientResponse
		err = registerResp.GetData().UnmarshalTo(&registerClientResponse)
		if err != nil {
			logger.Warnf("Error unmarshalling registerClientResponse: %v", err)
			return
		}
		clientId := registerClientResponse.ClientId
		logger.Debug("register success clientId:", clientId)
		sentryResp, err := client.GetSentryOptions(context.Background(), &internal_command.GetSentryOptionsRequest{})
		if err != nil || sentryResp.Code != int32(exception.ErrSuccess.Code) {
			logger.Warnf("Error getting sentry options: %v", err)
			return
		}
		var sentryOptions internal_command.SentryOptions
		err = sentryResp.GetData().UnmarshalTo(&sentryOptions)
		if err != nil {
			logger.Warnf("Error unmarshalling sentry options: %v", err)
			return
		}
		opt := &log.SentryOptions{}
		opt.Enable = sentryOptions.Enable
		opt.Level = sentryOptions.Level
		opt.UserId = sentryOptions.UserId
		opt.ProfilesSampleRate = sentryOptions.ProfilesSampleRate
		opt.TracesSampleRate = sentryOptions.TracesSampleRate
		logger.InitLogReporter(opt)
		logger.Debug("Sentry init success")
		executeClient := internal_command.NewExecuteCommandClient(conn)
		if err != nil {
			logger.Fatal("Failed to create ExecuteCommand client:", err)
		}
		md := metadata.New(map[string]string{
			constants.ClientIdKey: clientId,
		})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		stream, err := executeClient.RegisterInternalCommand(ctx)
		if err != nil {
			logger.Fatal("Failed to create stream:", err)
		}

		// Receive response from server
		for {
			rpcData, err := stream.Recv()
			if err != nil {
				logger.Fatal("Failed to receive response:", err)
				return
			}

			if rpcData.GetType() == internal_command.StreamMessageType_Unknown {
				continue
			}
			switch rpcData.GetType() {
			case internal_command.StreamMessageType_LockPcRequest:
				lockErr := s.co.LockWindows(false)
				if !lockErr.Equal(exception.ErrSuccess) {
					logger.Warnf("LockWindows err: %v", err)
					resp := &internal_command.RpcResponse{
						Code:    int32(lockErr.Code),
						Message: lockErr.Error(),
					}
					respData, err := anypb.New(resp)
					if err != nil {
						logger.Warnf("Error marshalling rpc response: %v", err)
						continue
					}
					stream.Send(&internal_command.RpcStream{
						Type: internal_command.StreamMessageType_Response,
						Data: respData,
					})
				}
			case internal_command.StreamMessageType_ExitProcessRequest:
				logger.Info("recv exit cmd,exit process")

				s._exitSignal.ExitChan <- 0
				<-s._done
				return
			}

		}

	}
}
