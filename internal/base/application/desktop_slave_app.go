package application

import (
	"context"
	"fadacontrol/internal/base/bootstrap"
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/constants"
	"fadacontrol/internal/base/logger"
	"fadacontrol/pkg/utils"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"os/user"
	"path/filepath"
	"runtime"
)

type DesktopSlaveServiceApp struct {
	ctx    context.Context
	root   *bootstrap.DesktopSlaveServiceBootstrap
	logger *logger.Logger
}

func NewDesktopSlaveServiceApp(lo *logger.Logger, ctx context.Context, root *bootstrap.DesktopSlaveServiceBootstrap) *DesktopSlaveServiceApp {
	return &DesktopSlaveServiceApp{logger: lo, ctx: ctx, root: root}
}
func (app *DesktopSlaveServiceApp) Stop() {

	app.root.Stop()
}
func (app *DesktopSlaveServiceApp) Start() {

	app.root.Start()
}

var appDesktopSlaveApp *DesktopSlaveServiceApp

func DesktopSlaveAppMain(debug bool, mode conf.StartMode, workDir string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if utils.DirCanWrite(workDir) {
		workDir, _ = filepath.Abs(workDir)
	} else {
		workDir = "./"
	}
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	var currentUserName = "default"
	if currentUser != nil {
		currentUserName = base58.Encode([]byte(currentUser.Name))
	}
	c := &conf.Conf{}
	c.LogName = currentUserName + "_" + conf.DefaultSlaveLogName
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

	c.Debug = c.Debug || debug
	if c.Debug {
		c.LogLevel = "debug"
	}
	app, err := initDesktopSlaveApplication(ctx)
	if err != nil {
		logger.Fatal("init desktop service err %v", err)
		return
	}
	appDesktopSlaveApp = app
	app.Start()
}
