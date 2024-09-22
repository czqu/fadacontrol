package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/router/admin_router"
	"fadacontrol/internal/router/common_router"
	"fadacontrol/internal/service/http_service"
)

type HttpBootstrap struct {
	_common *common_router.CommonRouter
	_adr    *admin_router.AdminRouter
	_conf   *conf.Conf
	_http   *http_service.HttpService
}

func NewHttpBootstrap(_conf *conf.Conf, _http *http_service.HttpService, _common *common_router.CommonRouter, _adr *admin_router.AdminRouter) *HttpBootstrap {
	return &HttpBootstrap{_conf: _conf, _http: _http, _common: _common, _adr: _adr}
}

func (s *HttpBootstrap) Start() error {

	go s._http.StartServer(s._common, HttpServiceApi)
	go s._http.StartServer(s._common, HttpsServiceApi)
	go s._http.StartServer(s._adr, HttpServiceAdmin)
	return nil

}
func (s *HttpBootstrap) Stop() error {
	return s._http.StopAllServer()

}
