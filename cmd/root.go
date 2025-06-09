package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configFile string
	verbose    bool
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "deploy",
	Short: "一个用于 NPM 和 Java 项目的部署工具",
	Long: `Deploy 是一个专门为 NPM 和 Java 项目设计的命令行部署工具。

它支持：
- NPM 项目的自动构建和打包
- Maven 和 Gradle Java 项目的构建
- 多环境配置管理
- 自定义部署脚本
- 版本管理和回滚

使用示例：
  deploy build                    # 自动检测并构建项目
  deploy build --type=npm         # 指定构建 NPM 项目
  deploy build --type=maven       # 指定构建 Maven 项目
  deploy detect                   # 检测项目类型`,
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "配置文件路径 (默认为 deploy.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细输出")

	// 添加子命令
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(initCmd)
}

// initConfig 初始化配置
func initConfig() {
	if configFile != "" {
		// 使用指定的配置文件
		fmt.Printf("使用配置文件: %s\n", configFile)
	}
}
