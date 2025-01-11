package bluetooth_service

import (
	"context"
	"fadacontrol/internal/service/control_pc"
	"fadacontrol/internal/service/unlock"
	"gorm.io/gorm"
)

type BluetoothService struct {
	co  *control_pc.ControlPCService
	u   *unlock.UnLockService
	ctx context.Context
	db  *gorm.DB
}

func NewBluetoothService(ctx context.Context, db *gorm.DB, co *control_pc.ControlPCService, u *unlock.UnLockService) *BluetoothService {
	return &BluetoothService{co: co, u: u, ctx: ctx, db: db}
}
