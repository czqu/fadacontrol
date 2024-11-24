package control_pc

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/sys"
	"gorm.io/gorm"
)

type ControlPCService struct {
	_db *gorm.DB

	_cmdSender func(cmd *schema.InternalCommand) error
}

func NewControlPCService(_db *gorm.DB) *ControlPCService {
	return &ControlPCService{_db: _db}

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

	cmd := &schema.InternalCommand{CommandType: schema.LockPC, Data: nil}
	err := control._cmdSender(cmd)
	if err != nil {
		logger.Warn("lock windows marshal failed")
		return exception.ErrSystemUnknownException
	}

	logger.Debug("lock windows succ")
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
		config.PowerSavingMode = false
		control._db.Save(&config)
		return exception.ErrSystemSetPowerSaveModeError
	}
	return nil
}
func (control *ControlPCService) GetPowerSavingModeStatus() (*schema.PowerSavingModeInfo, error) {
	var config entity.SysConfig
	if err := control._db.First(&config).Error; err != nil {
		return nil, err
	}
	ret := &schema.PowerSavingModeInfo{
		PowerSavingMode: config.PowerSavingMode,
	}
	return ret, nil

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
		config.PowerSavingMode = false
		control._db.Save(&config)
		return exception.ErrSystemUnknownException
	}

	return nil
}
