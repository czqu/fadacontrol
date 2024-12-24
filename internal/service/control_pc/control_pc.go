package control_pc

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/sys"
)

type ControlPCService struct {
	_cmdSender func(cmd *schema.InternalCommand) error
}

func NewControlPCService() *ControlPCService {
	return &ControlPCService{}

}

func (control *ControlPCService) Standby() *exception.Exception {
	return sys.Standby()

}
func (control *ControlPCService) Shutdown(tpe sys.ShutdownType) *exception.Exception {
	return sys.Shutdown(tpe)
}
func (control *ControlPCService) SetCommandSender(f func(cmd *schema.InternalCommand) error) {
	control._cmdSender = f

}
func (control *ControlPCService) LockWindows(useAgent bool) *exception.Exception {

	logger.Debug("lock windows")
	if !useAgent {
		return sys.LockWindows()
	}
	if control._cmdSender == nil {
		return exception.ErrSystemUnknownException
	}

	cmd := &schema.InternalCommand{CommandType: schema.LockPCCommandType, Data: nil}
	err := control._cmdSender(cmd)
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
