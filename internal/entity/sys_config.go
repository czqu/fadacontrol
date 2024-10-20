package entity

import "gorm.io/gorm"

type SysConfig struct {
	gorm.Model
	PowerSavingMode bool   `gorm:"default:false"`
	Region          int    `gorm:"not null;default:0"`
	Language        string `gorm:"not null;default:'en'"`
}
