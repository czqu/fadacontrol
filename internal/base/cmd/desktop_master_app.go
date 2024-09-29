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

type DesktopSlaveServiceApp struct {
	_conf  *conf.Conf
	db     *conf.DatabaseConf
	root   *bootstrap.DesktopSlaveServiceBootstrap
	logger *logger.Logger
}

func NewDesktopSlaveServiceApp(lo *logger.Logger, _conf *conf.Conf, db *conf.DatabaseConf, root *bootstrap.DesktopSlaveServiceBootstrap) *DesktopSlaveServiceApp {
	return &DesktopSlaveServiceApp{logger: lo, _conf: _conf, db: db, root: root}
}
func (app *DesktopSlaveServiceApp) Stop() {

	app.root.Stop()
}
func (app *DesktopSlaveServiceApp) Start() {

	app.root.Start()
}

var appDesktopDaemon *DesktopSlaveServiceApp

func DesktopSlaveAppMain(debug bool, mode conf.StartMode, workDir string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if utils.DirCanWrite(workDir) {
		workDir, _ = filepath.Abs(workDir)
	} else {
		workDir = "./"
	}
	c := &conf.Conf{}
	c.LogName = conf.DefaultSlaveLogName
	c.LogLevel = conf.DefaultLogLevel
	c.Debug = false
	c.StartMode = mode
	c.SetWorkdir(workDir)

	err := c.ReadConfigFromYml(workDir + "/config.yml")
	if err != nil {
		err = c.ReadConfigFromYml("config.yml")
		if err != nil {
			fmt.Println(err)

		}

	}
	c.Debug = c.Debug || debug
	if c.Debug {
		c.LogLevel = "debug"
	}
	logger.InitLog(c)
	dbFile := workDir + "/data/config.db"
	dbFile, err = filepath.Abs(dbFile)
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
	app, _ := initDesktopDaemonApplication(c, &conf.DatabaseConf{Driver: "sqlite", Connection: connection, MaxIdleConnection: 10, MaxOpenConnection: 100, Debug: c.Debug})
	appDesktopDaemon = app
	app.Start()
}
func StopDesktopDaemon() {
	if appDesktopDaemon != nil {
		appDesktopDaemon.Stop()
	}

}
