//go:build wireinject
// +build wireinject

package application

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/data"
	"fadacontrol/internal/service/update_service"
	"github.com/google/wire"
)

type RfuctApp struct {
	_conf   *conf.Conf
	_dbConf *conf.DatabaseConf
	_up     *update_service.UpdateService
}

func NewRfuctApp(_conf *conf.Conf, _dbConf *conf.DatabaseConf, _up *update_service.UpdateService) *RfuctApp {
	return &RfuctApp{_conf: _conf, _dbConf: _dbConf, _up: _up}
}

func initRfuctApplication(_conf *conf.Conf, _dbConf *conf.DatabaseConf) (*RfuctApp, error) {
	wire.Build(NewRfuctApp, data.NewDB, update_service.NewUpdateService)
	return &RfuctApp{_conf: _conf, _dbConf: _dbConf}, nil
}
