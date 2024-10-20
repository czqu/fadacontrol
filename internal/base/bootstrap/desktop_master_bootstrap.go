package bootstrap

import (
	"context"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type DesktopMasterServiceBootstrap struct {
	_db      *data.Data
	ble      *BleUnlockBootstrap
	discover *DiscoverBootstrap
	_http    *HttpBootstrap
	legacy_  *LegacyBootstrap
	lo       *logger.Logger
	_conf    *conf.Conf
	master   *internal_service.InternalMasterService

	done        chan interface{}
	rcb         *RemoteConnectBootstrap
	cp          *credential_provider_service.CredentialProviderService
	di          *DataInitBootstrap
	_co         *control_pc.ControlPCService
	pf          *ProfilingBootstrap
	_exitSignal *conf.ExitChanStruct
}

func NewDesktopMasterServiceBootstrap(_exitSignal *conf.ExitChanStruct, pf *ProfilingBootstrap, _co *control_pc.ControlPCService, di *DataInitBootstrap, cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, master *internal_service.InternalMasterService, _conf *conf.Conf, _db *data.Data, lo *logger.Logger, ble *BleUnlockBootstrap, d *DiscoverBootstrap, http_ *HttpBootstrap, legacy *LegacyBootstrap) *DesktopMasterServiceBootstrap {
	return &DesktopMasterServiceBootstrap{_exitSignal: _exitSignal, pf: pf, _co: _co, di: di, cp: cp, rcb: rcb, done: make(chan interface{}), master: master, _conf: _conf, _db: _db, lo: lo, ble: ble, discover: d, _http: http_, legacy_: legacy}
}
func (r *DesktopMasterServiceBootstrap) Start() {
	goroutine.RecoverGO(func() {
		r.pf.Start()
	})
	r.di.Start()

	r._co.RunPowerSavingMode()
	r._http.Start()
	if r._conf.StartMode == conf.ServiceMode {
		logger.Info("service mode")
		r.master.Start()
	}

	r.legacy_.Start()
	r.ble.Start()
	r.rcb.Start()
	goroutine.RecoverGO(func() {
		r.discover.Start()
	})
	goroutine.RecoverGO(func() {
		r.cp.Connect()

	})
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
}
func (r *DesktopMasterServiceBootstrap) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	logger.Debug("stopping root bootstrap")
	logger.Sync()
	stopCh := make(chan struct{}, 1)
	goroutine.RecoverGO(
		func() {
			r.pf.Stop()
			r.ble.Stop()
			r.discover.Stop()
			r._http.Stop()
			r.legacy_.Stop()
			r.rcb.Stop()
			r.master.Stop()
		})
	select {
	case <-stopCh:

	case <-ctx.Done():

	}

	r.done <- struct{}{}
}
func (r *DesktopMasterServiceBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
