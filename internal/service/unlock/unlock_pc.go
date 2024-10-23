package unlock

import (
	"bytes"
	"fadacontrol/pkg/sys"

	"encoding/json"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/credential_provider_service"
	"net"
)

type UnLockService struct {
	cp *credential_provider_service.CredentialProviderService
}

func NewUnLockService(cp *credential_provider_service.CredentialProviderService) *UnLockService {
	return &UnLockService{cp: cp}
}

func (u *UnLockService) HandleUnlockConnection(conn net.Conn) {
	logger.Info("new connection received", conn.RemoteAddr())
	defer conn.Close()

	data := make([]byte, 1024)
	_, err := conn.Read(data)
	conn.Close()
	if err != nil {
		logger.Error("Unable to read message:", err)
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

	u.UnlockPc(username, password)
}

func (u *UnLockService) UnlockPc(username string, password string) *exception.Exception {
	data := []byte("\x01" + "\x00" + username + "\x00" + password + "\x00")
	var packet = entity.PipePacket{}
	packet.Tpe = entity.UnlockReq
	packet.Size = uint32(len(data))
	packet.Data = data

	//todo
	ret := u.cp.SendData(&packet)
	if exception.ErrSuccess.NotEqual(ret) {

		logger.Debugf("try login")
		err := sys.TryLogin(username, password, "")
		if exception.ErrSuccess.NotEqual(err) {
			logger.Debugf("err: %v", err)
			return err
		}
		return exception.ErrUserUnlockNotInLockScreenState
	}
	return ret

}
