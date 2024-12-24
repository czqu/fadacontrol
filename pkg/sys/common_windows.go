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
		return fmt.Errorf("error writing  data to pipe: %v", err)
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

// Constants
const (
	WTS_CURRENT_SERVER_HANDLE = 0
	WTS_ACTIVE                = 0
	WTS_USER_NAME             = 5
)

var (
	wtsapi32                 = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSEnumerateSessions = wtsapi32.NewProc("WTSEnumerateSessionsW")
	procWTSQuerySessionInfo  = wtsapi32.NewProc("WTSQuerySessionInformationW")
	procWTSFreeMemory        = wtsapi32.NewProc("WTSFreeMemory")
	procWTSQueryUserToken    = wtsapi32.NewProc("WTSQueryUserToken")
)

type WTSSessionInfo struct {
	SessionID      uint32
	WinStationName *uint16
	State          uint32
}

// WTS_SESSION_INFO structure
type WTSSessionInfoExtended struct {
	SessionID      uint32
	WinStationName string
	State          uint32
	Username       string
}

// Imports from advapi32.dll
var (
	advapi32             = windows.NewLazySystemDLL("advapi32.dll")
	procDuplicateTokenEx = advapi32.NewProc("DuplicateTokenEx")
)

// Imports from kernel32.dll
var (
	moduserenv              *windows.LazyDLL = windows.NewLazySystemDLL("userenv.dll")
	kernel32                                 = windows.NewLazySystemDLL("kernel32.dll")
	procCreateProcessAsUser                  = kernel32.NewProc("CreateProcessAsUserW")
)
var (
	procCreateProcessAsUserW                     = advapi32.NewProc("CreateProcessAsUserW")
	procCreateEnvironmentBlock *windows.LazyProc = moduserenv.NewProc("CreateEnvironmentBlock")
)

func GetLastErrorMessage() string {
	// Get the last error code
	lastError := syscall.GetLastError()

	// If there's no error, return a success message
	if lastError == nil {
		return ""
	}

	// Return the formatted error message
	return fmt.Sprintf("%s", lastError)
}

const (
	CREATE_UNICODE_ENVIRONMENT uint16 = 0x00000400
	CREATE_NO_WINDOW                  = 0x08000000
	CREATE_NEW_CONSOLE                = 0x00000010
)

type SW int

const (
	SW_HIDE            SW = 0
	SW_SHOWNORMAL         = 1
	SW_NORMAL             = 1
	SW_SHOWMINIMIZED      = 2
	SW_SHOWMAXIMIZED      = 3
	SW_MAXIMIZE           = 3
	SW_SHOWNOACTIVATE     = 4
	SW_SHOW               = 5
	SW_MINIMIZE           = 6
	SW_SHOWMINNOACTIVE    = 7
	SW_SHOWNA             = 8
	SW_RESTORE            = 9
	SW_SHOWDEFAULT        = 10
	SW_MAX                = 1
)

func StartProcessForSession(sessionID uint32, appPath, cmdLine string, workDir string, runas bool) error {
	var (
		envInfo windows.Handle

		startupInfo windows.StartupInfo
		processInfo windows.ProcessInformation

		commandLine uintptr = 0
		workingDir  uintptr = 0

		err error
	)
	var userToken windows.Token

	// Get the user token for the session
	ret, _, err := procWTSQueryUserToken.Call(
		uintptr(sessionID),
		uintptr(unsafe.Pointer(&userToken)),
	)
	if ret == 0 {
		return fmt.Errorf("failed to query user token for session %d: %v %v", sessionID, err, GetLastErrorMessage())
	}
	defer userToken.Close()

	if returnCode, _, err := procCreateEnvironmentBlock.Call(uintptr(unsafe.Pointer(&envInfo)), uintptr(userToken), 0); returnCode == 0 {
		return fmt.Errorf("create environment details for process: %s", err)
	}

	creationFlags := CREATE_UNICODE_ENVIRONMENT | CREATE_NEW_CONSOLE
	startupInfo.ShowWindow = SW_SHOW
	startupInfo.Desktop = windows.StringToUTF16Ptr("winsta0\\default")

	if len(cmdLine) > 0 {
		commandLine = uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(cmdLine)))
	}
	if len(workDir) > 0 {
		workingDir = uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(workDir)))
	}
	if returnCode, _, err := procCreateProcessAsUser.Call(
		uintptr(userToken), uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(appPath))), commandLine, 0, 0, 0,
		uintptr(creationFlags), uintptr(envInfo), workingDir, uintptr(unsafe.Pointer(&startupInfo)), uintptr(unsafe.Pointer(&processInfo)),
	); returnCode == 0 {
		return fmt.Errorf("create process as user: %s", err)
	}
	return nil
}

