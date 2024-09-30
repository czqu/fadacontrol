package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"os"
	"os/signal"
	"syscall"
)

type DesktopSlaveServiceBootstrap struct {
	_conf *conf.Conf
	lo    *logger.Logger
	slave *internal_service.InternalSlaveService
	done  chan interface{}
	di    *DataInitBootstrap
	_co   *control_pc.ControlPCService
	pf    *ProfilingBootstrap
}

func NewDesktopSlaveServiceBootstrap(pf *ProfilingBootstrap, _co *control_pc.ControlPCService, di *DataInitBootstrap, _conf *conf.Conf, lo *logger.Logger, slave *internal_service.InternalSlaveService) *DesktopSlaveServiceBootstrap {
	return &DesktopSlaveServiceBootstrap{pf: pf, _co: _co, di: di, _conf: _conf, lo: lo, slave: slave, done: make(chan interface{})}
}

func (r *DesktopSlaveServiceBootstrap) Start() {

	goroutine.RecoverGO(func() {
		r.pf.Start()
	})
	r.di.Start()

	r._co.RunPowerSavingMode()
	ret := sys.SetPowerSavingMode(true)
	if ret == true {
		logger.Debug("set power saving mode")
	}
	r.slave.Start()
	goroutine.RecoverGO(
		func() {
			sChan := make(chan os.Signal, 1)
			signal.Notify(sChan,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			<-sChan
			logger.Debug("stopping...")
			r.Stop()
		})
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
	r.pf.Stop()
	logger.Sync()

	logger.Debug("stopping root bootstrap")

	r.slave.Stop()

	logger.Sync()
	r.done <- struct{}{}
}
