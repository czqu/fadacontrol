package cmd

import (
	"fadacontrol/internal/base/conf"
	"fadacontrol/pkg/utils"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var serviceMode bool
var installServiceMode bool
var unInstallServiceMode bool
var workDir string
var slaveMode bool
var debugMode bool
var commonMode bool
var rootCmd = &cobra.Command{
	Use:    "fadacontrol comandline",
	Short:  "fadacontrol comandline",
	Hidden: true,

	Run: func(cmd *cobra.Command, args []string) {

		if serviceMode {
			if debugMode {
				DesktopServiceMain(debugMode, conf.ServiceMode, workDir)
			} else {
				StartService()
			}

			return
		}
		if installServiceMode {

			InstallService("-s", "-w", workDir)

			return
		}
		if unInstallServiceMode {
			UninstallService()
			return
		}
		if slaveMode {
			fmt.Println("slave mode")
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
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
func init() {
	//rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "f", "", "config file")
	rootCmd.PersistentFlags().BoolVarP(&serviceMode, "service", "s", false, "service mode")
	rootCmd.PersistentFlags().BoolVarP(&installServiceMode, "install", "i", false, "install service")
	rootCmd.PersistentFlags().BoolVarP(&unInstallServiceMode, "uninstall", "u", false, "uninstall service")
	rootCmd.PersistentFlags().StringVarP(&workDir, "workdir", "w", "", "working directory")
	rootCmd.PersistentFlags().BoolVarP(&slaveMode, "slave", "", false, "slave-mode")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().BoolVarP(&commonMode, "common-mode", "", true, "common mode")
	//err := rootCmd.MarkPersistentFlagRequired("config")

}
func loadConfFile(filepath string) error {
	ret := utils.FileCanRead(filepath)
	if !ret {

	}

	return nil
}
