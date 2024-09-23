package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/sys"
	"time"
)

type DesktopSlaveServiceBootstrap struct {
	_conf *conf.Conf
	lo    *logger.Logger
	slave *internal_service.InternalSlaveService
	done  chan interface{}
	di    *DataInitBootstrap
	_co   *control_pc.ControlPCService
}

func NewDesktopSlaveServiceBootstrap(_co *control_pc.ControlPCService, di *DataInitBootstrap, _conf *conf.Conf, lo *logger.Logger, slave *internal_service.InternalSlaveService) *DesktopSlaveServiceBootstrap {
	return &DesktopSlaveServiceBootstrap{_co: _co, di: di, _conf: _conf, lo: lo, slave: slave, done: make(chan interface{})}
}

func (r *DesktopSlaveServiceBootstrap) Start() {

	r.di.Start()
	r.lo.InitLog()

	r._co.RunPowerSavingMode()
	ret := sys.SetPowerSavingMode(true)
	if ret == true {
		logger.Debug("set power saving mode")
	}
	r.slave.Start()
	r.Wait()
	return

}
func (r *DesktopSlaveServiceBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
func (r *DesktopSlaveServiceBootstrap) Stop() {

	logger.Debug("stopping root bootstrap")
	go func() {
		r.slave.Stop()
		time.Sleep(5 * time.Second)
	}()

	r.done <- struct{}{}
}
