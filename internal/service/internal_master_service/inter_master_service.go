package internal_master_service

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/schema/internal_command"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	_ "time"
)

type InternalMasterService struct {
	ctx          context.Context
	slavePath    string
	slaveWorkDir string
	startOnce    sync.Once
	stopOnce     sync.Once
}

func NewInternalMasterService(ctx context.Context) *InternalMasterService {
	return &InternalMasterService{ctx: ctx}

}
func (s *InternalMasterService) Start() error {
	logger.Info("starting slave program")

	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get executable path %v", err)
	}

	s.slaveWorkDir = filepath.Dir(path)
	s.slavePath = path
	logger.Info("slave program path: ", s.slavePath)
	logger.Sync()
	s.startOnce.Do(func() {

		goroutine.RecoverGO(func() {
			s.StartServer()
		})
	})

	return nil

}

func (s *InternalMasterService) Stop() error {
	s.stopOnce.Do(func() {
		s.StopServer()
	})
	return nil
}

type internalRpcServer struct {
	ProgramConf *conf.Conf
	internal_command.UnimplementedBaseServer
	internal_command.UnimplementedExecuteCommandServer
}
type rpcClientStatus int

const (
	Unknown rpcClientStatus = iota
	Connected
	Disconnected
)

type rpcClient struct {
	Id         string
	Username   string
	status     rpcClientStatus
	SendChan   chan *schema.InternalCommand
	statusLock sync.RWMutex
}

func (r *rpcClient) Connect() {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	r.status = Connected

}
func (r *rpcClient) Status() rpcClientStatus {
	r.statusLock.RLock()
	defer r.statusLock.RUnlock()
	return r.status
}
func (r *rpcClient) Close() {
	r.statusLock.Lock()
	defer r.statusLock.Unlock()
	if r.status != Disconnected {
		activeRpcClientMapLock.Lock()
		delete(activeRpcClientMap, r.Id)
		activeRpcClientMapLock.Unlock()
		usernameClientMapLock.Lock()
		delete(usernameClientMap, r.Username)
		usernameClientMapLock.Unlock()
		select {
		case r.SendChan <- &schema.InternalCommand{
			CommandType: schema.ExitCommandType,
			Data:        nil,
		}:
			logger.Info("send exit command")
		default:
			logger.Info("Command channel is full")
		}
		logger.Warn("close rpc client")
		close(r.SendChan)
	}

	r.status = Disconnected
}

var activeRpcClientMap = make(map[string]*rpcClient)
var activeRpcClientMapLock sync.RWMutex
var usernameClientMap = make(map[string]*rpcClient)
var usernameClientMapLock sync.RWMutex

func (s *internalRpcServer) RegisterClient(ctx context.Context, req *internal_command.ClientInfo) (*internal_command.RpcResponse, error) {
	// Handle the RegisterClient RPC
	logger.Debug("Registering client from ", req.Username)
	usernameClientMapLock.RLock()

	_, ok := usernameClientMap[req.Username]
	usernameClientMapLock.RUnlock()
	if ok {
		logger.Debug("Duplicate client from ", req.Username)
		return &internal_command.RpcResponse{
			Code:    int32(exception.ErrUserAlreadyExistsOneSlave.Code),
			Message: exception.ErrUserAlreadyExistsOneSlave.Error(),
			Data:    nil,
		}, nil
	}

	client := &rpcClient{
		Id:       uuid.NewString(),
		Username: req.Username,
		status:   Unknown,
		SendChan: make(chan *schema.InternalCommand, 10),
	}
	respData := &internal_command.RegisterClientResponse{
		ClientId: client.Id,
	}
	data, err := anypb.New(respData)
	if err != nil {
		logger.Error(err)
		return &internal_command.RpcResponse{
			Code:    int32(exception.ErrUnknownException.Code),
			Message: err.Error(),
			Data:    nil,
		}, err
	}

	activeRpcClientMapLock.Lock()
	activeRpcClientMap[client.Id] = client
	activeRpcClientMapLock.Unlock()
	usernameClientMapLock.Lock()
	usernameClientMap[client.Username] = client
	usernameClientMapLock.Unlock()

	defer func() {
		logger.Debug("success Registering client from ", req.Username)
	}()
	return &internal_command.RpcResponse{
		Code:    int32(exception.ErrSuccess.Code),
		Message: exception.ErrSuccess.Error(),
		Data:    data,
	}, nil
}

