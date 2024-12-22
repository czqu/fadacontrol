package unlock

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/credential_provider_service"
	"fadacontrol/pkg/sys"
	"fadacontrol/pkg/utils"
)

type UnLockService struct {
	cp *credential_provider_service.CredentialProviderService
}

func NewUnLockService(cp *credential_provider_service.CredentialProviderService) *UnLockService {
	return &UnLockService{cp: cp}
}

func (u *UnLockService) UnlockPc(username string, password string) *exception.Exception {
	data := []byte("\x01" + "\x00" + username + "\x00" + password + "\x00")
	var packet = entity.PipePacket{}
	packet.Tpe = entity.UnlockReq
	packet.Size = uint32(len(data))
	packet.Data = data
	packet.ReqId = u.cp.GenReqId()

	logger.Debugf("send packet")
	ret := u.cp.SendData(&packet)
	logger.Debugf("ret: %v", ret.Msg)
	if exception.ErrUserUnlockNotInLockScreenState.Equal(ret) {

		logger.Debugf("try login")
		domain, account := utils.SplitWindowsAccount(username)
		err := sys.TryLogin(account, password, domain)
		if exception.ErrSuccess.NotEqual(err) {
			logger.Errorf("err: %v", err)
			return err
		}
		return exception.ErrUserUnlockNotInLockScreenState
	}
	if exception.ErrSuccess.NotEqual(ret) {
		logger.Errorf("err: %v", ret)
	}
	return ret

}
