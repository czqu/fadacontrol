package sys

/*
#cgo LDFLAGS: -lsecur32
#include "common_windows.h"
*/
import "C"
import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/goroutine"
	"fmt"
	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"
	"net"
	"os/user"
	"strconv"
	"syscall"
	"unsafe"
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

func convertType(tpe ShutdownType) (uint32, *exception.Exception) {
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
		return 0, exception.ErrUserParameterError

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
		return exception.ErrUserParameterError
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
	securityDescriptor := "D:(A;;GA;;;S-1-5-32-544)(A;;GA;;;S-1-5-18)"
	sid, err := getCurrentUserSid()
	if err == nil {
		securityDescriptor = fmt.Sprintf("D:(A;;GA;;;S-1-5-32-544)(A;;GA;;;S-1-5-18)(A;;GA;;;%s)", sid)
	}
	config := &winio.PipeConfig{
		SecurityDescriptor: securityDescriptor, // Only administrators and system accounts have full control
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
		goroutine.RecoverGO(func() {
			handler(conn)
		})

	}

}
func getCurrentUserSid() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Uid, nil
}
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
func TryLogin(username, password, domain string) *exception.Exception {
	users, _err := EnumerateUsers()
	if _err == nil {
		for _, user := range users {
			if user.Username == username {
				break
			}
			if user.FullName == username {
				username = user.Username
				break
			}
		}
	}
	usernamePtr, _ := syscall.UTF16PtrFromString(username)
	passwordPtr, _ := syscall.UTF16PtrFromString(password)
	domainPtr, _ := syscall.UTF16PtrFromString(domain)

	result := C.TryLogin((*C.wchar_t)(unsafe.Pointer(usernamePtr)), (*C.wchar_t)(unsafe.Pointer(passwordPtr)), (*C.wchar_t)(unsafe.Pointer(domainPtr)))

	ret := int(result)
	err := exception.GetErrorByCode(ret)
	if err != nil {
		return err
	}

	logger.Debug("login code :" + strconv.Itoa(ret))
	return err
}

type USER_INFO_2 struct {
	Name         *uint16
	Password     *uint16
	PasswordAge  uint32
	Priv         uint32
	HomeDir      *uint16
	Comment      *uint16
	Flags        uint32
	ScriptPath   *uint16
	AuthFlags    uint32
	FullName     *uint16
	UsrComment   *uint16
	Params       *uint16
	Workstations *uint16
	LastLogon    uint32
	LastLogoff   uint32
	AcctExpires  uint32
	MaxStorage   uint32
	UnitsPerWeek uint32
	LogonHours   *byte
	BadPwCount   uint32
	NumLogons    uint32
	LogonServer  *uint16
	CountryCode  uint32
	CodePage     uint32
}

type UserInfo struct {
	Username  string
	FullName  string
	Comment   string
	UserType  uint32
	LastLogon uint32
	Flags     uint32
}

const (
	FILTER_NORMAL_ACCOUNT uint32 = 0x0002
	MAX_PREFERRED_LENGTH  uint32 = 0xFFFFFFFF
)

func UTF16PtrToString(p *uint16) string {
	if p == nil {
		return ""
	}
	return syscall.UTF16ToString((*[4096]uint16)(unsafe.Pointer(p))[:])
}

func EnumerateUsers() ([]UserInfo, error) {
	var (
		level        uint32 = 2
		entriesRead  uint32
		totalEntries uint32
		resumeHandle uint32
	)

	users := make([]UserInfo, 0)

	for {
		var bufptr *byte

		// Try to enumerate users
		err := windows.NetUserEnum(
			nil,
			level,
			FILTER_NORMAL_ACCOUNT,
			&bufptr,
			MAX_PREFERRED_LENGTH,
			&entriesRead,
			&totalEntries,
			&resumeHandle,
		)

		if err != nil && err != syscall.ERROR_MORE_DATA {
			return nil, err
		}

		// Exit if no entries were read
		if entriesRead == 0 {
			break
		}

		// Calculate the size of USER_INFO_2 structure
		//	size := unsafe.Sizeof(USER_INFO_2{})

		// Convert buffer to slice of USER_INFO_2
		userInfos := (*[1024]USER_INFO_2)(unsafe.Pointer(bufptr))[:entriesRead:entriesRead]

		// Process each user entry
		for i := uint32(0); i < entriesRead; i++ {
			if userInfos[i].Name == nil {
				continue
			}

			username := UTF16PtrToString(userInfos[i].Name)
			if username == "" {
				continue
			}

			user := UserInfo{
				Username:  username,
				FullName:  UTF16PtrToString(userInfos[i].FullName),
				Comment:   UTF16PtrToString(userInfos[i].Comment),
				UserType:  userInfos[i].Priv,
				LastLogon: userInfos[i].LastLogon,
				Flags:     userInfos[i].Flags,
			}

			users = append(users, user)
		}

		// Free the buffer allocated by the system
		windows.NetApiBufferFree(bufptr)

		//// Break if we've got all entries
		if err != syscall.ERROR_MORE_DATA {
			break
		}
	}

	return users, nil
}
