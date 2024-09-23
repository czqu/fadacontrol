package internal_service

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"net"
	"strconv"
	"time"
)

type InternalMasterService struct {
	_done      chan bool
	_chanGroup *conf.ChanGroup
}

func NewInternalMasterService(_chanGroup *conf.ChanGroup) *InternalMasterService {
	return &InternalMasterService{_chanGroup: _chanGroup, _done: make(chan bool)}

}
func (s *InternalMasterService) Start() error {
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
	defer conn.Close()
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

	keepAlive := 30 * time.Second

	for {
		select {
		case <-s._done:
			return
		case msg := <-s._chanGroup.InternalCommandSend:
			packet := &schema.InternalDataPacket{DataLength: uint16(len(msg)), Data: msg, DataType: schema.JsonData}
			logger.Debug("receive command    ")
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
