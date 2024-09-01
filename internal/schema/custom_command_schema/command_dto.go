package custom_command_schema

type CustomCommandReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
type CustomCommandResp struct {
	Id string `json:"id"`
}
