package bootstrap

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/discovery"
	"gorm.io/gorm"
)

type DiscoverBootstrap struct {
	db     *gorm.DB
	config entity.DiscoverConfig
}

func NewDiscoverBootstrap(db *gorm.DB) *DiscoverBootstrap {
	return &DiscoverBootstrap{db: db}
}
func (d *DiscoverBootstrap) Start() error {
	d.initConfig()
	if d.config.Enabled == true {
		logger.Info("starting discovery service")
		discovery.StartBroadcast()
	}

	return nil
}
func (d *DiscoverBootstrap) Stop() error {
	d.initConfig()
	return discovery.StopBroadcast()

}

func (d *DiscoverBootstrap) initConfig() {
	err := d.db.AutoMigrate(&entity.DiscoverConfig{})
	if err != nil {
		logger.Errorf("failed to migrate database")
		return
	}
	var count int64
	d.db.Model(&entity.DiscoverConfig{}).Count(&count)
	if count == 0 {

		d.config = entity.DiscoverConfig{
			Enabled: true,
		}
		d.db.Create(&d.config)

	}
	if err := d.db.First(&d.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
	}
}
