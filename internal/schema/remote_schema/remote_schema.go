package remote_schema

// RemoteConnectConfigRequest
type RemoteConnectConfigRequest struct {
	Enable         bool     `json:"enable"`
	ClientId       string   `json:"client_id"`
	TimeStampCheck bool     `json:"time_stamp_check"`
	ApiServerUrl   string   `json:"api_server_url"`
	MsgServerUrls  []string `json:"msg_server_urls"`
}

// RemoteConnectConfigResponse
type RemoteConnectConfigResponse struct {
	Enable         bool     `json:"enable"`
	ClientId       string   `json:"client_id"`
	SecurityKey    string   `json:"security_key"`
	TimeStampCheck bool     `json:"time_stamp_check"`
	ApiServerUrl   string   `json:"api_server_url"`
	MsgServerUrls  []string `json:"msg_server_urls"`
}

// RemoteMsgServerRequest
type RemoteMsgServerRequest struct {
	MsgServerUrl []string `json:"msg_server_url"`
}

// RemoteMsgServerResponse
type RemoteMsgServerResponse struct {
	MsgServerUrl []string `json:"msg_server_url"`
}
