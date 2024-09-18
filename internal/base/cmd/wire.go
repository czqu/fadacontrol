//go:build wireinject
// +build wireinject

package cmd

import (
	"fadacontrol/internal/base/bootstrap"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/controller"
	"fadacontrol/internal/router"
	"fadacontrol/internal/service/common_service"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/internal/service/custom_command_service"
	"fadacontrol/internal/service/remote_service"
	"fadacontrol/internal/service/unlock"
	"github.com/google/wire"
)

func initDesktopServiceApplication(_conf *conf.Conf, db *conf.DatabaseConf) (*DesktopServiceApp, error) {
	wire.Build(NewDesktopServiceApp, bootstrap.NewBleUnlockBootstrap, bootstrap.NewHttpBootstrap, bootstrap.NewDiscoverBootstrap,
		bootstrap.NewLegacyBootstrap, logger.NewLogger, unlock.NewUnLockService, data.NewDB, router.NewCommonRouter, router.NewAdminRouter, bootstrap.NewRootBootstrap,
		control_pc.NewLegacyControlService, control_pc.NewControlPCService, data.NewData, controller.NewControlPCController, controller.NewUnlockController,
		conf.NewChanGroup, bootstrap.NewDaemonConnectBootstrap, bootstrap.NewInternalServiceBootstrap, common_service.NewInternalService, controller.NewCustomCommandController,
		custom_command_service.NewCustomCommandService, remote_service.NewRemoteService,
		controller.NewRemoteController, bootstrap.NewRemoteConnectBootstrap, controller.NewDiscoverController, credential_provider_service.NewCredentialProviderService,
	)
	return &DesktopServiceApp{_conf: _conf, db: db}, nil
}
func initDesktopDaemonApplication(_conf *conf.Conf, db *conf.DatabaseConf) (*DesktopDaemonApp, error) {
	wire.Build(NewDesktopDaemonApp, bootstrap.NewDesktopDaemonBootstrap, bootstrap.NewInternalServiceBootstrap, common_service.NewInternalService,
		custom_command_service.NewCustomCommandService, logger.NewLogger)
	return &DesktopDaemonApp{}, nil
}
