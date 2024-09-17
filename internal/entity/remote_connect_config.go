package entity

import "gorm.io/gorm"

// RemoteConnectConfig represents the remote connection configuration
type RemoteConnectConfig struct {
	gorm.Model
	Enable         bool   `gorm:"not null;default:false"`
	ClientId       string `gorm:"not null;uniqueIndex:idx_remote_client_id;default:''"`
	Secret         string `gorm:"not null;default:''"`
	Key            string `gorm:"not null;default:''"`
	Salt           string `gorm:"not null;default:''"`
	TimeStampCheck bool   `gorm:"not null;default:false"`
}
type RemoteServer struct {
	gorm.Model
	MsgServerUrl string `gorm:"not null;uniqueIndex:idx_remote_msg_server_url;default:''"`
	ApiServerUrl string `gorm:"not null;uniqueIndex:idx_remote_api_server_url;default:''"`
}
