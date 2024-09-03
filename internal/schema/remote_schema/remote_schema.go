package remote_schema

type RemoteConfigReqDTO struct {
	Enabled      bool   `json:"enabled"`
	ApiServerUrl string `json:"api_server_url"`
	MsgServerUrl string `json:"msg_server_url"`
	Secret       string `json:"secret"`
	ClientId     string `json:"client_id"`
}
type RemoteConfigRespDTO struct {
	Enabled      bool   `json:"enabled"`
	ApiServerUrl string `json:"api_server_url"`
	MsgServerUrl string `json:"msg_server_url"`
	Key          string `json:"key"`
	ClientId     string `json:"client_id"`
}
