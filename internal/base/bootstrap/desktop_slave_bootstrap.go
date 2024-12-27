package bootstrap

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/internal_slave_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type DesktopSlaveServiceBootstrap struct {
	ctx   context.Context
	lo    *logger.Logger
	slave *internal_slave_service.InternalSlaveService

	_co       *control_pc.ControlPCService
	pf        *ProfilingBootstrap
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
}

func NewDesktopSlaveServiceBootstrap(ctx context.Context, pf *ProfilingBootstrap, _co *control_pc.ControlPCService, lo *logger.Logger, slave *internal_slave_service.InternalSlaveService) *DesktopSlaveServiceBootstrap {
	return &DesktopSlaveServiceBootstrap{ctx: ctx, pf: pf, _co: _co, lo: lo, slave: slave}
}

func (r *DesktopSlaveServiceBootstrap) Start() {
	r.startOnce.Do(func() {
		cancelFunc := r.ctx.Value(constants.CancelFuncKey)
		if cancelFunc == nil {
			if _, ok := cancelFunc.(context.CancelFunc); !ok {
				logger.Error("cancel func not found")
				return
			}

		}
		r.cancel = cancelFunc.(context.CancelFunc)
		_conf := utils.GetValueFromContext(r.ctx, constants.ConfKey, conf.NewDefaultConf())
		if _conf.StartMode == conf.UnknownMode {
			logger.Error("unknown start mode")
			return
		}
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
		select {
		case <-r.ctx.Done():
			logger.Debug("stopping...")
			r.Stop()
			return
		}

	})

}

func (r *DesktopSlaveServiceBootstrap) Stop() {
	r.stopOnce.Do(func() {
		logger.Sync()
		r.cancel()

		logger.Debug("stopping root bootstrap")

		r.pf.Stop()
		logger.Sync()

	})

}
