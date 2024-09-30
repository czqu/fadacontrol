package cmd

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/goroutine"
	"fadacontrol/pkg/sys"
)

type program struct {
}

func (p *program) Start(s sys.Svc) error {
	logger.Info("start service")
	goroutine.RecoverGO(p.run)
	return nil
}
func (p *program) run() {
	DesktopServiceMain(false, conf.ServiceMode, workDir)
}
func (p *program) Stop(s sys.Svc) error {
	logger.Info("stop service")
	StopDesktopService()
	return nil
}
func StartService() {
	var p program
	s, _ := sys.New(&p)
	err := s.Run()
	if err != nil {
		logger.Error(err.Error())
	}
}
func InstallService(args ...string) {

	var p program
	s, _ := sys.New(&p)

	err := s.Install(args...)
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
