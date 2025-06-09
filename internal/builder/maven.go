package builder

import (
	"deploy/internal/config"
	"deploy/internal/detector"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// MavenBuilder Maven 构建器
type MavenBuilder struct {
	config  *config.Config
	options *BuildOptions
}

// NewMavenBuilder 创建 Maven 构建器
func NewMavenBuilder(config *config.Config, options *BuildOptions) *MavenBuilder {
	return &MavenBuilder{
		config:  config,
		options: options,
	}
}

// GetType 获取构建器类型
func (m *MavenBuilder) GetType() detector.ProjectType {
	return detector.ProjectTypeMaven
}

// Validate 验证构建环境
func (m *MavenBuilder) Validate() error {
	// 检查 Java 是否安装
	if err := m.checkJava(); err != nil {
		return fmt.Errorf("Java 环境检查失败: %w", err)
	}

	// 检查 Maven 是否安装
	if err := m.checkMaven(); err != nil {
		return fmt.Errorf("Maven 环境检查失败: %w", err)
	}

	// 检查 pom.xml 是否存在
	pomPath := filepath.Join(m.options.ProjectPath, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("pom.xml 不存在: %s", pomPath)
	}

	return nil
}

// Build 执行构建
func (m *MavenBuilder) Build() (*BuildResult, error) {
	startTime := time.Now()

	fmt.Println("🚀 开始构建 Maven 项目...")

	// 切换到项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(m.options.ProjectPath); err != nil {
		return nil, fmt.Errorf("切换到项目目录失败: %w", err)
	}

	// 执行 Maven 构建
	if err := m.runMavenBuild(); err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("Maven 构建失败: %v", err),
		}, err
	}

	// 查找并打包构建产物
	artifactPath, files, size, err := m.packageArtifacts()
	if err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("打包失败: %v", err),
		}, err
	}

	buildTime := time.Since(startTime)

	fmt.Printf("✅ Maven 项目构建完成，耗时: %v\n", buildTime)
	fmt.Printf("📦 构建产物: %s (%.2f MB)\n", artifactPath, float64(size)/(1024*1024))

	return &BuildResult{
		Success:      true,
		ArtifactPath: artifactPath,
		Version:      m.options.Version,
		BuildTime:    buildTime.String(),
		Files:        files,
		Size:         size,
		Message:      "构建成功",
	}, nil
}

// checkJava 检查 Java 环境
func (m *MavenBuilder) checkJava() error {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Java 未安装或不在 PATH 中")
	}

	version := strings.Split(string(output), "\n")[0]
	fmt.Printf("✓ Java 版本: %s\n", strings.TrimSpace(version))

	// 检查版本是否符合要求
	if m.config.Java.JavaVersion != "" {
		fmt.Printf("  要求版本: %s\n", m.config.Java.JavaVersion)
	}

	return nil
}

// checkMaven 检查 Maven 环境
func (m *MavenBuilder) checkMaven() error {
	cmd := exec.Command("mvn", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Maven 未安装或不在 PATH 中")
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		fmt.Printf("✓ Maven 版本: %s\n", strings.TrimSpace(lines[0]))
	}

	return nil
}

// runMavenBuild 执行 Maven 构建
func (m *MavenBuilder) runMavenBuild() error {
	fmt.Println("🔨 执行 Maven 构建...")

	buildCmd := m.config.Java.BuildCommand
	if buildCmd == "" {
		if m.options.SkipTests {
			buildCmd = "mvn clean package -DskipTests"
		} else {
			buildCmd = "mvn clean package"
		}
	}

	parts := strings.Fields(buildCmd)
	cmd := exec.Command(parts[0], parts[1:]...)

	if m.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 %s 失败: %w", buildCmd, err)
	}

	fmt.Println("✓ Maven 构建完成")
	return nil
}

// packageArtifacts 打包构建产物
func (m *MavenBuilder) packageArtifacts() (string, []string, int64, error) {
	fmt.Println("📦 查找并打包构建产物...")

	// 查找 JAR 文件
	jarFiles, err := m.findJarFiles()
	if err != nil {
		return "", nil, 0, fmt.Errorf("查找 JAR 文件失败: %w", err)
	}

	if len(jarFiles) == 0 {
		return "", nil, 0, fmt.Errorf("未找到 JAR 文件")
	}

	// 选择主要的 JAR 文件（通常是不带 sources 和 javadoc 的）
	mainJar := m.selectMainJar(jarFiles)

	// 创建输出目录
	outputDir := m.options.OutputPath
	if outputDir == "" {
		outputDir = "./build"
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", nil, 0, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成版本号
	version := m.options.Version
	if version == "" {
		version = time.Now().Format("20060102-150405")
	}

	// 复制 JAR 文件到输出目录
	artifactName := fmt.Sprintf("%s-%s.jar", m.config.Project.Name, version)
	artifactPath := filepath.Join(outputDir, artifactName)

	if err := m.copyFile(mainJar, artifactPath); err != nil {
		return "", nil, 0, fmt.Errorf("复制 JAR 文件失败: %w", err)
	}

	// 获取文件信息
	stat, err := os.Stat(artifactPath)
	if err != nil {
		return "", nil, 0, fmt.Errorf("获取文件信息失败: %w", err)
	}

	files := []string{filepath.Base(artifactPath)}
	size := stat.Size()

	fmt.Printf("✓ 打包完成: %s\n", artifactPath)
	return artifactPath, files, size, nil
}

// findJarFiles 查找 JAR 文件
func (m *MavenBuilder) findJarFiles() ([]string, error) {
	targetDir := "target"

	var jarFiles []string

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".jar") {
			jarFiles = append(jarFiles, path)
		}

		return nil
	})

	return jarFiles, err
}

// selectMainJar 选择主要的 JAR 文件
func (m *MavenBuilder) selectMainJar(jarFiles []string) string {
	// 优先选择不包含 sources、javadoc、tests 的 JAR 文件
	for _, jar := range jarFiles {
		name := filepath.Base(jar)
		if !strings.Contains(name, "sources") &&
			!strings.Contains(name, "javadoc") &&
			!strings.Contains(name, "tests") {
			return jar
		}
	}

	// 如果没找到，返回第一个
	if len(jarFiles) > 0 {
		return jarFiles[0]
	}

	return ""
}

// copyFile 复制文件
func (m *MavenBuilder) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
