package entity

import "gorm.io/gorm"

// RemoteConfig represents the remote connection configuration
type RemoteConfig struct {
	gorm.Model
	Enable                bool `gorm:"not null;default:false"`
	DefaultRemoteServerId uint `gorm:"not null;uniqueIndex:idx_remote_default_remote_server_id;default:0"`
}
type RemoteServer struct {
	gorm.Model
	ApiServerUrl         string `gorm:"not null;uniqueIndex:idx_remote_api_server_url1;default:''"`
	AccessKey            string `gorm:"not null;"`
	AccessSecret         string `gorm:"not null;"`
	Token                string `gorm:"not null;"`
	TokenExpiresAt       int64  `gorm:"not null;default:0"`
	ClientId             string `gorm:"not null;uniqueIndex:idx_remote_client_id;default:''"`
	EnableSignatureCheck bool   `gorm:"not null;default:true"`
}
type RemoteRmttServers struct {
	gorm.Model
	MsgServerUrl   string `gorm:"not null;uniqueIndex:idx_remote_msg_server_url;default:''"`
	RemoteServerId uint   `gorm:"not null;"`
	Weight         int    `gorm:"not null;"`
	Enable         bool   `gorm:"not null;default:true"`
}
