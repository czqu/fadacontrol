package remote_schema

// RemoteConnectConfigRequest
type RemoteConfigRequest struct {
	Enable      bool   `json:"enable"`
	ApiServerId uint32 `json:"api_server_id"`
}
type RemoteConfigResponse struct {
	Enable      bool   `json:"enable"`
	ApiServerId uint32 `json:"api_server_id"`
}
type RemoteApiServerConfigRequest struct {
	ApiServerUrl         string `json:"api_server_url"`
	AccessKey            string `json:"access_key"`
	AccessSecret         string `json:"access_secret"`
	EnableSignatureCheck bool   `json:"enable_signature_check"`
}

type RemoteApiServerConfigResponse struct {
	Id                   uint32         `json:"id"`
	ApiServerUrl         string         `json:"api_server_url"`
	AccessKey            string         `json:"access_key"`
	AccessSecret         string         `json:"access_secret"`
	Token                string         `json:"token"`
	TokenExpiresAt       int64          `json:"token_expires_at"`
	ClientId             string         `json:"client_id"`
	EnableSignatureCheck bool           `json:"enable_signature_check"`
	MsgServerUrls        []MsgServerUrl `json:"msg_server_urls"`
}
type MsgServerUrl struct {
	Id           uint32 `json:"id"`
	MsgServerUrl string `json:"msg_server_url"`
	Weight       int    `json:"weight"`
	Enable       bool   `json:"enable"`
}
type CredentialResponse struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	SecurityKey  string `json:"security_key"`
}
