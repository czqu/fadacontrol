package schema

type UpdateInfoResponse struct {
	Release UpdateInfo `json:"release"`
	Beta    UpdateInfo `json:"beta"`
	Dev     UpdateInfo `json:"dev"`
	Canary  UpdateInfo `json:"canary"`
}
type UpdateInfo struct {
	Version     string   `json:"version"`
	VersionCode int      `json:"version_code"`
	UpdateURL   string   `json:"update_url"`
	Mandatory   bool     `json:"mandatory"`
	ReleaseNote []string `json:"release_note"`
}

type ReleaseNotes struct {
	Change []string `json:"change"`
}

type UpdateInfoClientResp struct {
	CanUpdate   bool     `json:"can_update"`
	Channel     string   `json:"channel"`
	Version     string   `json:"version"`
	VersionCode int      `json:"version_code"`
	UpdateURL   string   `json:"update_url"`
	Mandatory   bool     `json:"mandatory"`
	ReleaseNote []string `json:"release_note"`
}
