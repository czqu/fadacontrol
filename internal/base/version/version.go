package version

import (
	"time"
)

var BuildDate string
var Edition string
var _VersionName string
var version string
var GitCommit string
var AuthorEmail string

func GetBuildInfo() string {
	return GetVersionName() + " " + Edition + " " + "build-" + GetVersion() + "-" + GitCommit

}
func GetVersionName() string {
	return _VersionName
}
func GetVersion() string {

	if version != "" {
		return version
	}
	t, err := GetBuildDate()
	if err != nil {
		return ""
	}
	return t.Format("060102") + GetEdition()
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
		return "09"
	}
}
