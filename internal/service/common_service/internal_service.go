package common_service

import (
	"encoding/json"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/schema/custom_command_schema"
	"fadacontrol/internal/service/custom_command_service"
	"fadacontrol/pkg/sys"
	"github.com/mitchellh/mapstructure"
	"net"
)

type InternalService struct {
	cu *custom_command_service.CustomCommandService
}

func NewInternalService(cu *custom_command_service.CustomCommandService) *InternalService {
	return &InternalService{cu: cu}
}

func (s *InternalService) Handler(conn net.Conn) {
	defer conn.Close()
	tcpConn := conn.(*net.TCPConn)
	err := tcpConn.SetKeepAlive(true)
	if err != nil {
		logger.Warn("Error setting keep-alive:", err)
	}
	for {
		packet := &schema.InternalDataPacket{}
		err := packet.Unpack(conn)
		if err != nil {
			logger.Warnf("Error unpacking packet: %v", err)
			break
		}
		if packet.DataType == schema.JsonData {
			err := s.JsonDataHandler(packet)
			if err != nil {
				logger.Warnf("JsonDataHandler err: %v", err)
			}
		}
		if packet.DataType == schema.BinaryData {
			err := s.BinaryDataHandler(packet)
			if err != nil {
				logger.Warnf("BinaryDataHandler err: %v", err)
			}
		}

	}

}
func (s *InternalService) JsonDataHandler(packet *schema.InternalDataPacket) error {
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
		err := sys.LockWindows()
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

func (s *InternalService) BinaryDataHandler(packet *schema.InternalDataPacket) error {
	return nil
}
