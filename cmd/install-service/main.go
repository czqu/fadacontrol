package main

import (
	"fadacontrol/internal/base/logger"
	"fadacontrol/internal/base/service"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var installServiceMode bool
var unInstallServiceMode bool
var workDir string
var rootCmd = &cobra.Command{
	Use:    "fadacontrol comandline",
	Short:  "fadacontrol comandline",
	Hidden: true, Run: func(c *cobra.Command, args []string) {
		if installServiceMode {
			execPath, err := os.Executable()
			if err != nil {
				logger.Error(err)
				return
			}
			dir := filepath.Dir(execPath)
			execPath = filepath.Join(dir, "core-service.exe")
			service.InstallService(execPath, "-s", "-w", workDir)

			return
		}
		if unInstallServiceMode {
			service.UninstallService()
			return
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&installServiceMode, "install", "i", false, "install service")
	rootCmd.PersistentFlags().BoolVarP(&unInstallServiceMode, "uninstall", "u", false, "uninstall service")
	rootCmd.PersistentFlags().StringVarP(&workDir, "workdir", "w", "", "working directory")
}
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
func main() {

	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Sync()
		}
	}()
	Execute()
}
