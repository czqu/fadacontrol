package application

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/internal/base/logger"
	"github.com/spf13/cobra"
	"os"
)

var serviceMode bool
var workDir string
var slaveMode bool
var debugMode bool
var commonMode bool
var rootPassword string
var rootCmd = &cobra.Command{
	Use:    "fadacontrol comandline",
	Short:  "fadacontrol comandline",
	Hidden: true,

	Run: func(cmd *cobra.Command, args []string) {
		if rootPassword != "" {
			conf.RootPassword = rootPassword
			conf.ResetPassword = true

		}

		if serviceMode {
			if debugMode {
				DesktopServiceMain(debugMode, conf.ServiceMode, workDir)
			} else {
				StartService()
			}

			return
		}

		if slaveMode {
			logger.Info("slave service start")
			DesktopSlaveAppMain(debugMode, conf.SlaveMode, workDir)
			return
		}
		if commonMode {
			DesktopServiceMain(debugMode, conf.CommonMode, workDir)
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
func init() {
	//rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "config file")
	rootCmd.PersistentFlags().BoolVarP(&serviceMode, "service", "s", false, "service mode")
	rootCmd.PersistentFlags().StringVarP(&workDir, "workdir", "w", "", "working directory")
	rootCmd.PersistentFlags().BoolVarP(&slaveMode, "slave", "", false, "slave-mode")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().BoolVarP(&commonMode, "common-mode", "", true, "common mode")
	rootCmd.PersistentFlags().StringVarP(&rootPassword, "root-password", "", "", "reset root password")
	//err := rootCmd.MarkPersistentFlagRequired("config")

}
