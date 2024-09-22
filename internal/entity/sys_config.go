package entity

import "gorm.io/gorm"

type SysConfig struct {
	gorm.Model
	PowerSavingMode bool `json:"power_saving_mode"`
}
