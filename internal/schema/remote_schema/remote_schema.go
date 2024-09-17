package remote_schema

type RemoteServer struct {
	MsgServerUrl string `json:"msg_server_url"`
	ApiServerUrl string `json:"api_server_url"`
}
type RemoteConfigReqDTO struct {
	Enabled  bool           `json:"enabled"`
	Server   []RemoteServer `json:"server"`
	Secret   string         `json:"secret"`
	ClientId string         `json:"client_id"`
}
type RemoteConfigRespDTO struct {
	Enabled  bool           `json:"enabled"`
	Server   []RemoteServer `json:"server"`
	Key      string         `json:"key"`
	ClientId string         `json:"client_id"`
}
