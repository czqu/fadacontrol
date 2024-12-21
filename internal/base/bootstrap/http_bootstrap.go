package bootstrap

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/router/admin_router"
	"fadacontrol/internal/router/common_router"
	"fadacontrol/internal/service/http_service"
	"fadacontrol/internal/service/jwt_service"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/utils"
	"sync"
	"time"
)

type HttpBootstrap struct {
	_common   *common_router.CommonRouter
	_adr      *admin_router.AdminRouter
	_conf     *conf.Conf
	_http     *http_service.HttpService
	_jwt      *jwt_service.JwtService
	startOnce sync.Once
	stopOnce  sync.Once
}

func NewHttpBootstrap(_jwt *jwt_service.JwtService, _conf *conf.Conf, _http *http_service.HttpService, _common *common_router.CommonRouter, _adr *admin_router.AdminRouter) *HttpBootstrap {
	return &HttpBootstrap{_jwt: _jwt, _conf: _conf, _http: _http, _common: _common, _adr: _adr}
}

func (s *HttpBootstrap) Start() error {

	s.startOnce.Do(func() {

		err := s.killOther()
		if err == nil {
			time.Sleep(5 * time.Second)
		}

		goroutine.RecoverGO(func() {
			s._http.StartServer(s._common, HttpsServiceApi)
		})
		goroutine.RecoverGO(func() {
			s._http.StartServer(s._adr, HttpServiceAdmin)
		})

	})

	return nil

}
func (s *HttpBootstrap) Stop() error {
	s.stopOnce.Do(func() {
		s._http.StopAllServer()
	})
	return nil

}
func (s *HttpBootstrap) killOther() error {
	client, err := utils.NewClientBuilder().SetTimeout(1 * time.Second).Build()
	if err != nil {
		return err
	}
	token, err := s._jwt.GenerateToken("root")
	if err != nil {
		return err
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	_, err = client.Post("http://localhost:2093/admin/api/v1/sys/stop", "accept: application/json", nil, headers)
	return err
}
