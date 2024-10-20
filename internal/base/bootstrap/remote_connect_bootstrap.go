package bootstrap

import (
	"errors"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/service/remote_service"
	"fadacontrol/pkg/goroutine"
	"golang.org/x/sys/windows"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var RemoteConnectBootstrapInstance *RemoteConnectBootstrap

type RemoteConnectBootstrap struct {
	_conf *conf.Conf

	db              *gorm.DB
	re              *remote_service.RemoteService
	haveSetCallback bool
}

func NewRemoteConnectBootstrap(_conf *conf.Conf, db *gorm.DB, re *remote_service.RemoteService) *RemoteConnectBootstrap {
	RemoteConnectBootstrapInstance = &RemoteConnectBootstrap{re: re, _conf: _conf, db: db}
	return RemoteConnectBootstrapInstance
}
func (r *RemoteConnectBootstrap) Start() error {
	if !r.haveSetCallback {
		SetNetworkChangeCallback(func() {
			r.re.RestartService()
			r.haveSetCallback = true
		})
	}

	goroutine.RecoverGO(func() {
		r.re.StartService()
	})

	return nil
}
func (r *RemoteConnectBootstrap) Stop() error {
	return r.re.StopService()

}

var (
	modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

	procNotifyIpInterfaceChange = modiphlpapi.NewProc("NotifyIpInterfaceChange")
	lastCalled                  time.Time
	networkChangeRunLock        sync.Mutex
	onNetworkChangeCallback     func()
)

type networkContext struct{}

func SetNetworkChangeCallback(callback func()) error {
	onNetworkChangeCallback = callback

	context := &networkContext{}
	interfaceChange := windows.Handle(0)
	lastCalled = time.Now().Add(-time.Minute)
	ret, _, err := procNotifyIpInterfaceChange.Call(syscall.AF_UNSPEC,
		syscall.NewCallback(networkMonitorCallback),
		uintptr(unsafe.Pointer(context)),
		0,
		uintptr(unsafe.Pointer(&interfaceChange))) //this must be pointer
	if err != nil {
		return err
	}
	if ret != 0 {
		return errors.New("network change callback failed " + strconv.Itoa(int(ret)))
	}
	return nil

}
func networkMonitorCallback(callerContext, row, notificationType uintptr) uintptr {
	if !networkChangeRunLock.TryLock() {
		return 0
	}
	defer networkChangeRunLock.Unlock()
	now := time.Now()
	if now.Sub(lastCalled) >= time.Minute {
		onNetworkChangeCallback()
		lastCalled = now
	}

	return 0
}