// EnumerateSessions retrieves all active sessions on the system.
func EnumerateSessions() ([]WTSSessionInfoExtended, error) {
	var sessionInfo uintptr
	var count uint32

	ret, _, err := procWTSEnumerateSessions.Call(
		WTS_CURRENT_SERVER_HANDLE,
		0,
		1,
		uintptr(unsafe.Pointer(&sessionInfo)),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("failed to enumerate sessions: %v", err)
	}
	defer procWTSFreeMemory.Call(sessionInfo)

	// Parse session information
	sessions := make([]WTSSessionInfoExtended, count)
	rawSessions := (*[1 << 16]WTSSessionInfo)(unsafe.Pointer(sessionInfo))[:count:count]

	for i, rawSession := range rawSessions {
		username, err := GetSessionUsername(rawSession.SessionID)
		if err != nil {
			username = "(unknown)"
		}

		sessions[i] = WTSSessionInfoExtended{
			SessionID:      rawSession.SessionID,
			WinStationName: windows.UTF16PtrToString(rawSession.WinStationName),
			State:          rawSession.State,
			Username:       username,
		}
	}

	return sessions, nil
}

func GetSessionUsername(sessionID uint32) (string, error) {
	var buffer *uint16
	var bytesReturned uint32

	ret, _, err := procWTSQuerySessionInfo.Call(
		WTS_CURRENT_SERVER_HANDLE,
		uintptr(sessionID),
		WTS_USER_NAME,
		uintptr(unsafe.Pointer(&buffer)),
		uintptr(unsafe.Pointer(&bytesReturned)),
	)
	if ret == 0 {
		return "", fmt.Errorf("failed to query session information: %v", err)
	}
	defer procWTSFreeMemory.Call(uintptr(unsafe.Pointer(buffer)))

	username := windows.UTF16PtrToString(buffer)
	return username, nil
}

func RunProgramForAllUser(programPath string, commandline, workdir string) error {

	if programPath == "" {
		return fmt.Errorf("program path is empty")
	}
	sessions, err := EnumerateSessions()
	if err != nil {

		return err
	}
	err = EnablePrivilege("SeTcbPrivilege")
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.State == WTS_ACTIVE && session.Username != "" {
			err := StartProcessForSession(session.SessionID, programPath, commandline, workdir, true)
			if err != nil {
				logger.Errorf("failed to launch program for session %d: %v", session.SessionID, err)
				continue
			}
			logger.Debugf("launched program for session %d,username: %s", session.SessionID, session.Username)
		}
	}
	return nil
}

const (
	SE_PRIVILEGE_ENABLED = 0x00000002
)

var (
	procAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
	procLookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueW")
)

func EnablePrivilege(privilegeName string) error {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("failed to open process token: %w", err)
	}
	defer token.Close()

	var luid windows.LUID
	privName, err := windows.UTF16PtrFromString(privilegeName)
	if err != nil {
		return fmt.Errorf("failed to encode privilege name: %w", err)
	}

	ret, _, err := procLookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(privName)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if ret == 0 {
		return fmt.Errorf("failed to lookup privilege value: %w", err)
	}

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: SE_PRIVILEGE_ENABLED},
		},
	}

	ret, _, err = procAdjustTokenPrivileges.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("failed to adjust token privileges: %w", err)
	}

	return nil
}
