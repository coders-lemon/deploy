package cmd

import (
	"deploy/internal/detector"
	"deploy/internal/utils"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

// detectCmd 检测命令
var detectCmd = &cobra.Command{
	Use:   "detect [项目路径]",
	Short: "检测项目类型",
	Long: `检测当前目录或指定目录的项目类型。

支持检测：
- NPM 项目 (package.json)
- Maven 项目 (pom.xml)
- Gradle 项目 (build.gradle 或 build.gradle.kts)

示例：
  deploy detect                    # 检测当前目录
  deploy detect ./my-app           # 检测指定目录
  deploy detect --path=./my-app    # 使用 --path 指定目录`,
	RunE: runDetect,
}

func init() {
	detectCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "项目路径")
}

// runDetect 执行检测
func runDetect(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 检测项目类型...")

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

	// 检测项目类型
	projectInfo, err := detector.DetectProject(absProjectPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("检测失败: %v", err))
		return err
	}

	// 显示检测结果
	utils.PrintSuccess(fmt.Sprintf("检测到项目类型: %s", projectInfo.Type))
	fmt.Printf("📋 项目名称: %s\n", projectInfo.Name)

	if projectInfo.BuildCommand != "" {
		fmt.Printf("🔨 默认构建命令: %s\n", projectInfo.BuildCommand)
	}

	if projectInfo.ArtifactPath != "" {
		fmt.Printf("📦 构建产物路径: %s\n", projectInfo.ArtifactPath)
	}

	// 显示详细的检测信息
	fmt.Println("\n📊 详细检测信息:")

	if detector.IsNPMProject(absProjectPath) {
		fmt.Println("  ✓ 发现 package.json - NPM 项目")
	}

	if detector.IsMavenProject(absProjectPath) {
		fmt.Println("  ✓ 发现 pom.xml - Maven 项目")
	}

	if detector.IsGradleProject(absProjectPath) {
		fmt.Println("  ✓ 发现 build.gradle - Gradle 项目")
	}

	// 给出构建建议
	fmt.Println("\n💡 构建建议:")
	switch projectInfo.Type {
	case detector.ProjectTypeNPM:
		fmt.Println("  使用命令: deploy build --type=npm")
	case detector.ProjectTypeMaven:
		fmt.Println("  使用命令: deploy build --type=maven")
	case detector.ProjectTypeGradle:
		fmt.Println("  使用命令: deploy build --type=gradle")
	default:
		fmt.Println("  使用命令: deploy build (自动检测)")
	}

	return nil
}
