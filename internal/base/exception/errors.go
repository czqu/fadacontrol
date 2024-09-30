package exception

import "fmt"

type Exception struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *Exception) Error() string {
	return e.Msg
}
func (e *Exception) Equal(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *Exception:
		return e.Code == err.(*Exception).Code
	default:
		return false
	}

}
func (e *Exception) NotEqual(err error) bool {
	if err == nil {
		return true
	}
	switch err.(type) {
	case *Exception:
		return e.Code != err.(*Exception).Code
	default:
		return true
	}
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
