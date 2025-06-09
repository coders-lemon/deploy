package cmd

import (
	"deploy/internal/config"
	"deploy/internal/detector"
	"deploy/internal/utils"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	force bool
)

// initCmd 初始化命令
var initCmd = &cobra.Command{
	Use:   "init [项目路径]",
	Short: "初始化配置文件",
	Long: `在当前目录或指定目录创建 deploy.yaml 配置文件。

如果检测到项目类型，会自动生成相应的配置。

示例：
  deploy init                      # 在当前目录初始化
  deploy init ./my-app             # 在指定目录初始化
  deploy init --path=./my-app      # 使用 --path 指定目录
  deploy init --force              # 强制覆盖已存在的配置文件`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "项目路径")
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "强制覆盖已存在的配置文件")
}

// runInit 执行初始化
func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 初始化配置文件...")

	// 如果有位置参数，使用第一个参数作为项目路径
	if len(args) > 0 {
		projectPath = args[0]
	}

	// 检查项目路径
	if !utils.DirExists(projectPath) {
		return fmt.Errorf("项目路径不存在: %s", projectPath)
	}

	// 获取绝对路径
	absProjectPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("获取项目绝对路径失败: %w", err)
	}

	fmt.Printf("📁 项目路径: %s\n", absProjectPath)

	// 配置文件路径
	configPath := filepath.Join(absProjectPath, "deploy.yaml")

	// 检查配置文件是否已存在
	if utils.FileExists(configPath) && !force {
		return fmt.Errorf("配置文件已存在: %s\n使用 --force 参数强制覆盖", configPath)
	}

	// 检测项目类型
	fmt.Println("🔍 检测项目类型...")
	projectInfo, err := detector.DetectProject(absProjectPath)
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("无法检测项目类型: %v", err))
		utils.PrintInfo("将使用默认配置")
	} else {
		utils.PrintSuccess(fmt.Sprintf("检测到项目类型: %s", projectInfo.Type))
	}

	// 创建配置
	cfg := config.GetDefaultConfig()

	// 根据检测结果调整配置
	if projectInfo != nil {
		cfg.Project.Name = projectInfo.Name
		cfg.Project.Type = string(projectInfo.Type)

		// 根据项目类型调整配置
		switch projectInfo.Type {
		case detector.ProjectTypeNPM:
			fmt.Println("📦 配置 NPM 项目设置...")
			// NPM 配置已经在默认配置中设置好了
		case detector.ProjectTypeMaven:
			fmt.Println("☕ 配置 Maven 项目设置...")
			cfg.Java.BuildTool = "maven"
		case detector.ProjectTypeGradle:
			fmt.Println("🐘 配置 Gradle 项目设置...")
			cfg.Java.BuildTool = "gradle"
			cfg.Java.BuildCommand = "./gradlew clean build"
		}
	} else {
		// 使用目录名作为项目名
		cfg.Project.Name = utils.GetProjectName(absProjectPath)
	}

	// 保存配置文件
	fmt.Println("💾 保存配置文件...")
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("配置文件已创建: %s", configPath))

	// 显示配置摘要
	fmt.Println("\n📋 配置摘要:")
	fmt.Printf("  项目名称: %s\n", cfg.Project.Name)
	fmt.Printf("  项目类型: %s\n", cfg.Project.Type)

	if cfg.Project.Type == "npm" || cfg.Project.Type == "auto" {
		fmt.Printf("  NPM 构建命令: %s\n", cfg.NPM.BuildCommand)
		fmt.Printf("  NPM 构建目录: %s\n", cfg.NPM.BuildDir)
	}

	if cfg.Project.Type == "maven" || cfg.Project.Type == "gradle" || cfg.Project.Type == "auto" {
		fmt.Printf("  Java 构建工具: %s\n", cfg.Java.BuildTool)
		fmt.Printf("  Java 构建命令: %s\n", cfg.Java.BuildCommand)
	}

	// 显示下一步建议
	fmt.Println("\n💡 下一步:")
	fmt.Println("  1. 编辑 deploy.yaml 文件以自定义配置")
	fmt.Println("  2. 运行 'deploy detect' 验证项目检测")
	fmt.Println("  3. 运行 'deploy build' 开始构建项目")

	return nil
}
