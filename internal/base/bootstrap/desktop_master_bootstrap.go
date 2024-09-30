package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/internal/service/internal_service"
	"fadacontrol/pkg/goroutine"
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

	done chan interface{}
	rcb  *RemoteConnectBootstrap
	cp   *credential_provider_service.CredentialProviderService
	di   *DataInitBootstrap
	_co  *control_pc.ControlPCService
}

func NewDesktopMasterServiceBootstrap(_co *control_pc.ControlPCService, di *DataInitBootstrap, cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, master *internal_service.InternalMasterService, _conf *conf.Conf, _db *data.Data, lo *logger.Logger, ble *BleUnlockBootstrap, d *DiscoverBootstrap, http_ *HttpBootstrap, legacy *LegacyBootstrap) *DesktopMasterServiceBootstrap {
	return &DesktopMasterServiceBootstrap{_co: _co, di: di, cp: cp, rcb: rcb, done: make(chan interface{}), master: master, _conf: _conf, _db: _db, lo: lo, ble: ble, discover: d, _http: http_, legacy_: legacy}
}
func (r *DesktopMasterServiceBootstrap) Start() {
	r.di.Start()

	r._co.RunPowerSavingMode()

	r.master.Start()
	r._http.Start()
	r.legacy_.Start()
	r.ble.Start()
	r.rcb.Start()
	goroutine.RecoverGO(func() {
		r.discover.Start()
	})
	goroutine.RecoverGO(func() {
		r.cp.Connect()

	})

	r.Wait()
}
func (r *DesktopMasterServiceBootstrap) Stop() {

	logger.Debug("stopping root bootstrap")
	logger.Sync()
	goroutine.RecoverGO(func() {
		r.ble.Stop()
		r.discover.Stop()
		r._http.Stop()
		r.legacy_.Stop()
		r.rcb.Stop()
		r.master.Stop()
	})

	r.done <- struct{}{}
}
func (r *DesktopMasterServiceBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
