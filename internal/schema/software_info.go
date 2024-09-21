package schema

type SoftwareInfo struct {
	LogPath       string          `json:"log_path"`
	LogLevel      string          `json:"log_level"`
	WorkDir       string          `json:"work_dir"`
	Version       string          `json:"version"`
	Edition       string          `json:"edition"`
	ServiceInfo   []ServiceInfo   `json:"service_info"`
	AlgorithmInfo []AlgorithmInfo `json:"algorithm_info"`
	BuildInfo     string          `json:"build_info"`
	AuthorEmail   string          `json:"author_email"`
}
type ServiceInfo struct {
	ServiceName string `json:"service_name"`
}
type AlgorithmInfo struct {
	AlgorithmName string `json:"algorithm_name"`
}
