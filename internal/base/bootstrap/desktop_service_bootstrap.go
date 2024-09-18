package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/pkg/sys"
)

type DesktopServiceBootstrap struct {
	_db      *data.Data
	ble      *BleUnlockBootstrap
	discover *DiscoverBootstrap
	http_    *HttpBootstrap
	legacy_  *LegacyBootstrap
	lo       *logger.Logger
	_conf    *conf.Conf
	_daemon  *DaemonConnectBootstrap
	it       *InternalServiceBootstrap
	done     chan interface{}
	rcb      *RemoteConnectBootstrap
	cp       *credential_provider_service.CredentialProviderService
}

func NewRootBootstrap(cp *credential_provider_service.CredentialProviderService, rcb *RemoteConnectBootstrap, it *InternalServiceBootstrap, _daemon *DaemonConnectBootstrap, _conf *conf.Conf, _db *data.Data, lo *logger.Logger, ble *BleUnlockBootstrap, d *DiscoverBootstrap, http_ *HttpBootstrap, legacy *LegacyBootstrap) *DesktopServiceBootstrap {
	return &DesktopServiceBootstrap{cp: cp, rcb: rcb, done: make(chan interface{}), it: it, _daemon: _daemon, _conf: _conf, _db: _db, lo: lo, ble: ble, discover: d, http_: http_, legacy_: legacy}
}
func (r *DesktopServiceBootstrap) Start() {

	r.lo.InitLog()
	ret := sys.SetPowerSavingMode(true)
	if ret == true {
		logger.Debug("set power saving mode")
	}
	go r.discover.Start()
	r._daemon.Start()
	r.http_.Start()
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
		r.http_.Stop()
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
