package control_pc

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/sys"
	"gorm.io/gorm"
)

type ControlPCService struct {
	_db          *gorm.DB
	commandGroup *conf.ChanGroup
}

func NewControlPCService(_db *gorm.DB, commandGroup *conf.ChanGroup) *ControlPCService {
	return &ControlPCService{_db: _db, commandGroup: commandGroup}

}

func (control *ControlPCService) Standby() *exception.Exception {
	return sys.Standby()

}
func (control *ControlPCService) Shutdown(tpe sys.ShutdownType) *exception.Exception {
	return sys.Shutdown(tpe)
}
func (control *ControlPCService) LockWindows(useAgent bool) *exception.Exception {

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

func (control *ControlPCService) SetPowerSavingMode(enable bool) error {

	var config entity.SysConfig
	if err := control._db.First(&config).Error; err != nil {
		return err
	}
	config.PowerSavingMode = enable

	if err := control._db.Save(&config).Error; err != nil {
		return err
	}

	ret := false
	if enable {
		ret = sys.SetPowerSavingMode(true)

	} else {
		ret = sys.SetPowerSavingMode(false)
	}
	if ret == false {
		return exception.ErrSystemSetPowerSaveModeError
	}

	return nil
}
func (control *ControlPCService) RunPowerSavingMode() *exception.Exception {

	var config entity.SysConfig
	if err := control._db.First(&config).Error; err != nil {
		return exception.ErrSystemUnknownException
	}
	enable := config.PowerSavingMode

	if !enable {
		return nil
	}

	ret := sys.SetPowerSavingMode(true)

	if ret == false {
		return exception.ErrSystemUnknownException
	}

	return nil
}
