package internal_service

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/schema/custom_command_schema"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/custom_command_service"
	"github.com/mitchellh/mapstructure"
	"net"
	"strconv"
	"time"
)

type InternalSlaveService struct {
	_done chan bool

	conf *conf.Conf
	cu   *custom_command_service.CustomCommandService
	co   *control_pc.ControlPCService
}

func NewInternalSlaveService(cu *custom_command_service.CustomCommandService, co *control_pc.ControlPCService, _chanGroup *conf.ChanGroup, conf *conf.Conf) *InternalSlaveService {
	return &InternalSlaveService{cu: cu, co: co, conf: conf, _done: make(chan bool)}
}
func (s *InternalSlaveService) Start() {
	port := 2095
	host := "127.0.0.1"
	addr := host + ":" + strconv.Itoa(port)
	go s.connectToServer(addr)
}
func (s *InternalSlaveService) Stop() {
	s._done <- true
}

const (
	initialBackoff = 1 * time.Second
	maxBackoff     = 16 * time.Second
)

func (s *InternalSlaveService) connectToServer(addr string) {
	backoff := initialBackoff

	for {
		conn, err := net.Dial("tcp", addr)

		select {
		case <-s._done:
			return
		default:
			break
		}
		if err != nil || conn == nil {
			logger.Debugf("Error connecting to server: %v\n", err)

			// Wait for the backoff time and try again
			time.Sleep(backoff)

			// Increase the backoff time until maxBackoff is reached
			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					return
				}
			}
			continue
		}

		tcpConn := conn.(*net.TCPConn)
		err = tcpConn.SetKeepAlive(true)
		if err != nil {
			logger.Warn("Error setting keep-alive:", err)
		}
		// Reset the backoff time after a successful connection
		backoff = initialBackoff
		for {
			packet := &schema.InternalDataPacket{}
			err := packet.Unpack(conn)
			if err != nil {
				logger.Warnf("Error unpacking packet: %v", err)
				break
			}
			logger.Debug("recv a packet from server")
			if packet.DataType == schema.JsonData {
				err := s.JsonDataHandler(conn, packet)
				if err != nil {
					logger.Warnf("JsonDataHandler err: %v", err)
					break
				}
			}
			if packet.DataType == schema.BinaryData {
				err := s.BinaryDataHandler(conn, packet)
				if err != nil {
					logger.Warnf("BinaryDataHandler err: %v", err)
					break
				}
			}

		}
		conn.Close()

	}
}
func (s *InternalSlaveService) JsonDataHandler(conn net.Conn, packet *schema.InternalDataPacket) error {
	if packet.DataType != schema.JsonData {
		return nil
	}
	j := packet.Data
	cmd := schema.InternalCommand{}
	err := json.Unmarshal(j, &cmd)
	if err != nil {
		return err
	}
	if cmd.CommandType == schema.LockPC {
		logger.Debug("Lock PC")
		err := s.co.LockWindows(false)
		return err
	}
	if cmd.CommandType == schema.Hello {
		logger.Debug("Hello from server")
		return nil
	}
	if cmd.CommandType == schema.KeepLive {
		logger.Debug("Keep Live")
		return nil
	}
	if cmd.CommandType == schema.CustomCommand {
		logger.Debug("Custom Command")
		data := cmd.Data
		ccs := custom_command_schema.Command{}
		err := mapstructure.Decode(data, &ccs)
		if err != nil {
			return err
		}
		stdout := custom_command_schema.NewCustomWriter()
		stderr := custom_command_schema.NewCustomWriter()
		err = s.cu.ExecuteCommand(ccs, stdout, stderr)
		if err != nil {
			return err
		}

	}
	return nil
}

func (s *InternalSlaveService) BinaryDataHandler(conn net.Conn, packet *schema.InternalDataPacket) error {
	return nil
}