package unlock

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/pkg/sys"

	"net"
)

type UnLockService struct {
}

func NewUnLockService() *UnLockService {
	return &UnLockService{}
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

var resp = make(chan entity.UnLockResponse)

const (
	pipePrefix    = `\\.\pipe\fc.pipe.`
	pipeCacheSize = 4 * 1024
)
const UnlockPipeName = pipePrefix + "v1.unlock"

func (u *UnLockService) UnlockPc(username string, password string) *exception.Exception {

	go func() {
		err := sys.SendToNamedPipeWithHandler(UnlockPipeName, []byte("\x01"+"\x00"+username+"\x00"+password+"\x00"), u.unlockHandler)
		if err != nil {
			logger.Errorf("send data err,while unlock pc，more: %v", err)
			resp <- entity.UnLockResponse{Err: exception.ErrSystemUnknownException}
		}

	}()

	res := <-resp

	logger.Infof("unlock status  :%d %s", res.Err.Code, res.Err.Msg)
	return res.Err

}
func (u *UnLockService) unlockHandler(conn net.Conn) {
	data := make([]byte, pipeCacheSize)
	data = data[:4]
	_, err := conn.Read(data)
	if err != nil {
		resp <- entity.UnLockResponse{exception.ErrSystemUnknownException}
		return
	}

	code := int(binary.LittleEndian.Uint32(data))
	logger.Debug("unlock code:", code)
	if len(data) > binary.Size(code) {
		resp <- entity.UnLockResponse{exception.GetErrorByCode(code)}
	} else {
		logger.Error("Received data too small to contain an int64")
		resp <- entity.UnLockResponse{exception.ErrSystemUnknownException}
		return
	}

}
