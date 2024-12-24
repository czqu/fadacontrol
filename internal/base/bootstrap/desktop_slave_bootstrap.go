package bootstrap

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type DesktopSlaveServiceBootstrap struct {
	_conf       *conf.Conf
	lo          *logger.Logger
	slave       *internal_service.InternalSlaveService
	done        chan interface{}
	di          *DataInitBootstrap
	_co         *control_pc.ControlPCService
	pf          *ProfilingBootstrap
	_exitSignal *conf.ExitChanStruct
}

func NewDesktopSlaveServiceBootstrap(_exitSignal *conf.ExitChanStruct, pf *ProfilingBootstrap, _co *control_pc.ControlPCService, di *DataInitBootstrap, _conf *conf.Conf, lo *logger.Logger, slave *internal_service.InternalSlaveService) *DesktopSlaveServiceBootstrap {
	return &DesktopSlaveServiceBootstrap{_exitSignal: _exitSignal, pf: pf, _co: _co, di: di, _conf: _conf, lo: lo, slave: slave, done: make(chan interface{})}
}

func (r *DesktopSlaveServiceBootstrap) Start() {

	goroutine.RecoverGO(func() {
		r.pf.Start()
	})

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
	goroutine.RecoverGO(
		func() {
			<-r._exitSignal.ExitChan
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	stopCh := make(chan struct{}, 1)
	logger.Debug("stopping root bootstrap")
	goroutine.RecoverGO(
		func() {
			r.pf.Stop()
			logger.Sync()
			r.slave.Stop()
		})

	select {
	case <-stopCh:

	case <-ctx.Done():

	}
	logger.Sync()
	r.done <- struct{}{}
	logger.Info("slave exit")
}
