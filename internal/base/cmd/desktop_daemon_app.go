package cmd

import (
	"fadacontrol/internal/base/bootstrap"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type DesktopDaemonApp struct {
	_conf  *conf.Conf
	db     *conf.DatabaseConf
	root   *bootstrap.DesktopDaemonBootstrap
	logger *logger.Logger
}

func NewDesktopDaemonApp(lo *logger.Logger, _conf *conf.Conf, db *conf.DatabaseConf, root *bootstrap.DesktopDaemonBootstrap) *DesktopDaemonApp {
	return &DesktopDaemonApp{logger: lo, _conf: _conf, db: db, root: root}
}
func (app *DesktopDaemonApp) Stop() {

	app.root.Stop()
}
func (app *DesktopDaemonApp) Start() {

	app.root.Start()
}

var appDesktopDaemon *DesktopDaemonApp

func DesktopDaemonAppMain(debug bool, mode conf.StartMode, workDir string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if utils.DirCanWrite(workDir) {
		workDir, _ = filepath.Abs(workDir)
	} else {
		workDir = "./"
	}
	c := &conf.Conf{}
	c.LogName = "daemon.log"
	c.LogLevel = "warn"
	c.Debug = false
	c.StartMode = mode
	workDir, _ = filepath.Abs(workDir)
	c.SetWorkdir(workDir)
	err := c.ReadConfigFromYml(workDir + "/config.yml")
	if err != nil {
		err = c.ReadConfigFromYml("config.yml")
		if err != nil {
			fmt.Println(err)

		}

	}
	dbFile := workDir + "/data/config.db"
	dbFile, err = filepath.Localize(dbFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	dbFile, err = filepath.Abs(dbFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !utils.FileExists(dbFile) {
		if err := os.MkdirAll(filepath.Dir(dbFile), os.ModePerm); err != nil {
			fmt.Println(err)
			return
		}

		_, err = os.Create(dbFile)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	connection := "file:" + dbFile + "?cache=shared&mode=rwc&_journal_mode=WAL"

	c.Debug = c.Debug || debug
	if c.Debug {
		c.LogLevel = "debug"
	}
	app, _ := initDesktopDaemonApplication(c, &conf.DatabaseConf{Driver: "sqlite", Connection: connection, MaxIdleConnection: 10, MaxOpenConnection: 100})
	appDesktopDaemon = app
	app.Start()
}
func StopDesktopDaemon() {
	if appDesktopDaemon != nil {
		appDesktopDaemon.Stop()
	}

}
