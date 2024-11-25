package control_pc

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/internal/service/internal_master_service"
	"fadacontrol/pkg/sys"
	"fmt"
	"net"
)

type ControlPCService struct {
	_im *internal_master_service.InternalMasterService
}

func NewControlPCService(_im *internal_master_service.InternalMasterService) *ControlPCService {
	return &ControlPCService{_im: _im}

}

func (control *ControlPCService) Standby() *exception.Exception {
	return sys.Standby()

}
func (control *ControlPCService) Shutdown(tpe sys.ShutdownType) *exception.Exception {
	return sys.Shutdown(tpe)
}

func (control *ControlPCService) LockWindows(useAgent bool) *exception.Exception {

	logger.Debug("lock windows")
	if !useAgent {
		return sys.LockWindows()
	}
	ex := control._im.RunSlave()
	if ex != nil {
		logger.Warn("run slave failed")
		return exception.ErrSystemUnknownException.SetMsg(ex.Error())
	}

	cmd := &schema.InternalCommand{CommandType: schema.LockPCCommandType, Data: nil}
	err := control._im.SendCommandAll(cmd)
	if err != nil {
		logger.Warn("lock windows marshal failed")
		return exception.ErrSystemUnknownException
	}

	logger.Debug("lock windows succ")
	return exception.ErrSuccess

}
func (control *ControlPCService) SetPowerSavingMode(enable bool) error {

	ret := false
	if enable {
		ret = sys.SetPowerSavingMode(true)

	} else {
		ret = sys.SetPowerSavingMode(false)
	}
	logger.Debug("set power saving mode: ", ret)
	return nil
}

func (control *ControlPCService) RunPowerSavingMode() *exception.Exception {

	sys.SetPowerSavingMode(true)

	return nil
}
func (control *ControlPCService) PowerOnOtherDevices(macAddr []byte) error {

	if len(macAddr) != 6 {
		return fmt.Errorf("invalid MAC address length: %d", len(macAddr))
	}

	magicPacket := make([]byte, 6+16*len(macAddr))

	for i := 0; i < 6; i++ {
		magicPacket[i] = 0xFF
	}

	for i := 6; i < len(magicPacket); i += len(macAddr) {
		copy(magicPacket[i:], macAddr)
	}

	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	err = conn.(*net.UDPConn).SetWriteBuffer(len(magicPacket))
	if err != nil {
		return fmt.Errorf("failed to set write buffer: %v", err)
	}

	_, err = conn.Write(magicPacket)
	if err != nil {
		return fmt.Errorf("failed to send magic packet: %v", err)
	}

	return nil
}
