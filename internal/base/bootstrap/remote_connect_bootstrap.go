package bootstrap

import (
	"context"
	"fadacontrol/internal/service/remote_service"
	"fadacontrol/pkg/goroutine"
	"gorm.io/gorm"
)

var RemoteConnectBootstrapInstance *RemoteConnectBootstrap

type RemoteConnectBootstrap struct {
	ctx context.Context

	db *gorm.DB
	re *remote_service.RemoteService
}

func NewRemoteConnectBootstrap(ctx context.Context, db *gorm.DB, re *remote_service.RemoteService) *RemoteConnectBootstrap {
	RemoteConnectBootstrapInstance = &RemoteConnectBootstrap{re: re, ctx: ctx, db: db}
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
func (r *RemoteConnectBootstrap) Restart() error {
	return r.re.RestartService()
}
