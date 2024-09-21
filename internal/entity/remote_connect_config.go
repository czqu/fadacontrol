package entity

import "gorm.io/gorm"

// RemoteConnectConfig represents the remote connection configuration
type RemoteConnectConfig struct {
	gorm.Model
	Enable         bool   `gorm:"not null;default:false"`
	ClientId       string `gorm:"not null;uniqueIndex:idx_remote_client_id;default:''"`
	SecurityKey    string `gorm:"not null;default:''" json:"-"`
	Token          string `gorm:"not null;default:''" json:"-"`
	TimeStampCheck bool   `gorm:"not null;default:false"`
	ApiServerUrl   string `gorm:"not null;uniqueIndex:idx_remote_api_server_url;default:''"`
}
type RemoteMsgServer struct {
	gorm.Model
	MsgServerUrl          string `gorm:"not null;uniqueIndex:idx_remote_msg_server_url;default:''"`
	RemoteConnectConfigId uint   `gorm:"not null;"`
}
