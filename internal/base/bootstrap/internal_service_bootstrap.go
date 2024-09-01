package bootstrap

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/common_service"

	"net"
	"strconv"
)

type InternalServiceBootstrap struct {
	it   *common_service.InternalService
	sign chan interface{}
}

func NewInternalServiceBootstrap(it *common_service.InternalService) *InternalServiceBootstrap {
	return &InternalServiceBootstrap{it: it, sign: make(chan interface{})}
}

func (s *InternalServiceBootstrap) Start() error {
	go s.StartServer()
	return nil

}
func (s *InternalServiceBootstrap) Stop() error {
	return s.StopServer()
}

func (s *InternalServiceBootstrap) StartServer() error {
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
		<-s.sign
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
		go s.it.Handler(conn)
	}

}

func (s *InternalServiceBootstrap) StopServer() error {
	s.sign <- 0
	return nil
}
