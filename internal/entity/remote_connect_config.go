package entity

import "gorm.io/gorm"

// RemoteConnectConfig represents the remote connection configuration
type RemoteConnectConfig struct {
	gorm.Model
	Url            string `gorm:"not null"`
	Enable         bool   `gorm:"not null;default:false"`
	ClientId       string `gorm:"not null;uniqueIndex:idx_remote_client_id"`
	Secret         string `gorm:"not null"`
	Salt           string `gorm:"not null"`
	TimeStampCheck bool   `gorm:"not null;default:false"`
}
