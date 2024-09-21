package discovery_service

import (
	"fadacontrol/internal/base/exception"
	"fadacontrol/internal/entity"
	"fadacontrol/internal/schema"
	"gorm.io/gorm"
)

type DiscoverService struct {
	db *gorm.DB
}

func NewDiscoverService(db *gorm.DB) *DiscoverService {
	return &DiscoverService{db: db}
}

func (s *DiscoverService) GetDiscoverConfig() (*schema.DiscoverSchema, error) {
	var config entity.DiscoverConfig
	err := s.db.First(&config).Error
	if err != nil {
		return nil, exception.ErrUnknownException
	}
	return &schema.DiscoverSchema{Enabled: config.Enabled}, err
}

func (s *DiscoverService) PatchDiscoverServiceConfig(content map[string]interface{}) error {

	var config entity.DiscoverConfig
	if err := s.db.First(&config).Error; err != nil {
		return exception.ErrResourceNotFound
	}
	if err := s.db.Model(&config).Updates(content).Error; err != nil {
		return err
	}
	return nil

}
