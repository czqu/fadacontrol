package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
)

type DesktopServiceBootstrap struct {
	_db      *data.Data
	ble      *BleUnlockBootstrap
	discover *DiscoverBootstrap
	_http    *HttpBootstrap
	legacy_  *LegacyBootstrap
	lo       *logger.Logger
	_conf    *conf.Conf
	_daemon  *DaemonConnectBootstrap

	done chan interface{}
	rcb  *RemoteConnectBootstrap
	cp   *credential_provider_service.CredentialProviderService
	di   *DataInitBootstrap
	_co  *control_pc.ControlPCService
}

func NewDesktopServiceBootstrap(_co *control_pc.ControlPCService, di *DataInitBootstrap, cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, _daemon *DaemonConnectBootstrap, _conf *conf.Conf, _db *data.Data, lo *logger.Logger, ble *BleUnlockBootstrap, d *DiscoverBootstrap, http_ *HttpBootstrap, legacy *LegacyBootstrap) *DesktopServiceBootstrap {
	return &DesktopServiceBootstrap{_co: _co, di: di, cp: cp, rcb: rcb, done: make(chan interface{}), _daemon: _daemon, _conf: _conf, _db: _db, lo: lo, ble: ble, discover: d, _http: http_, legacy_: legacy}
}
func (r *DesktopServiceBootstrap) Start() {
	r.di.Start()
	r.lo.InitLog()

	r._co.RunPowerSavingMode()
	go r.discover.Start()
	r._daemon.Start()
	r._http.Start()
	r.legacy_.Start()
	r.ble.Start()
	r.rcb.Start()
	go r.cp.Connect()
	//go func() {
	//	signalCh := make(chan os.Signal, 1)
	//	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	//	logger.Debug("received signal")
	//	<-signalCh
	//	r.Stop()
	//}()
	r.Wait()
}
func (r *DesktopServiceBootstrap) Stop() {

	logger.Debug("stopping root bootstrap")
	go func() {
		r.ble.Stop()
		r.discover.Stop()
		r._http.Stop()
		r.legacy_.Stop()
		r.rcb.Stop()
		r._daemon.Stop()
	}()

	r.done <- struct{}{}
}
func (r *DesktopServiceBootstrap) Wait() {
	select {
	case <-r.done:
		return
	}
}
