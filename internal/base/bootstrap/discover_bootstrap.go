package bootstrap

import (
	"fadacontrol/internal/entity"
	"fadacontrol/internal/service/discovery_service"
)

type DiscoverBootstrap struct {
	config entity.DiscoverConfig
	_dis   *discovery_service.DiscoverService
}

func NewDiscoverBootstrap(_dis *discovery_service.DiscoverService) *DiscoverBootstrap {
	return &DiscoverBootstrap{_dis: _dis}
}
func (d *DiscoverBootstrap) Start() error {
	d._dis.StartService()

	return nil
}
func (d *DiscoverBootstrap) Stop() error {

	return d._dis.StopService()

}
func (d *DiscoverBootstrap) Restart() error {

	return d._dis.RestartService()

}
