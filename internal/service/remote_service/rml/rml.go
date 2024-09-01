package rml

import (
	"encoding/json"
	"fadacontrol/internal/base/logger"
)

type ControlAction uint32

const (
	Unknown ControlAction = iota
	Unlock
	END
)

type ControlActionJson struct {
	Type int              `json:"type"`
	Data UnlockActionJson `json:"data"`
}
type UnlockActionJson struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ReadJson(jsonStr string) (username, password string, err error) {

	var controlAction ControlActionJson
	err = json.Unmarshal([]byte(jsonStr), &controlAction)
	if (controlAction.Type) != int(Unlock) {
		return "", "", nil
	}
	if err != nil {
		logger.Error(err)
		return "", "", err
	}
	username = controlAction.Data.Username
	password = controlAction.Data.Password
	return username, password, nil
}
