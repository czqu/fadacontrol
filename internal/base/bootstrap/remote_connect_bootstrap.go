package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/service/remote_service"
	"fadacontrol/pkg/goroutine"
	"gorm.io/gorm"
)

var RemoteConnectBootstrapInstance *RemoteConnectBootstrap

type RemoteConnectBootstrap struct {
	_conf *conf.Conf

	db *gorm.DB
	re *remote_service.RemoteService
}

func NewRemoteConnectBootstrap(_conf *conf.Conf, db *gorm.DB, re *remote_service.RemoteService) *RemoteConnectBootstrap {
	RemoteConnectBootstrapInstance = &RemoteConnectBootstrap{re: re, _conf: _conf, db: db}
	return RemoteConnectBootstrapInstance
}
func (r *RemoteConnectBootstrap) Start() error {

	goroutine.RecoverGO(func() {
		r.re.StartService()
	})

	return nil
}
func (r *RemoteConnectBootstrap) Stop() error {
	return r.re.StopService()

}
