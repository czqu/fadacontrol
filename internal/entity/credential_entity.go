package entity

import "gorm.io/gorm"

type Credential struct {
	gorm.Model
	AccessKey    string `gorm:"uniqueIndex:unique_index"`
	AccessSecret string `gorm:"not null;"`
	SecurityKey  string `gorm:"not null;"`
}
