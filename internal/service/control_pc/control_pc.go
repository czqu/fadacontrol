package control_pc

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/sys"
)

type ControlPCService struct {
	commandGroup *conf.ChanGroup
}

func NewControlPCService(commandGroup *conf.ChanGroup) *ControlPCService {
	return &ControlPCService{commandGroup: commandGroup}

}

func (control *ControlPCService) Standby() *exception.Exception {
	return sys.Standby()

}
func (control *ControlPCService) Shutdown() *exception.Exception {
	return sys.Shutdown()
}
func (control *ControlPCService) LockWindows(useAgent bool) *exception.Exception {
	//

	if !useAgent {
		return sys.LockWindows()
	}
	cmd := &schema.InternalCommand{CommandType: schema.LockPC, Data: nil}
	cmdStr, err := json.Marshal(cmd)
	if err != nil {
		return exception.ErrSystemUnknownException
	}
	go func() {
		control.commandGroup.InternalCommandSend <- cmdStr
	}()

	return exception.ErrSuccess

}
