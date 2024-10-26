package utils

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fmt"
	"golang.org/x/sys/windows"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type Interface struct {
	MACAddr       string   `json:"mac_addr"`
	InterfaceName string   `json:"interface_name"`
	IPAddresses   []net.IP `json:"ip_addresses"`
}
type AddressType uint8

const (
	IPV4 AddressType = iota
	IPV6
	UNSET
)

func GetValidInterface(t AddressType) ([]Interface, error) {
	interfaces, err := net.Interfaces()
	ret := make([]Interface, 0)
	if err != nil {
		return nil, err
	}
	if len(interfaces) == 0 {
		return ret, nil
	}
	for _, i := range interfaces {
		if !((i.Flags&net.FlagRunning) != 0 && (i.Flags&net.FlagLoopback) == 0 && ((i.Flags&net.FlagBroadcast) != 0 || (i.Flags&net.FlagMulticast) != 0)) {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		var macAddrinfo = new(Interface)
		macAddrinfo.InterfaceName = i.Name
		macAddrinfo.MACAddr = formatMAC(i.HardwareAddr)
		macAddrinfo.IPAddresses = make([]net.IP, 0)
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {

				if t == IPV4 && ipnet.IP.To4() == nil {
					continue
				}
				if t == IPV6 && ipnet.IP.To4() != nil {
					continue
				}
				ip, _, _ := net.ParseCIDR(addr.String())

				macAddrinfo.IPAddresses = append(macAddrinfo.IPAddresses, ip)

			}

		}
		ret = append(ret, *macAddrinfo)

	}
	return ret, nil

}
func formatMAC(mac []byte) string {
	if len(mac) == 0 {
		return ""
	}
	macStr := fmt.Sprintf("%x", mac)
	parts := make([]string, 0, len(mac))
	for i := 0; i < len(macStr); i += 2 {
		parts = append(parts, macStr[i:i+2])
	}
	return strings.Join(parts, ":")
}

var (
	modiphlpapi                 = windows.NewLazySystemDLL("iphlpapi.dll")
	procNotifyIpInterfaceChange = modiphlpapi.NewProc("NotifyIpInterfaceChange")
	lastCalled                  time.Time
	networkChangeRunLock        sync.Mutex
	networkChangeCallback       []func()
	networkChangeCallbackRWLock sync.RWMutex
)

type networkContext struct{}

func init() {
	context := &networkContext{}
	interfaceChange := windows.Handle(0)
	lastCalled = time.Now().Add(-time.Minute)
	ret, _, err := procNotifyIpInterfaceChange.Call(syscall.AF_UNSPEC,
		syscall.NewCallback(networkMonitorCallback),
		uintptr(unsafe.Pointer(context)),
		0,
		uintptr(unsafe.Pointer(&interfaceChange))) //this must be pointer
	if err != nil {
		logger.Error("network change callback failed " + err.Error())
		return
	}
	if ret != 0 {
		logger.Error("network change callback failed " + strconv.Itoa(int(ret)))
		return
	}

}
func AddNetworkChangeCallback(callback func()) {
	networkChangeCallbackRWLock.Lock()
	defer networkChangeCallbackRWLock.Unlock()
	networkChangeCallback = append(networkChangeCallback, callback)

}
func networkMonitorCallback(callerContext, row, notificationType uintptr) uintptr {
	if !networkChangeRunLock.TryLock() {
		return 0
	}
	defer networkChangeRunLock.Unlock()
	now := time.Now()
	//if the last call is more than 1 minute ago, call the callback
	//otherwise, ignore the callback
	networkChangeCallbackRWLock.RLock()
	defer networkChangeCallbackRWLock.RUnlock()
	if now.Sub(lastCalled) >= conf.NetWorkChangeServiceRestartInterval {
		for _, callback := range networkChangeCallback {
			callback()
		}
		lastCalled = now
	}

	return 0
}
