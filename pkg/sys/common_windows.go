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
	EWX_HYBRID_SHUTDOWN = 0x00400000
	EWX_LOGOFF          = 0
	EWX_POWEROFF        = 0x00000008
	EWX_REBOOT          = 0x00000002
	EWX_RESTARTAPPS     = 0x00000040
	EWX_SHUTDOWN        = 0x00000001
	EWX_FORCE           = 0x00000004
)

func Shutdown() *exception.Exception {
	result := C.PreCheckShutdownWindows()
	if int(result) != 0 {
		logger.Errorf("shutdown err %d", int(result))
		return exception.GetErrorByCode(int(result))
	}
	result = C.ShutdownWindows(EWX_SHUTDOWN | EWX_FORCE)
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

func ListenNamedPipeWithHandler(pipeName string, handler func(conn net.Conn)) error {
	config := &winio.PipeConfig{
		SecurityDescriptor: "", // 默认安全描述符
		MessageMode:        false,
		InputBufferSize:    4096,
		OutputBufferSize:   4096,
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
