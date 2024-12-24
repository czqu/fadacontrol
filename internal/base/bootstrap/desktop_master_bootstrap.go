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
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type DesktopMasterServiceBootstrap struct {
	_db      *data.Data
	discover *DiscoverBootstrap
	_http    *HttpBootstrap
	lo       *logger.Logger
	ctx      context.Context
	master   *internal_service.InternalMasterService

	done        chan interface{}
	rcb         *RemoteConnectBootstrap
	cp          *credential_provider_service.CredentialProviderService
	di          *DataInitBootstrap
	_co         *control_pc.ControlPCService
	pf          *ProfilingBootstrap
	_exitSignal *conf.ExitChanStruct
	startOnce   sync.Once
	stopOnce    sync.Once
}

func NewDesktopMasterServiceBootstrap(_exitSignal *conf.ExitChanStruct, pf *ProfilingBootstrap, _co *control_pc.ControlPCService, di *DataInitBootstrap, cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, master *internal_service.InternalMasterService, _context context.Context, _db *data.Data, lo *logger.Logger, d *DiscoverBootstrap, http_ *HttpBootstrap) *DesktopMasterServiceBootstrap {
	return &DesktopMasterServiceBootstrap{_exitSignal: _exitSignal, pf: pf, _co: _co, di: di, cp: cp, rcb: rcb, done: make(chan interface{}), master: master, ctx: _context, _db: _db, lo: lo, discover: d, _http: http_}
}
func (r *DesktopMasterServiceBootstrap) Start() {
	r.startOnce.Do(func() {
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
					logger.Info("starting slave program")

					path, err := os.Executable()
					if err != nil {
						logger.Error("cannot get executable path", err)
					}
					logger.Info("slave program path", path)
					dir := filepath.Dir(path)
					logger.Sync()
					err = sys.RunProgramForAllUser(path, "\""+path+"\" --slave", dir)
					if err != nil {
						logger.Error("cannot run slave program", err)
					}
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
		goroutine.RecoverGO(
			func() {
				<-r._exitSignal.ExitChan
				logger.Debug("stopping...")
				r.Stop()
			})

		r.Wait()
	})

}
func (r *DesktopMasterServiceBootstrap) Stop() {
	r.stopOnce.Do(
		func() {
			defer func() {
				logger.Info("app stopped")
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			logger.Debug("stopping root bootstrap")
			logger.Sync()
			stopCh := make(chan struct{}, 1)
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
			case <-stopCh:

			case <-ctx.Done():

			}

			r.done <- struct{}{}
		})

}
func (r *DesktopMasterServiceBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
