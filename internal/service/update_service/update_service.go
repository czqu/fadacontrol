package update_service

import (
	"encoding/json"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/util"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
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
	return version.GetRegionFromCode(config.Region).String()
}

const updateUrl = "https://update.czqu.net/"

func (u *UpdateService) GetI18nInfo() *schema.I18nInfo {
	config := entity.SysConfig{}
	lang := conf.LanguageEnglish.String()
	if err := u._db.First(&config).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		lang = string(conf.ProductLanguageFromString(config.Language))
	}
	region := u.GetRegion()

	return &schema.I18nInfo{
		Region:   region,
		Language: lang,
	}

}
func (u *UpdateService) SetLanguage(language string) error {
	var config entity.SysConfig
	if err := u._db.First(&config).Error; err != nil {
		return err
	}
	config.Language = string(conf.ProductLanguageFromString(language))
	return u._db.Save(&config).Error
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

	if region == version.RegionDev {
		region = version.RegionGlobal
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
		CanUpdate:   u.CanUpdate(versionCode, strconv.Itoa(info.VersionCode), info.Rev, version.GetRev()),
		Channel:     string(edition),
		Version:     info.Version,
		VersionCode: info.VersionCode,
		UpdateURL:   info.UpdateURL,
		Mandatory:   info.Mandatory,
		ReleaseNote: info.ReleaseNote,
		Rev:         info.Rev,
	}
	return ret, nil
}
func (u *UpdateService) CanUpdate(oldVersion, newVersion, newRev, oldRev string) bool {
	if oldVersion == "" {
		return true
	}
	oldVersionCode, err := strconv.Atoi(oldVersion)
	if err != nil {
		return true
	}
	newVersionCode, err := strconv.Atoi(newVersion)
	if err != nil {
		return true
	}
	if newVersionCode == oldVersionCode && newRev != oldRev {
		return true
	}
	return newVersionCode > oldVersionCode

}
func (u *UpdateService) ShouldUpdateEdition(info *schema.UpdateInfoResponse) version.ProductEdition {
	nowVersionCode := version.GetVersion()
	nowEdition := version.GetEdition()
	nowRev := version.GetRev()
	switch nowEdition {

	case version.EditionRelease:
		return version.EditionRelease
	default:
		if u.CanUpdate(nowVersionCode, strconv.Itoa(info.Release.VersionCode), info.Release.Rev, nowRev) {
			return version.EditionRelease
		} else {
			return nowEdition
		}
	}

}

func (u *UpdateService) GetSupportModules() (schema.SupportModule, error) {
	supportModules := schema.SupportModule{}
	supportModules.ModuleName = make([]string, 0)
	util.SupportModulesLock.RLock()
	for k := range util.SupportModulesCache {
		supportModules.ModuleName = append(supportModules.ModuleName, k)
	}
	util.SupportModulesLock.RUnlock()

	config := entity.SysConfig{}

	region := version.RegionGlobal
	if err := u._db.First(&config).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		region = version.GetRegionFromCode(config.Region)
	}

	remoteSupportModules, err := utils.GetRemoteConfig("supported_modules", region, []string{})
	if err == nil {
		if _, ok := remoteSupportModules.([]interface{}); ok {
			util.SupportModulesLock.Lock()
			util.SupportModulesCache = make(map[string]bool)
			for _, v := range remoteSupportModules.([]interface{}) {
				if key, ok := v.(string); ok {
					util.SupportModulesCache[key] = true
				}

			}
			util.SupportModulesLock.Unlock()
			util.SupportModulesLock.RLock()
			supportModules.ModuleName = []string{}
			for k := range util.SupportModulesCache {
				supportModules.ModuleName = append(supportModules.ModuleName, k)
			}
			util.SupportModulesLock.RUnlock()

		}

	}
	return supportModules, err
}
