package update_service

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/utils"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type UpdateService struct {
	_db *gorm.DB
}

func NewUpdateService(_db *gorm.DB) *UpdateService {
	return &UpdateService{_db}
}
func (u *UpdateService) SetRegion(region version.ProductRegion) error {
	var config entity.SysConfig

	if err := u._db.First(&config).Error; err != nil {
		return err
	}
	config.Region = int(region)
	return u._db.Save(&config).Error
}
func (u *UpdateService) GetRegion() string {
	var config entity.SysConfig

	if err := u._db.First(&config).Error; err != nil {
		return version.RegionGlobal.String()
	}
	return version.ProductRegion(config.Region).String()
}

const updateUrl = "https://update.czqu.net/"

func (u *UpdateService) GetI18nInfo() *schema.I18nInfo {
	config := entity.SysConfig{}
	lang := conf.LanguageEnglish.String()
	if err := u._db.First(&config).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		//todo
	}
	region := u.GetRegion()

	return &schema.I18nInfo{
		Region:   region,
		Language: lang,
	}

}

func (u *UpdateService) CheckUpdate(lang string) (*schema.UpdateInfoClientResp, error) {

	client, err := utils.NewClientBuilder().SetTimeout(5 * time.Second).Build()
	if err != nil {
		return nil, err
	}
	url := updateUrl
	config := entity.SysConfig{}

	region := version.RegionGlobal
	productLang := conf.ProductLanguageFromString(lang)
	if err := u._db.First(&config).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		region = version.GetRegionFromCode(config.Region)
	}

	url = url + version.ProductName + "/" + region.String() + "/" + productLang.String() + "/" + "info.json"

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	updateInfoResp := &schema.UpdateInfoResponse{}
	err = json.Unmarshal([]byte(resp), updateInfoResp)
	if err != nil {
		return nil, err
	}

	edition := u.ShouldUpdateEdition(updateInfoResp)

	info := &schema.UpdateInfo{}
	if edition == version.EditionRelease {
		info = &updateInfoResp.Release
	} else if edition == version.EditionBeta {
		info = &updateInfoResp.Beta
	} else if edition == version.EditionDev {
		info = &updateInfoResp.Dev
	} else {
		info = &updateInfoResp.Canary
	}
	versionCode := version.GetVersion()
	ret := &schema.UpdateInfoClientResp{
		CanUpdate:   u.CanUpdate(versionCode, strconv.Itoa(info.VersionCode)),
		Channel:     string(edition),
		Version:     info.Version,
		VersionCode: info.VersionCode,
		UpdateURL:   info.UpdateURL,
		Mandatory:   info.Mandatory,
		ReleaseNote: info.ReleaseNote,
	}
	return ret, nil
}
func (u *UpdateService) CanUpdate(oldVersion, newVersion string) bool {
	if oldVersion == newVersion {
		return false
	}
	if oldVersion == "" {
		return true
	}
	oldVersionCode, err := strconv.Atoi(oldVersion)
	if err != nil {
		return true
	}
	newVersionCode, err := strconv.Atoi(newVersion)
	if err != nil {
		return false
	}
	return newVersionCode > oldVersionCode

}
func (u *UpdateService) ShouldUpdateEdition(info *schema.UpdateInfoResponse) version.ProductEdition {
	nowVersionCode := version.GetVersion()
	nowEdition := version.GetEdition()
	switch nowEdition {

	case version.EditionRelease:
		return version.EditionRelease
	default:
		if u.CanUpdate(nowVersionCode, strconv.Itoa(info.Release.VersionCode)) {
			return version.EditionRelease
		} else {
			return nowEdition
		}
	}

}
