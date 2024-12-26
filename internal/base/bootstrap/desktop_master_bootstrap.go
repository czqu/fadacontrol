package bootstrap

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type DesktopMasterServiceBootstrap struct {
	_db       *data.Data
	discover  *DiscoverBootstrap
	_http     *HttpBootstrap
	lo        *logger.Logger
	ctx       context.Context
	master    *internal_service.InternalMasterService
	rcb       *RemoteConnectBootstrap
	cp        *credential_provider_service.CredentialProviderService
	di        *DataInitBootstrap
	_co       *control_pc.ControlPCService
	pf        *ProfilingBootstrap
	startOnce sync.Once
	stopOnce  sync.Once
	cancel    context.CancelFunc
}

func NewDesktopMasterServiceBootstrap(pf *ProfilingBootstrap, _co *control_pc.ControlPCService, di *DataInitBootstrap, cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, master *internal_service.InternalMasterService, _context context.Context, _db *data.Data, lo *logger.Logger, d *DiscoverBootstrap, http_ *HttpBootstrap) *DesktopMasterServiceBootstrap {
	return &DesktopMasterServiceBootstrap{pf: pf, _co: _co, di: di, cp: cp, rcb: rcb, master: master, ctx: _context, _db: _db, lo: lo, discover: d, _http: http_}
}
func (r *DesktopMasterServiceBootstrap) Start() {
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
		logger.Info("starting app")
		goroutine.RecoverGO(func() {
			r.pf.Start()
		})

		r.di.Start()
		if !conf.ResetPassword {
			r._co.RunPowerSavingMode()
			r._http.Start()

			if _conf.StartMode == conf.ServiceMode {
				logger.Info("service mode")
				r.master.Start()
			}

			r.rcb.Start()
			goroutine.RecoverGO(func() {
				r.discover.Start()
			})
			goroutine.RecoverGO(func() {
				r.cp.Start()

			})
			utils.AddNetworkChangeCallback(func() {
				logger.Warn("network change")
				r.rcb.Restart()
				r.discover.Restart()
			})

			goroutine.RecoverGO(
				func() {
					sys.RunSlave()
				})

		}

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
		}

	})

}
func (r *DesktopMasterServiceBootstrap) Stop() {
	r.stopOnce.Do(
		func() {
			defer func() {
				logger.Info("app stopped")
			}()
			r.cancel()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			logger.Debug("stopping root bootstrap")
			logger.Sync()
			goroutine.RecoverGO(
				func() {
					r.pf.Stop()
					if !conf.ResetPassword {
						r.master.Stop()
						r.discover.Stop()
						r._http.Stop()
						r.rcb.Stop()

					}

				})
			select {

			case <-ctx.Done():
				return

			}

		})

}
