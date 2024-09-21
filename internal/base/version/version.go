package version

import "time"

var BuildDate string
var Edition string
var version = "240921"

func GetVersion() (string, error) {

	if version != "" {
		return version, nil
	}
	t, err := GetBuildDate()
	if err != nil {
		return "", err
	}
	return t.Format("060102") + GetEdition(), nil
}
func GetBuildDate() (time.Time, error) {
	layout := time.RFC3339
	t, err := time.Parse(layout, BuildDate)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func GetEdition() string {
	switch Edition {
	case "release":
		return "01"
	case "beta":
		return "03"
	case "dev":
		return "05"
	case "canary":
		return "07"
	default:
		return "07"
	}
}
