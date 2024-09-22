package schema

type UpdateInfo struct {
	Version     string `json:"version"`
	VersionCode int    `json:"version_code"`
	UpdateURL   string `json:"update_url"`
	BinaryUrl   string `json:"binary_url"`
	Mandatory   bool   `json:"mandatory"`
}

type Updates struct {
	Release UpdateInfo `json:"release"`
	Beta    UpdateInfo `json:"beta"`
	Dev     UpdateInfo `json:"dev"`
	Canny   UpdateInfo `json:"canny"`
}
type ReleaseNotes struct {
	Change []string `json:"change"`
}
