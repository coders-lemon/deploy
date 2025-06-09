package cmd

import (
	"deploy/internal/builder"
	"deploy/internal/config"
	"deploy/internal/detector"
	"deploy/internal/utils"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	buildType   string
	outputPath  string
	version     string
	skipTests   bool
	projectPath string
)

// buildCmd 构建命令
var buildCmd = &cobra.Command{
	Use:   "build [项目路径]",
	Short: "构建项目",
	Long: `构建项目并生成部署包。

支持的项目类型：
- npm: Node.js 项目
- maven: Maven Java 项目  
- gradle: Gradle Java 项目
- auto: 自动检测项目类型

示例：
  deploy build                           # 构建当前目录项目
  deploy build ./my-app                  # 构建指定目录项目
  deploy build --type=npm                # 构建 NPM 项目
  deploy build --type=maven              # 构建 Maven 项目
  deploy build --path=./my-app           # 使用 --path 指定目录
  deploy build --output=./dist           # 指定输出目录
  deploy build --version=1.0.0           # 指定版本号
  deploy build --skip-tests              # 跳过测试`,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().StringVarP(&buildType, "type", "t", "auto", "项目类型 (npm, maven, gradle, auto)")
	buildCmd.Flags().StringVarP(&outputPath, "output", "o", "", "输出目录 (默认为 ./build)")
	buildCmd.Flags().StringVar(&version, "version", "", "版本号 (默认为时间戳)")
	buildCmd.Flags().BoolVar(&skipTests, "skip-tests", false, "跳过测试")
	buildCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "项目路径")
}

// runBuild 执行构建
func runBuild(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 开始构建项目...")

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

	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("加载配置失败，使用默认配置: %v", err))
		cfg = config.GetDefaultConfig()
	}

	// 如果没有指定项目名称，从路径推断
	if cfg.Project.Name == "" || cfg.Project.Name == "my-app" {
		cfg.Project.Name = utils.GetProjectName(absProjectPath)
	}

	// 检测项目类型
	var projectType detector.ProjectType
	if buildType == "auto" {
		projectInfo, err := detector.DetectProject(absProjectPath)
		if err != nil {
			return fmt.Errorf("自动检测项目类型失败: %w", err)
		}
		projectType = projectInfo.Type
		fmt.Printf("🔍 检测到项目类型: %s\n", projectType)
	} else {
		switch buildType {
		case "npm":
			projectType = detector.ProjectTypeNPM
		case "maven":
			projectType = detector.ProjectTypeMaven
		case "gradle":
			projectType = detector.ProjectTypeGradle
		default:
			return fmt.Errorf("不支持的项目类型: %s", buildType)
		}
		fmt.Printf("📋 使用指定项目类型: %s\n", projectType)
	}

	// 创建构建选项
	buildOptions := &builder.BuildOptions{
		ProjectPath: absProjectPath,
		Environment: "build",
		OutputPath:  outputPath,
		Version:     version,
		Verbose:     verbose,
		SkipTests:   skipTests,
	}

	// 执行构建
	result, err := builder.BuildProject(cfg, buildOptions)
	if err != nil {
		utils.PrintError(fmt.Sprintf("构建失败: %v", err))
		return err
	}

	// 显示构建结果
	if result.Success {
		utils.PrintSuccess("构建完成!")
		fmt.Printf("📦 构建产物: %s\n", result.ArtifactPath)
		fmt.Printf("📊 文件大小: %s\n", utils.FormatFileSize(result.Size))
		fmt.Printf("⏱️  构建耗时: %s\n", result.BuildTime)
		if len(result.Files) > 0 {
			fmt.Printf("📁 包含文件: %d 个\n", len(result.Files))
		}
	} else {
		utils.PrintError(fmt.Sprintf("构建失败: %s", result.Message))
		return fmt.Errorf("构建失败")
	}

	return nil
}

// loadConfig 加载配置文件
func loadConfig() (*config.Config, error) {
	configPath := configFile
	if configPath == "" {
		configPath = "deploy.yaml"
	}

	// 如果配置文件不存在，返回错误但不终止程序
	if !utils.FileExists(configPath) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	return config.LoadConfig(configPath)
}
