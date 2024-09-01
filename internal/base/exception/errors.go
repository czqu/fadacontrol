package exception

import "fmt"

type Exception struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Exception) Error() string {
	return e.Msg
}
func (e *Exception) Equal(err *Exception) bool {
	return e.Code == err.Code
}
func (e *Exception) NotEqual(err *Exception) bool {
	return e.Code != err.Code
}
func (e *Exception) GetMsg() string {
	return e.Msg
}
func (e *Exception) GetCode() int {
	return e.Code
}
func (e *Exception) ToString() string {
	return fmt.Sprintf("%d: %s", e.GetCode(), e.GetMsg())
}
