package sys

import (
	"fadacontrol/internal/schema/remote_schema"
)

type ShutdownType int

const (
	Unknown ShutdownType = iota
	// E=>通用
	S_E_LOGOFF         // 注销当前用户
	S_E_FORCE_SHUTDOWN // 强制关闭所有应用程序并关机 在Windows上即 (EWX_SHUTDOWN | EWX_FORCE)
	S_E_FORCE_REBOOT   // 强制关闭所有应用程序并重启 在Windows上即(EWX_REBOOT | EWX_FORCE)

	//EWX => Windows 专有
	S_EWX_SHUTDOWN        // 关机但不关闭电源
	S_EWX_REBOOT          // 重启计算机
	S_EWX_FORCE           // 强制关闭所有应用程序
	S_EWX_POWEROFF        // 关机并关闭电源
	S_EWX_RESTARTAPPS     // 重启应用程序
	S_EWX_HYBRID_SHUTDOWN // 混合关机

	S_EWX_FORCE_POWEROFF                    // 强制关闭所有应用程序并关机同时关闭电源 (EWX_POWEROFF | EWX_FORCE)
	S_EWX_REBOOT_RESTARTAPPS                // 重启计算机并重启应用程序 (EWX_REBOOT | EWX_RESTARTAPPS)
	S_EWX_FORCE_REBOOT_RESTARTAPPS          // 强制关闭应用程序后重启并重启应用程序 (EWX_REBOOT | EWX_FORCE | EWX_RESTARTAPPS)
	S_EWX_SHUTDOWN_RESTARTAPPS              // 关机但不关闭电源，关机后重启应用程序 (EWX_SHUTDOWN | EWX_RESTARTAPPS)
	S_EWX_HYBRID_SHUTDOWN_FORCE             // 混合关机并强制关闭应用程序 (EWX_HYBRID_SHUTDOWN | EWX_FORCE)
	S_EWX_HYBRID_SHUTDOWN_RESTARTAPPS       // 混合关机并重启应用程序 (EWX_HYBRID_SHUTDOWN | EWX_RESTARTAPPS)
	S_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS // 混合关机、强制关闭应用程序并重启应用程序 (EWX_HYBRID_SHUTDOWN | EWX_FORCE | EWX_RESTARTAPPS
)

func ProtoTypeToShutdownType(protoType remote_schema.ShutdownType) ShutdownType {
	switch protoType {
	case remote_schema.ShutdownType_E_FORCE_REBOOT:
		return S_E_FORCE_REBOOT
	case remote_schema.ShutdownType_E_LOGOFF:
		return S_E_LOGOFF
	case remote_schema.ShutdownType_E_FORCE_SHUTDOWN:
		return S_E_FORCE_SHUTDOWN
	case remote_schema.ShutdownType_EWX_FORCE_POWEROFF:
		return S_EWX_FORCE_POWEROFF
	case remote_schema.ShutdownType_EWX_REBOOT:
		return S_EWX_REBOOT
	case remote_schema.ShutdownType_EWX_REBOOT_RESTARTAPPS:
		return S_EWX_REBOOT_RESTARTAPPS
	case remote_schema.ShutdownType_EWX_POWEROFF:
		return S_EWX_POWEROFF

	case remote_schema.ShutdownType_EWX_HYBRID_SHUTDOWN:
		return S_EWX_HYBRID_SHUTDOWN
	case remote_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_FORCE:
		return S_EWX_HYBRID_SHUTDOWN_FORCE
	case remote_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_RESTARTAPPS:
		return S_EWX_HYBRID_SHUTDOWN_RESTARTAPPS
	case remote_schema.ShutdownType_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS:
		return S_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS
	case remote_schema.ShutdownType_EWX_SHUTDOWN_RESTARTAPPS:
		return S_EWX_SHUTDOWN_RESTARTAPPS
	case remote_schema.ShutdownType_EWX_SHUTDOWN:
		return S_EWX_SHUTDOWN
	case remote_schema.ShutdownType_EWX_FORCE_REBOOT_RESTARTAPPS:
		return S_EWX_FORCE_REBOOT_RESTARTAPPS

	default:
		return Unknown
	}
}
