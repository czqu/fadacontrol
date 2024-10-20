package schema

import "fadacontrol/internal/base/exception"

type UnLockData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UnLockResponse struct {
	Err *exception.Exception `json:"code"`
}
