package entity

import (
	"gorm.io/gorm"
)

type BluetoothConfig struct {
	gorm.Model
	Enabled bool `gorm:"default:false"`
}
