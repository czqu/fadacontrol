package bootstrap

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/discovery_service"
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
	d.readConfig()
	if d.config.Enabled == true {
		logger.Info("starting discovery service")
		discovery_service.StartBroadcast()
	}

	return nil
}
func (d *DiscoverBootstrap) Stop() error {

	return discovery_service.StopBroadcast()

}
func (d *DiscoverBootstrap) readConfig() {
	if err := d.db.First(&d.config).Error; err != nil {
		logger.Errorf("failed to find database: %v", err)
	}
}
