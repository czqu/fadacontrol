package data

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/entity"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

func NewAdapterByDB(db *gorm.DB) (*gormadapter.Adapter, error) {
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &entity.CasbinRule{})
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return adapter, nil
}
func NewEnforcer(adapter *gormadapter.Adapter) (*casbin.Enforcer, error) {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "(r.sub == p.sub || p.sub == '*') && (r.obj == p.obj || p.obj == '*'||keyMatch2(r.obj, p.obj)) && (r.act == p.act || p.act == '*')")

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, err
	}
	return enforcer, err
}
