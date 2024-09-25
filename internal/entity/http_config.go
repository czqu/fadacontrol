package entity

import "gorm.io/gorm"

type HttpConfig struct {
	gorm.Model
	ServiceName string `gorm:"not null;uniqueIndex:idx_http_service_name"`
	Enable      bool   `gorm:"not null;default:false"`
	Host        string `gorm:"not null;default:'0.0.0.0'"`
	Port        int    `gorm:"not null;uniqueIndex:idx_http_service_port"`
	Cer         string `gorm:"not null;default:''"`
	Key         string `gorm:"not null;default:''"`
	EnableHttp3 bool   `gorm:"not null;default:false"`
}
