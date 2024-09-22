package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/pkg/sys"
	"time"
)

type DesktopDaemonBootstrap struct {
	_conf *conf.Conf
	lo    *logger.Logger
	it    *InternalServiceBootstrap
	done  chan interface{}
	di    *DataInitBootstrap
	_co   *control_pc.ControlPCService
}

func NewDesktopDaemonBootstrap(_co *control_pc.ControlPCService, di *DataInitBootstrap, _conf *conf.Conf, lo *logger.Logger, it *InternalServiceBootstrap) *DesktopDaemonBootstrap {
	return &DesktopDaemonBootstrap{_co: _co, di: di, _conf: _conf, lo: lo, it: it, done: make(chan interface{})}
}

func (r *DesktopDaemonBootstrap) Start() {

	r.di.Start()
	r.lo.InitLog()

	r._co.RunPowerSavingMode()
	ret := sys.SetPowerSavingMode(true)
	if ret == true {
		logger.Debug("set power saving mode")
	}
	r.it.Start()
	r.Wait()
	return

}
func (r *DesktopDaemonBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
func (r *DesktopDaemonBootstrap) Stop() {

	logger.Debug("stopping root bootstrap")
	go func() {
		r.it.Stop()
		time.Sleep(5 * time.Second)
	}()

	r.done <- struct{}{}
}
