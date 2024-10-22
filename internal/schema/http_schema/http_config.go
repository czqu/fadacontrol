package http_schema

type HttpConfigRequest struct {
	Enable bool   `json:"enable"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}
type HttpsConfigRequest struct {
	Enable      bool   `json:"enable"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Cer         string `json:"cer"`
	Key         string `json:"key"`
	EnableHttp3 bool   `json:"enable_http3"`
}
type HttpConfigResponse struct {
	Enable bool   `json:"enable"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}
type HttpsConfigResponse struct {
	Enable      bool   `json:"enable"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Cer         string `json:"cer"`
	Key         string `json:"key"`
	EnableHttp3 bool   `json:"enable_http3"`
}
