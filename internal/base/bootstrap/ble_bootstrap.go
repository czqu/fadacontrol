package bootstrap

import (
	"bytes"
	"encoding/json"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/unlock"
	"fadacontrol/pkg/sys/bluetooth"

	"net"
)

type BleUnlockBootstrap struct {
	u        *unlock.UnLockService
	listener net.Listener
}

func NewBleUnlockBootstrap(u *unlock.UnLockService) *BleUnlockBootstrap {
	return &BleUnlockBootstrap{u: u}
}

//func init() {
//	RegisterService(&BleUnlockBootstrap{})
//}

func (d *BleUnlockBootstrap) Start() error {
	go d.StartServer()
	return nil
}
func (d *BleUnlockBootstrap) StartServer() error {
	config := &bluetooth.Config{ServiceInstanceName: "RemoteFingerprint Service ", Comment: "RemoteFingerprint Service "}

	serviceClassId := bluetooth.GUID{
		Data1: 0x4E5877C0,
		Data2: 0x8297,
		Data3: 0x4AAE,
		Data4: [8]byte{0xB7, 0xBD, 0x73, 0xA8, 0xCB, 0xC1, 0xED, 0xAF},
	}
	var err error
	d.listener, err = bluetooth.Listen(serviceClassId, config)
	if err != nil {
		logger.Errorf("Failed to listen: %v", err)
		return err
	}
	defer d.listener.Close()
	for {
		logger.Infof("Waiting for a ble connection...")
		conn, err := d.listener.Accept()
		if err != nil {
			logger.Errorf("Failed to accept connection: %v", err)
			return err
		}

		go d.handleConnection(conn)
	}
}
func (d *BleUnlockBootstrap) handleConnection(conn net.Conn) {
	defer conn.Close()

	data := make([]byte, 1024)
	_, err := conn.Read(data)
	if err != nil {
		logger.Error("Unable to read the message:", err)
		return
	}
	var message map[string]string
	data = bytes.TrimRight(data, "\x00")
	err = json.Unmarshal(data, &message)
	if err != nil {
		logger.Error("Unable to parse message:", err)
		return
	}
	username, ok := message["username"]
	if !ok {
		logger.Error("username not found")
		return
	}

	password, ok := message["passwd"]
	if !ok {
		logger.Error("Password not found")
		return
	}

	d.u.UnlockPc(username, password)
}
func (d *BleUnlockBootstrap) Stop() error {
	return d.listener.Close()

}
