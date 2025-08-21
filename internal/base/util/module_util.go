package util

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/version"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/utils"
	"sync"

	"gorm.io/gorm"
)

type ProductModule string

const (
	BluetoothUnlockModule ProductModule = "bluetooth_unlock"
	RemoteUnlockModule    ProductModule = "remote_unlock"
)

func (p ProductModule) String() string {
	return string(p)
}

var (
	SupportModulesCache map[string]bool = make(map[string]bool)
	SupportModulesLock  sync.RWMutex
)

func (p ProductModule) IsSupport() bool {
	SupportModulesLock.RLock()
	defer SupportModulesLock.RUnlock()
	_, ok := SupportModulesCache[p.String()]
	return ok
}
func (p ProductModule) NotSupport() bool {
	return !p.IsSupport()
}

func GetSupportModules(db *gorm.DB) (schema.SupportModule, error) {
	supportModules := schema.SupportModule{}
	supportModules.ModuleName = make([]string, 0)
	SupportModulesLock.RLock()
	for k := range SupportModulesCache {
		supportModules.ModuleName = append(supportModules.ModuleName, k)
	}
	SupportModulesLock.RUnlock()

	config := entity.SysConfig{}

	region := version.RegionGlobal
	if err := db.First(&config).Error; err != nil {
		logger.Errorf("failed to get config %v", err)
	} else {
		region = version.GetRegionFromCode(config.Region)
	}

	remoteSupportModules, err := utils.GetRemoteConfig("supported_modules", region, []string{})
	if err == nil {
		if _, ok := remoteSupportModules.([]interface{}); ok {
			SupportModulesLock.Lock()
			SupportModulesCache = make(map[string]bool)
			for _, v := range remoteSupportModules.([]interface{}) {
				if key, ok := v.(string); ok {
					SupportModulesCache[key] = true
				}

			}
			SupportModulesLock.Unlock()
			SupportModulesLock.RLock()
			supportModules.ModuleName = []string{}
			for k := range SupportModulesCache {
				supportModules.ModuleName = append(supportModules.ModuleName, k)
			}
			SupportModulesLock.RUnlock()
			goroutine.RecoverGO(func() {
				SupportModulesLock.RLock()
				defer SupportModulesLock.RUnlock()
				err := SyncModules(db, SupportModulesCache)
				if err != nil {
					logger.Errorf("failed to sync modules %v", err)
				}
			})

		}

	}

	return supportModules, err
}
func SyncModules(db *gorm.DB, newModulesCache map[string]bool) error {
	return db.Transaction(func(tx *gorm.DB) error {

		var existingModules []entity.SupportModule
		if err := tx.Find(&existingModules).Error; err != nil {
			return err
		}

		existingModulesMap := make(map[string]entity.SupportModule)
		for _, m := range existingModules {
			existingModulesMap[m.ModuleName] = m
		}

		var modulesToDelete []string
		var modulesToInsert []entity.SupportModule

		for moduleName, m := range existingModulesMap {
			if _, ok := newModulesCache[moduleName]; !ok {
				modulesToDelete = append(modulesToDelete, m.ModuleName)
			}
		}

		for moduleName, enabled := range newModulesCache {
			if _, ok := existingModulesMap[moduleName]; !ok && enabled {
				modulesToInsert = append(modulesToInsert, entity.SupportModule{
					ModuleName: moduleName,
				})
			}
		}

		if len(modulesToDelete) > 0 {
			if err := tx.Where("module_name IN ?", modulesToDelete).Delete(&entity.SupportModule{}).Error; err != nil {
				return err
			}
		}

		if len(modulesToInsert) > 0 {
			if err := tx.CreateInBatches(modulesToInsert, 100).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
