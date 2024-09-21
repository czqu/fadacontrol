package schema

type ResponseData struct {
	RequestId string      `json:"request_id"`
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
}
