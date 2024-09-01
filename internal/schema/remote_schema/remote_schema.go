package remote_schema

type RemoteConfigDTO struct {
	Enabled  bool   `json:"enabled"`
	Url      string `json:"url"`
	Secret   string `json:"secret"`
	ClientId string `json:"client_id"`
}
