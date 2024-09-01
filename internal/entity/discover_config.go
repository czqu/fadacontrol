package entity

import "gorm.io/gorm"

type DiscoverConfig struct {
	gorm.Model
	Enabled bool `gorm:"default:true"`
}
