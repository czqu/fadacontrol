package internal_service

import (
	"encoding/json"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/control_pc"
	"net"
	"strconv"
	"sync"
	"time"
)

type InternalMasterService struct {
	_done       chan bool
	_activeConn map[string]net.Conn
	cp          *control_pc.ControlPCService
	_lock       sync.Mutex
}

func NewInternalMasterService(cp *control_pc.ControlPCService) *InternalMasterService {
	return &InternalMasterService{_done: make(chan bool), cp: cp, _activeConn: make(map[string]net.Conn)}

}
func (s *InternalMasterService) Start() error {
	s.cp.SetCommandSender(s.SendCommand)
	go s.StartServer()
	return nil

}
func (s *InternalMasterService) Stop() error {
	return s.StopServer()
}

func (s *InternalMasterService) StartServer() error {
	port := 2095
	host := "127.0.0.1"
	addr := host + ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		logger.Errorf("failed to listen: %v", err)
		return err
	}
	defer listener.Close()

	logger.Infof("Starting Socket server on %s:%d", host, port)
	done := make(chan struct{})
	go func() {
		<-s._done
		listener.Close()
		done <- struct{}{}
		close(done)
		logger.Info("socket close at", port)
	}()
	for {

		conn, err := listener.Accept()

		logger.Debug("recv a connection from", conn.RemoteAddr())

		select {
		case <-done:
			logger.Info("Server stopped, exiting accept loop.")
			return nil
		default:
			break
		}
		if err != nil {
			logger.Errorf("failed to accept connection: %v", err)
			continue
		}
		go s.Handler(conn)
	}

}

func (s *InternalMasterService) StopServer() error {
	s._done <- true
	return nil
}
func (s *InternalMasterService) Handler(conn net.Conn) {
	s._lock.Lock()
	s._activeConn[conn.RemoteAddr().String()] = conn
	s._lock.Unlock()

	defer func() {
		s._lock.Lock()
		delete(s._activeConn, conn.RemoteAddr().String())
		s._lock.Unlock()
		conn.Close()

	}()

	tcpConn := conn.(*net.TCPConn)
	err := tcpConn.SetKeepAlive(true)
	if err != nil {
		logger.Warn("Error setting keep-alive:", err)
	}

	cmd := schema.InternalCommand{CommandType: schema.Hello, Data: nil}
	msg, err := json.Marshal(cmd)
	if err != nil {
		logger.Error(err)
		conn.Close()
		return
	}
	packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
	data, err := packet.Pack()
	if err != nil {
		logger.Error(err)
		return
	}
	conn.Write(data)

	keepAlive := 60 * time.Second

	for {
		select {
		case <-s._done:
			return

		case <-time.After(keepAlive):
			cmd := schema.InternalCommand{CommandType: schema.KeepLive, Data: nil}
			logger.Debug("send keep live to server")
			msg, err := json.Marshal(cmd)
			packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
			data, err := packet.Pack()
			if err != nil {
				logger.Error(err)
				break
			}
			_, err = conn.Write(data)
			if err != nil {
				logger.Debug(err)
				return
			}
		}
	}
}
func (s *InternalMasterService) SendCommand(cmd *schema.InternalCommand) error {
	logger.Debug("receive command    ")
	msg, err := json.Marshal(cmd)
	if err != nil {
		logger.Error(err)
		return err
	}
	packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}

	data, err := packet.Pack()
	if err != nil {
		logger.Error(err)
		return err
	}

	s._lock.Lock()

	for _, conn := range s._activeConn {
		_, err = conn.Write(data)
		if err != nil {
			logger.Debug(err)
			return err
		}
	}
	s._lock.Unlock()
	logger.Debug("send command")
	return nil
}
