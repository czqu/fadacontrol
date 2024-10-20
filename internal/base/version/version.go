package version

import (
	"time"
)

type ProductRegion int16

const (
	RegionGlobal ProductRegion = iota
	RegionCN
)

var regionNameMap = map[ProductRegion]string{
	RegionGlobal: "global",
	RegionCN:     "cn",
}

type ProductEdition string

const (
	EditionRelease ProductEdition = "release"
	EditionBeta    ProductEdition = "beta"
	EditionDev     ProductEdition = "dev"
	EditionCanary  ProductEdition = "canary"
	EditionNightly ProductEdition = "nightly"
)

func GetRegionName(region ProductRegion) string {
	if name, ok := regionNameMap[region]; ok {
		return name
	} else {
		return "global"
	}
}
func GetRegionFromCode(code int16) ProductRegion {
	switch code {
	case 1:
		return RegionCN
	default:
		return RegionGlobal
	}
}

var ProductName = "fadacontrol"
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
	return t.Format("060102") + GetEditionCode()
}
func GetBuildDate() (time.Time, error) {
	layout := time.RFC3339
	t, err := time.Parse(layout, BuildDate)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func GetEditionCode() string {
	switch Edition {
	case "release":
		return "00"
	case "beta":
		return "03"
	case "dev":
		return "05"
	case "canary":
		return "07"
	case "nightly":
		return "08"

	default:
		return "08"
	}
}
func GetEdition() ProductEdition {
	switch Edition {
	case "release":
		return EditionRelease
	case "beta":
		return EditionBeta
	case "dev":
		return EditionDev
	case "canary":
		return EditionCanary
	case "nightly":
		return EditionNightly
	default:
		return EditionNightly
	}
}
