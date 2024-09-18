package sys

/*
#cgo LDFLAGS: -lsecur32
#include "common_windows.h"
*/
import "C"
import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fmt"
	"github.com/Microsoft/go-winio"
	"net"
)

const (
	EWX_HYBRID_SHUTDOWN                   = 0x00400000
	EWX_LOGOFF                            = 0
	EWX_POWEROFF                          = 0x00000008
	EWX_REBOOT                            = 0x00000002
	EWX_RESTARTAPPS                       = 0x00000040
	EWX_SHUTDOWN                          = 0x00000001
	EWX_FORCE                             = 0x00000004
	EWX_FORCE_POWEROFF                    = EWX_POWEROFF | EWX_FORCE                          // 强制关闭所有应用程序并关机同时关闭电源 (EWX_POWEROFF | EWX_FORCE)
	EWX_REBOOT_RESTARTAPPS                = EWX_REBOOT | EWX_RESTARTAPPS                      // 重启计算机并重启应用程序 (EWX_REBOOT | EWX_RESTARTAPPS)
	EWX_FORCE_REBOOT_RESTARTAPPS          = EWX_REBOOT | EWX_FORCE | EWX_RESTARTAPPS          // 强制关闭应用程序后重启并重启应用程序 (EWX_REBOOT | EWX_FORCE | EWX_RESTARTAPPS)
	EWX_SHUTDOWN_RESTARTAPPS              = EWX_SHUTDOWN | EWX_RESTARTAPPS                    // 关机但不关闭电源，关机后重启应用程序 (EWX_SHUTDOWN | EWX_RESTARTAPPS)
	EWX_HYBRID_SHUTDOWN_FORCE             = EWX_HYBRID_SHUTDOWN | EWX_FORCE                   // 混合关机并强制关闭应用程序 (EWX_HYBRID_SHUTDOWN | EWX_FORCE)
	EWX_HYBRID_SHUTDOWN_RESTARTAPPS       = EWX_HYBRID_SHUTDOWN | EWX_RESTARTAPPS             // 混合关机并重启应用程序 (EWX_HYBRID_SHUTDOWN | EWX_RESTARTAPPS)
	EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS = EWX_HYBRID_SHUTDOWN | EWX_FORCE | EWX_RESTARTAPPS // 混合关机、强制关闭应用程序并重启应用程序 (EWX_HYBRID_SHUTDOWN | EWX_FORCE | EWX_RESTARTAPPS

	E_FORCE_SHUTDOWN = EWX_SHUTDOWN | EWX_FORCE // 强制关闭所有应用程序并关机 在Windows上即 (EWX_SHUTDOWN | EWX_FORCE)
	E_FORCE_REBOOT   = EWX_REBOOT | EWX_FORCE   // 强制关闭所有应用程序并重启 在Windows上即(EWX_REBOOT | EWX_FORCE)

)

