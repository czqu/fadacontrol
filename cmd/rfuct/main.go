package main

import (
	"fadacontrol/cmd/rfuct/application"
	"fadacontrol/internal/base/conf"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// 介绍函数
func printIntro(cmd *cobra.Command, args []string) {
	fmt.Println("This is a command-line tool for managing configuration information.")
}

// 配置A函数，接收bool和string参数
func configA(cmd *cobra.Command, args []string) {
	var flagBool bool
	var flagString string
	cmd.Flags().BoolVar(&flagBool, "bool", false, "配置A的布尔值")
	cmd.Flags().StringVar(&flagString, "string", "", "配置A的字符串值")
	cmd.ParseFlags(args)
	fmt.Printf("配置A: 布尔值=%v， 字符串值=%s\n", flagBool, flagString)
}

// 配置B函数，接收bool参数
func configB(cmd *cobra.Command, args []string) {
	var flagBool bool
	cmd.Flags().BoolVar(&flagBool, "bool", false, "配置B的布尔值")
	cmd.ParseFlags(args)
	fmt.Printf("配置B: 布尔值=%v\n", flagBool)
}

// 配置C函数，接收bool参数
func configC(cmd *cobra.Command, args []string) {
	var flagBool bool
	cmd.Flags().BoolVar(&flagBool, "bool", false, "配置C的布尔值")
	cmd.ParseFlags(args)
	fmt.Printf("配置C: 布尔值=%v\n", flagBool)
}

// 关于信息函数
func aboutInfo(cmd *cobra.Command, args []string) {
	fmt.Println("关于信息: 这是一个命令行工具的关于页面，版本 1.0.0")
}

// 检查更新函数
func checkUpdate(cmd *cobra.Command, args []string) {
	fmt.Println("检查更新: 当前是最新版本。")
}

func main() {
	connection := "file:" + "dbFile" + "?cache=shared&mode=rwc&_journal_mode=WAL"
	_, _ = application.InitRfuctApplication(&conf.Conf{}, &conf.DatabaseConf{Driver: "sqlite", Connection: connection, MaxIdleConnection: 10, MaxOpenConnection: 100, Debug: false})
	var rootCmd = &cobra.Command{Use: "rfuct"}

	var cmdIntro = &cobra.Command{
		Use:   "intro",
		Short: "显示工具介绍",
		Run:   printIntro,
	}

	var cmdConfig = &cobra.Command{
		Use:   "config",
		Short: "set config",
	}

	// 创建配置子命令 配置A
	var cmdConfigA = &cobra.Command{
		Use:   "a",
		Short: "config A",
		Run:   configA,
	}

	var cmdConfigB = &cobra.Command{
		Use:   "b",
		Short: "config B",
		Run:   configB,
	}

	var cmdConfigC = &cobra.Command{
		Use:   "c",
		Short: "config C",
		Run:   configC,
	}

	// 创建关于子命令
	var cmdAbout = &cobra.Command{
		Use:   "about",
		Short: "关于工具",
	}

	// 创建子命令 关于信息
	var cmdAboutInfo = &cobra.Command{
		Use:   "info",
		Short: "显示关于信息",
		Run:   aboutInfo,
	}

	// 创建子命令 检查更新
	var cmdCheckUpdate = &cobra.Command{
		Use:   "update",
		Short: "检查工具更新",
		Run:   checkUpdate,
	}

	// 将所有子命令添加到根命令下
	rootCmd.AddCommand(cmdIntro)
	rootCmd.AddCommand(cmdConfig)
	rootCmd.AddCommand(cmdAbout)

	// 将配置子命令添加到配置管理下
	cmdConfig.AddCommand(cmdConfigA)
	cmdConfig.AddCommand(cmdConfigB)
	cmdConfig.AddCommand(cmdConfigC)

	// 将关于信息和检查更新添加到关于下
	cmdAbout.AddCommand(cmdAboutInfo)
	cmdAbout.AddCommand(cmdCheckUpdate)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
