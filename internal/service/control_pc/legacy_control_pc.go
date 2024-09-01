package control_pc

import (
	"bytes"
	"encoding/json"
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/service/unlock"
	"net"
)

type LegacyControlService struct {
	o *ControlPCService
	u *unlock.UnLockService
}

func NewLegacyControlService(o *ControlPCService, u *unlock.UnLockService) *LegacyControlService {
	return &LegacyControlService{o: o, u: u}
}

type QueryMsg struct {
	Action      string `json:"action"`
	Path        string `json:"path"`
	Password    string `json:"password"`
	RawPassword string `json:"raw_password"`
	RawUsername string `json:"raw_username"`
}
type FileInfoMsg struct {
	Version int `json:"version"`
	//FileName string `json:"file_name"`
	//FileSize      int          `json:"file_size"`
	//FileSizeHigh  int          `json:"file_size_high"`
	Status        int          `json:"status"`
	MacAddr       string       `json:"macaddr"`
	Mac           string       `json:"mac"`
	System        string       `json:"system"`
	Detail        []FileDetail `json:"detail"`
	CurrentFolder string       `json:"current_folder"`
}
type RespMsg struct {
	Version int    `json:"version"`
	Status  int    `json:"status"`
	MacAddr string `json:"macaddr"`
	Mac     string `json:"mac"`
	System  string `json:"system"`
}

type FileDetail struct {
	FileName   string `json:"file_name"`
	Attributes int    `json:"attributes"`
}

func (l *LegacyControlService) HandleControlConnection(conn net.Conn) {

	data := make([]byte, 1024)
	_, err := conn.Read(data)
	conn.Close()
	logger.Info("receive conn:", conn.RemoteAddr())
	if err != nil {
		logger.Error("Unable to read the message:", err)
		return
	}

	var message QueryMsg

	data = bytes.TrimRight(data, "\x00")
	err = json.Unmarshal(data, &message)
	if err != nil {
		logger.Error("error:", err)
		return
	}

	switch message.Action {

	case "Query":

		data, err = json.Marshal(buildFileERR())
		if err != nil {
			logger.Error("error:", err)
		}

		break
	case "standby":
		err := l.o.Standby()
		if err != nil {
			logger.Error(err)
			data = buildErr()
		}
		data = buildSuccess()
		break
	case "shutdown":

		err := l.o.Shutdown()
		if err != nil {
			logger.Error(err)
			data = buildErr()
		}
		data = buildSuccess()
		break
	case "lock":

		err := l.o.LockWindows(true)
		if err != nil && err != exception.ErrSuccess {
			logger.Error(err)
			data = buildErr()
		}
		data = buildSuccess()
		break
	case "unlock":
		err := l.u.UnlockPc(message.RawUsername, message.RawPassword)
		if err != nil {
			logger.Error(err)
			data = buildErr()
		}
		data = buildSuccess()
		break
	default:
		msg := &RespMsg{
			Version: 16,
			Status:  -1,
			MacAddr: "ffffffffffff",
			Mac:     "ffffffffffff",
			System:  "Windows",
		}
		data, err = json.Marshal(msg)
		if err != nil {
			logger.Error("error:", err)
			return
		}

	}
	logger.Debug("write data...")
	conn.Write(data)
}
func buildSuccess() []byte {
	msg := &RespMsg{
		Version: 16,
		Status:  -1,
		MacAddr: "ffffffffffff",
		Mac:     "ffffffffffff",
		System:  "Windows",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error:", err)
		return nil
	}

	return data
}
func buildErr() []byte {
	msg := &RespMsg{
		Version: 16,
		Status:  0,
		MacAddr: "ffffffffffff",
		Mac:     "ffffffffffff",
		System:  "Windows",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("error:", err)
		return nil
	}

	return data
}
func buildFileERR() *FileInfoMsg {
	msg := &FileInfoMsg{
		Version: 24082400,

		Status:        0,
		MacAddr:       "6C2408FF0EA4",
		Mac:           "6C2408FF0EA4",
		System:        "Windows",
		CurrentFolder: "/.",
		Detail: []FileDetail{
			{
				FileName:   "App version is outdated, please upgrade!",
				Attributes: 1},
		},
	}
	return msg
}
