package application

import (
	"context"
	"fadacontrol/internal/base/bootstrap"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
	"os"
	"path/filepath"
	"runtime"
)

type DesktopServiceApp struct {
	db    *conf.DatabaseConf
	ctx   context.Context
	root  *bootstrap.DesktopMasterServiceBootstrap
	debug bool
}

func NewDesktopServiceApp(ctx context.Context, db *conf.DatabaseConf, root *bootstrap.DesktopMasterServiceBootstrap) *DesktopServiceApp {
	return &DesktopServiceApp{db: db, root: root, ctx: ctx}
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

	configPath, err := c.ReadConfigFromYml(filepath.Join(workDir, "config.yml"))
	if err != nil {
		configPath, err = c.ReadConfigFromYml("config.yml")
		if err != nil {
			logger.Info("no config file found,use default config")
		}

	}
	c.SetPath(configPath)
	c.Debug = c.Debug || debug
	if c.Debug {
		c.LogLevel = "debug"
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, constants.CancelFuncKey, cancel)
	ctx = context.WithValue(ctx, constants.ConfKey, c)

	logger.InitLog(ctx)
	utils.NetworkChangeCallbackInit()
	dbFile := filepath.Join(workDir, "data", "config.db")
	dbFile, err = filepath.Abs(dbFile)
	if err != nil {
		logger.Errorf("get db file err %v", err)
		return
	}
	dbFile, err = filepath.Abs(dbFile)
	if err != nil {
		logger.Errorf("get db file err %v", err)
		return
	}

	if !utils.FileExists(dbFile) {
		if err := os.MkdirAll(filepath.Dir(dbFile), os.ModePerm); err != nil {
			logger.Errorf("create db file err %v", err)
			return
		}

		_, err = os.Create(dbFile)
		if err != nil {
			logger.Errorf("create db file err %v", err)
			return
		}
	}

	connection := "file:" + dbFile + "?cache=shared&mode=rwc&_journal_mode=WAL"

	app, err := initDesktopServiceApplication(ctx, &conf.DatabaseConf{Driver: "sqlite", Connection: connection, MaxIdleConnection: 10, MaxOpenConnection: 100, Debug: c.Debug})
	if err != nil {
		logger.Fatal("init desktop service err %v", err)
		return
	}
	appDesktopService = app
	app.Start()
}
func StopDesktopService() {
	if appDesktopService != nil {
		appDesktopService.Stop()
	}

}
