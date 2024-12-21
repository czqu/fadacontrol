package service

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/sys"
)

// for install service,do not implement Start,Stop
type program struct {
}

func (p *program) Start(s sys.Svc) error {

	return nil
}
func (p *program) run() {

}
func (p *program) Stop(s sys.Svc) error {

	return nil
}

func InstallService(path string, args ...string) {

	var p program
	s, _ := sys.New(&p)
	err := s.Install(path, args...)
	if err != nil {
		logger.Error(err.Error())
	}
}
func UninstallService() {
	var p program
	s, _ := sys.New(&p)
	err := s.Uninstall()
	if err != nil {
		logger.Error(err.Error())
	}

}
