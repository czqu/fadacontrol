package util

import "sync"

type ProductModule string

const (
	BluetoothUnlockModule ProductModule = "bluetooth_unlock"
	RemoteUnlockModule    ProductModule = "remote_unlock"
)

func (p ProductModule) String() string {
	return string(p)
}

var (
	SupportModulesCache map[string]bool
	SupportModulesLock  sync.RWMutex
)

func (p ProductModule) IsSupport() bool {
	SupportModulesLock.RLock()
	defer SupportModulesLock.RUnlock()
	_, ok := SupportModulesCache[p.String()]
	return ok
}