func convertType(tpe ShutdownType) (uint32, error) {
	switch tpe {
	case S_E_LOGOFF:
		return EWX_LOGOFF, nil
	case S_E_FORCE_SHUTDOWN:
		return E_FORCE_SHUTDOWN, nil
	case S_E_FORCE_REBOOT:
		return E_FORCE_REBOOT, nil

	case S_EWX_REBOOT:
		return EWX_REBOOT, nil
	case S_EWX_FORCE:
		return EWX_FORCE, nil
	case S_EWX_FORCE_POWEROFF:
		return EWX_FORCE_POWEROFF, nil
	case S_EWX_POWEROFF:
		return EWX_POWEROFF, nil
	case S_EWX_RESTARTAPPS:
		return EWX_RESTARTAPPS, nil
	case S_EWX_SHUTDOWN:
		return EWX_SHUTDOWN, nil
	case S_EWX_REBOOT_RESTARTAPPS:
		return EWX_REBOOT_RESTARTAPPS, nil
	case S_EWX_FORCE_REBOOT_RESTARTAPPS:
		return EWX_FORCE_REBOOT_RESTARTAPPS, nil
	case S_EWX_SHUTDOWN_RESTARTAPPS:
		return EWX_SHUTDOWN_RESTARTAPPS, nil
	case S_EWX_HYBRID_SHUTDOWN_FORCE:
		return EWX_HYBRID_SHUTDOWN_FORCE, nil
	case S_EWX_HYBRID_SHUTDOWN_RESTARTAPPS:
		return EWX_HYBRID_SHUTDOWN_RESTARTAPPS, nil
	case S_EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS:
		return EWX_HYBRID_SHUTDOWN_FORCE_RESTARTAPPS, nil
	case S_EWX_HYBRID_SHUTDOWN:
		return EWX_HYBRID_SHUTDOWN, nil
	default:
		return 0, fmt.Errorf("unsupport shutdown type %d", tpe)

	}
}
func Shutdown(tpe ShutdownType) *exception.Exception {
	result := C.PreCheckShutdownWindows()
	if int(result) != 0 {
		logger.Errorf("shutdown err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	t, err := convertType(tpe)
	if err != nil {
		return exception.ErrParameterError
	}
	result = C.ShutdownWindows(C.UINT(t))
	if int(result) != 0 {
		logger.Errorf("shutdown err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	return exception.ErrSuccess
}

func Standby() *exception.Exception {
	result := C.PreCheckStandbyWindows()
	if int(result) != 0 {
		logger.Errorf("Standby err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	result = C.StandbyWindows()
	if int(result) != 0 {
		logger.Errorf("Standby err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	return exception.ErrSuccess
}
func LockWindows() *exception.Exception {
	result := C.LockWindows()
	if int(result) != 0 {
		logger.Errorf("LockWindows err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	return exception.ErrSuccess
}

func CheckGrant() int {
	result := C.checkGrant()
	return int(result)
}

func ListenNamedPipe(pipeName string, handler func(conn net.Conn)) error {

	pipeConfig := &winio.PipeConfig{
		SecurityDescriptor: "D:P(A;;GA;;;WD)",
	}
	pipeListener, err := winio.ListenPipe(pipeName, pipeConfig)
	if err != nil {
		return fmt.Errorf("failed to listen on named pipe: %v", err)
	}

	conn, err := pipeListener.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept connection on named pipe: %v", err)
	}

	handler(conn)
	return nil
}
func SendToNamedPipe(pipeName string, data []byte) error {

	pipeHandle, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to named pipe: %v", err)
	}
	defer pipeHandle.Close()

	if _, err := pipeHandle.Write(data); err != nil {
		return fmt.Errorf("error writing JSON data to pipe: %v", err)
	}

	return nil
}
func ReceiveFromNamedPipe(pipeName string, pipeCacheSize int) ([]byte, error) {

	pipeHandle, err := winio.DialPipe(pipeName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to named pipe: %v", err)
	}
	defer pipeHandle.Close()

	data := make([]byte, pipeCacheSize)
	_, err = pipeHandle.Read(data)
	if err != nil {
		return nil, fmt.Errorf("error reading from pipe: %v", err)
	}

	return data, nil
}

func ListenNamedPipeWithHandler(pipeName string, handler func(conn net.Conn), inputBufferSize, outputBufferSize int32) error {
	config := &winio.PipeConfig{
		SecurityDescriptor: "D:(A;;GA;;;S-1-5-32-544)(A;;GA;;;S-1-5-18)", // Only administrators and system accounts have full control
		MessageMode:        false,
		InputBufferSize:    inputBufferSize,
		OutputBufferSize:   outputBufferSize,
	}
	pipeHandle, err := winio.ListenPipe(pipeName, config)

	if err != nil {
		return fmt.Errorf("failed to connect to named pipe: %v", err)
	}
	for {
		conn, err := pipeHandle.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection on named pipe: %v", err)
		}

		logger.Debugf("new pipe accepted")
		handler(conn)
	}

}

// func WriteNamedPipeWithHandler(pipeName string, data []byte, handler func(conn net.Conn)) error {
//
//		pipeHandle, err := winio.DialPipe(pipeName, nil)
//		if err != nil {
//			return fmt.Errorf("failed to connect to named pipe: %v", err)
//		}
//		defer pipeHandle.Close()
//
//		if _, err := pipeHandle.Write(data); err != nil {
//			return fmt.Errorf("error writing JSON data to pipe: %v", err)
//		}
//		handler(pipeHandle)
//
//		return nil
//	}
func SetPowerSavingMode(enable bool) bool {
	var cEnable C.bool
	if enable {
		cEnable = C.bool(true)
	} else {
		cEnable = C.bool(false)
	}

	result := C.SetProcessPowerSavingMode(cEnable)
	return bool(result)
}