func (s *internalRpcServer) GetSentryOptions(ctx context.Context, req *internal_command.GetSentryOptionsRequest) (*internal_command.RpcResponse, error) {
	if s.ProgramConf == nil || s.ProgramConf.LogReporterOpt == nil {
		logger.Warn("Sentry options not set")
		return &internal_command.RpcResponse{
			Code:    int32(exception.ErrUserResourceNotFound.Code),
			Message: exception.ErrUserResourceNotFound.Error(),
			Data:    nil,
		}, nil
	}

	options := &internal_command.SentryOptions{
		Enable:             s.ProgramConf.LogReporterOpt.Enable,
		UserId:             s.ProgramConf.LogReporterOpt.UserId,
		TracesSampleRate:   s.ProgramConf.LogReporterOpt.TracesSampleRate,
		ProfilesSampleRate: s.ProgramConf.LogReporterOpt.ProfilesSampleRate,
		Level:              s.ProgramConf.LogReporterOpt.Level,
	}
	anyData, err := anypb.New(options)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("could not marshal sentry options: %v", err)
	}

	return &internal_command.RpcResponse{
		Code:    int32(exception.ErrSuccess.Code),
		Message: "Sentry options retrieved successfully",
		Data:    anyData,
	}, nil
}
func (s *internalRpcServer) RegisterInternalCommand(stream internal_command.ExecuteCommand_RegisterInternalCommandServer) error {

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		logger.Warn("Failed to get metadata from context")
		return status.Errorf(codes.Unauthenticated, "metadata not found")
	}
	clientId := ""
	if len(md.Get(constants.ClientIdKey)) > 0 {
		clientId = md.Get(constants.ClientIdKey)[0]
	}
	if clientId == "" {
		logger.Warn("clientId not found in metadata")
		return status.Errorf(codes.Unauthenticated, "client ID not found in context")
	}
	logger.Info("slave clientId: ", clientId)
	activeRpcClientMapLock.Lock()
	client, ok := activeRpcClientMap[clientId]
	activeRpcClientMapLock.Unlock()

	client.Connect()
	defer func() {
		client.Close()

	}()
	if !ok {
		logger.Warn("client not found,not registered")
		return status.Errorf(codes.Unauthenticated, "client ID not found,not registered")
	}

	goroutine.RecoverGO(func() {
		_, cancel := context.WithCancel(stream.Context())
		logger.Debug("start sending internal command thread")
		msg := &internal_command.RpcStream{}
		defer func() {
			logger.Debug("stop sending internal command thread")
			cancel()
			client.Close()
		}()
		for ic := range client.SendChan {
			switch ic.CommandType {
			case schema.UnknownCommandType:
				continue
			case schema.ExitCommandType:
				msg.Type = internal_command.StreamMessageType_ExitProcessRequest
				msg.Data = nil
				break
			case schema.LockPCCommandType:
				msg.Type = internal_command.StreamMessageType_LockPcRequest
				msg.Data = nil
				break
			}

			if err := stream.Send(msg); err != nil {
				logger.Errorf("Error sending internal command: %v", err)
				return
			}
			if msg.Type == internal_command.StreamMessageType_ExitProcessRequest {
				return
			}
		}
	})

	slaveBlockerLock.Lock()
	if slaveBlocker != nil {
		slaveBlocker.Cancel()
	}
	slaveBlockerLock.Unlock()
	for {
		logger.Debug("start recv data")
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case <-stream.Context().Done():
			logger.Info("this stream will be stopped,clientId: ", clientId)
			return nil
		default:
			break
		}
		data := req.GetData()
		if data == nil || req.Type == internal_command.StreamMessageType_Unknown {
			continue
		}

		if req.Type == internal_command.StreamMessageType_Response {
			continue
		}
	}
}
func (s *InternalMasterService) StartServer() error {
	port := 2095
	host := "127.0.0.1"
	addr := host + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)
	defer listener.Close()
	if err != nil {
		logger.Errorf("failed to listen: %v", err)
		return err
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(goroutine.UnaryServerInterceptor),
		grpc.StreamInterceptor(goroutine.StreamServerInterceptor),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     0,
			MaxConnectionAge:      0,
			MaxConnectionAgeGrace: 0,
			Time:                  1 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(
			keepalive.EnforcementPolicy{
				MinTime:             5 * time.Second,
				PermitWithoutStream: true,
			}),
	)

	_conf := utils.GetValueFromContext(s.ctx, constants.ConfKey, conf.NewDefaultConf())
	internal_command.RegisterBaseServer(grpcServer, &internalRpcServer{ProgramConf: _conf})
	internal_command.RegisterExecuteCommandServer(grpcServer, &internalRpcServer{ProgramConf: _conf})
	logger.Infof("rpc server listening on :%d", port)
	goroutine.RecoverGO(func() {
		err = grpcServer.Serve(listener)
		if err != nil {
			logger.Fatal("rpc server failed to serve: %v", err)
		}
	})
	select {
	case <-s.ctx.Done():
		s.StopServer()
		logger.Info("rpc server will be stopped")
		grpcServer.GracefulStop()
		logger.Info("rpc server stopped")
		return nil
	}
}

