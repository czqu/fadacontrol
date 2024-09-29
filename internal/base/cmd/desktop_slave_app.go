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

type DesktopServiceApp struct {
	_conf *conf.Conf
	db    *conf.DatabaseConf

	logger *logger.Logger
	root   *bootstrap.DesktopMasterServiceBootstrap
	debug  bool
}

func NewDesktopServiceApp(lo *logger.Logger, _conf *conf.Conf, db *conf.DatabaseConf, root *bootstrap.DesktopMasterServiceBootstrap) *DesktopServiceApp {
	return &DesktopServiceApp{logger: lo, _conf: _conf, db: db, root: root}
}
func (app *DesktopServiceApp) Stop() {

	app.root.Stop()
}
func (app *DesktopServiceApp) Start() {

	app.root.Start()
}

var appDesktopService *DesktopServiceApp

func DesktopServiceMain(debug bool, mode conf.StartMode, workDir string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if utils.DirCanWrite(workDir) {
		workDir, _ = filepath.Abs(workDir)
	} else {
		workDir = "./"
	}
	c := &conf.Conf{}
	c.LogName = conf.DefaultMasterLogName
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

	app, _ := initDesktopServiceApplication(c, &conf.DatabaseConf{Driver: "sqlite", Connection: connection, MaxIdleConnection: 10, MaxOpenConnection: 100, Debug: c.Debug})
	appDesktopService = app
	app.Start()
}
func StopDesktopService() {
	if appDesktopService != nil {
		appDesktopService.Stop()
	}

}
