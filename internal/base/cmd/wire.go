//go:build wireinject
// +build wireinject

package cmd

import (
	"fadacontrol/internal/base/bootstrap"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/middleware"
	"fadacontrol/internal/controller/admin_controller"
	"fadacontrol/internal/controller/common_controller"
	"fadacontrol/internal/router"
	"fadacontrol/internal/service/auth_service"
	"fadacontrol/internal/service/common_service"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/internal/service/custom_command_service"
	"fadacontrol/internal/service/discovery_service"
	"fadacontrol/internal/service/http_service"
	"fadacontrol/internal/service/jwt_service"
	"fadacontrol/internal/service/remote_service"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/internal/service/user_service"
	"github.com/google/wire"
)

func initDesktopServiceApplication(_conf *conf.Conf, db *conf.DatabaseConf) (*DesktopServiceApp, error) {
	wire.Build(NewDesktopServiceApp, bootstrap.NewBleUnlockBootstrap, bootstrap.NewHttpBootstrap, bootstrap.NewDiscoverBootstrap,
		bootstrap.NewLegacyBootstrap, logger.NewLogger, unlock.NewUnLockService, data.NewDB, router.NewCommonRouter, router.NewAdminRouter, bootstrap.NewRootBootstrap,
		control_pc.NewLegacyControlService, control_pc.NewControlPCService, data.NewData, common_controller.NewControlPCController, common_controller.NewUnlockController,
		conf.NewChanGroup, bootstrap.NewDaemonConnectBootstrap, bootstrap.NewInternalServiceBootstrap, common_service.NewInternalService, common_controller.NewCustomCommandController,
		custom_command_service.NewCustomCommandService, remote_service.NewRemoteService,
		admin_controller.NewRemoteController, bootstrap.NewRemoteConnectBootstrap, admin_controller.NewDiscoverController, credential_provider_service.NewCredentialProviderService,
		bootstrap.NewDataInitBootstrap, data.NewAdapterByDB, data.NewEnforcer, common_controller.NewAuthController,
		middleware.NewJwtMiddleware, jwt_service.NewJwtService, auth_service.NewAuthService, user_service.NewUserService, discovery_service.NewDiscoverService,
		common_controller.NewSysInfoController, admin_controller.NewHttpController, http_service.NewHttpService,
	)
	return &DesktopServiceApp{_conf: _conf, db: db}, nil
}
func initDesktopDaemonApplication(_conf *conf.Conf, db *conf.DatabaseConf) (*DesktopDaemonApp, error) {
	wire.Build(NewDesktopDaemonApp, bootstrap.NewDesktopDaemonBootstrap, bootstrap.NewInternalServiceBootstrap, common_service.NewInternalService,
		custom_command_service.NewCustomCommandService, logger.NewLogger, bootstrap.NewDataInitBootstrap, data.NewDB,
		data.NewAdapterByDB, data.NewEnforcer)

	return &DesktopDaemonApp{}, nil
}