func (s *InternalMasterService) StopServer() error {
	s.stopOnce.Do(func() {
		err := s.StopAllSlave()
		if err != nil {
			logger.Error(err)

		}

	})

	return nil
}

func (s *InternalMasterService) StopAllSlave() error {
	logger.Debug("send exit command to all client")
	cmd := schema.InternalCommand{CommandType: schema.ExitCommandType, Data: nil}
	err := s.SendCommandAll(&cmd)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
func (s *InternalMasterService) SendCommand(client *rpcClient, cmd *schema.InternalCommand) error {

	if client.Status() != Connected {
		return fmt.Errorf("client %s is not connected", client.Username)
	}

	client.SendChan <- cmd
	return nil
}
func (s *InternalMasterService) SendCommandAll(cmd *schema.InternalCommand) error {
	logger.Debug("receive command    ")

	logger.Debug("lock..")

	activeRpcClientMapLock.RLock()
	for _, client := range activeRpcClientMap {

		goroutine.RecoverGO(
			func() {
				err := s.SendCommand(client, cmd)
				if err != nil {
					logger.Error(err)
				}
			})

	}
	activeRpcClientMapLock.RUnlock()
	logger.Debug("unlock..")

	logger.Debug("send command")
	return nil
}

var slaveBlocker *goroutine.Blocker
var slaveBlockerLock sync.Mutex

func (s *InternalMasterService) RunSlave() error {
	if s.slavePath == "" || s.slaveWorkDir == "" {
		return fmt.Errorf("slave path or work dir is empty")
	}

	excludedUsers := make(map[string]bool)
	usernameClientMapLock.RLock()
	for _, client := range usernameClientMap {
		_, usernameNoDomain := utils.SplitWindowsAccount(client.Username)
		excludedUsers[client.Username] = true
		excludedUsers[usernameNoDomain] = true
	}
	usernameClientMapLock.RUnlock()
	cnt, err := sys.RunProgramForAllUser(s.slavePath, "\""+s.slavePath+"\" --slave", s.slaveWorkDir, excludedUsers)
	if err != nil {
		return fmt.Errorf("cannot run slave program:%v", err)
	}
	if cnt > 0 {
		slaveBlockerLock.Lock()
		slaveBlocker = goroutine.NewBlocker(5 * time.Second)
		slaveBlockerLock.Unlock()
		slaveBlocker.Wait()
	}

	return nil
}
