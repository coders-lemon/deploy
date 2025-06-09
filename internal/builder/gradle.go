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

// GradleBuilder Gradle 构建器
type GradleBuilder struct {
	config  *config.Config
	options *BuildOptions
}

// NewGradleBuilder 创建 Gradle 构建器
func NewGradleBuilder(config *config.Config, options *BuildOptions) *GradleBuilder {
	return &GradleBuilder{
		config:  config,
		options: options,
	}
}

// GetType 获取构建器类型
func (g *GradleBuilder) GetType() detector.ProjectType {
	return detector.ProjectTypeGradle
}

// Validate 验证构建环境
func (g *GradleBuilder) Validate() error {
	// 检查 Java 是否安装
	if err := g.checkJava(); err != nil {
		return fmt.Errorf("Java 环境检查失败: %w", err)
	}

	// 检查 Gradle 是否可用
	if err := g.checkGradle(); err != nil {
		return fmt.Errorf("Gradle 环境检查失败: %w", err)
	}

	// 检查 build.gradle 是否存在
	buildGradlePath := filepath.Join(g.options.ProjectPath, "build.gradle")
	buildGradleKtsPath := filepath.Join(g.options.ProjectPath, "build.gradle.kts")

	if _, err := os.Stat(buildGradlePath); err != nil {
		if _, err := os.Stat(buildGradleKtsPath); err != nil {
			return fmt.Errorf("build.gradle 或 build.gradle.kts 不存在")
		}
	}

	return nil
}

// Build 执行构建
func (g *GradleBuilder) Build() (*BuildResult, error) {
	startTime := time.Now()

	fmt.Println("🚀 开始构建 Gradle 项目...")

	// 切换到项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(g.options.ProjectPath); err != nil {
		return nil, fmt.Errorf("切换到项目目录失败: %w", err)
	}

	// 执行 Gradle 构建
	if err := g.runGradleBuild(); err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("Gradle 构建失败: %v", err),
		}, err
	}

	// 查找并打包构建产物
	artifactPath, files, size, err := g.packageArtifacts()
	if err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("打包失败: %v", err),
		}, err
	}

	buildTime := time.Since(startTime)

	fmt.Printf("✅ Gradle 项目构建完成，耗时: %v\n", buildTime)
	fmt.Printf("📦 构建产物: %s (%.2f MB)\n", artifactPath, float64(size)/(1024*1024))

	return &BuildResult{
		Success:      true,
		ArtifactPath: artifactPath,
		Version:      g.options.Version,
		BuildTime:    buildTime.String(),
		Files:        files,
		Size:         size,
		Message:      "构建成功",
	}, nil
}

// checkJava 检查 Java 环境
func (g *GradleBuilder) checkJava() error {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Java 未安装或不在 PATH 中")
	}

	version := strings.Split(string(output), "\n")[0]
	fmt.Printf("✓ Java 版本: %s\n", strings.TrimSpace(version))

	// 检查版本是否符合要求
	if g.config.Java.JavaVersion != "" {
		fmt.Printf("  要求版本: %s\n", g.config.Java.JavaVersion)
	}

	return nil
}

// checkGradle 检查 Gradle 环境
func (g *GradleBuilder) checkGradle() error {
	// 优先使用项目本地的 gradlew
	gradlewPath := filepath.Join(g.options.ProjectPath, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		// 确保 gradlew 有执行权限
		if err := os.Chmod(gradlewPath, 0755); err != nil {
			fmt.Printf("⚠️  设置 gradlew 执行权限失败: %v\n", err)
		}

		cmd := exec.Command(gradlewPath, "--version")
		cmd.Dir = g.options.ProjectPath
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Gradle") {
					fmt.Printf("✓ %s (使用项目 gradlew)\n", strings.TrimSpace(line))
					break
				}
			}
			return nil
		}
	}

	// 如果 gradlew 不可用，尝试系统的 gradle
	cmd := exec.Command("gradle", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Gradle 未安装或不在 PATH 中，且项目没有可用的 gradlew")
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Gradle") {
			fmt.Printf("✓ %s (使用系统 gradle)\n", strings.TrimSpace(line))
			break
		}
	}

	return nil
}

// runGradleBuild 执行 Gradle 构建
func (g *GradleBuilder) runGradleBuild() error {
	fmt.Println("🔨 执行 Gradle 构建...")

	// 确定使用的 Gradle 命令
	gradleCmd := g.getGradleCommand()

	// 构建任务
	tasks := []string{"clean", "build"}
	if g.options.SkipTests {
		tasks = []string{"clean", "build", "-x", "test"}
	}

	// 从配置中获取自定义构建命令
	if g.config.Java.BuildCommand != "" && g.config.Java.BuildTool == "gradle" {
		parts := strings.Fields(g.config.Java.BuildCommand)
		if len(parts) > 1 {
			tasks = parts[1:] // 跳过 gradle/gradlew 命令本身
		}
	}

	cmd := exec.Command(gradleCmd, tasks...)
	cmd.Dir = g.options.ProjectPath

	if g.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 %s %s 失败: %w", gradleCmd, strings.Join(tasks, " "), err)
	}

	fmt.Println("✓ Gradle 构建完成")
	return nil
}

// getGradleCommand 获取 Gradle 命令
func (g *GradleBuilder) getGradleCommand() string {
	// 优先使用项目本地的 gradlew
	gradlewPath := filepath.Join(g.options.ProjectPath, "gradlew")
	if _, err := os.Stat(gradlewPath); err == nil {
		return gradlewPath
	}

	// 使用系统的 gradle
	return "gradle"
}

// packageArtifacts 打包构建产物
func (g *GradleBuilder) packageArtifacts() (string, []string, int64, error) {
	fmt.Println("📦 查找并打包构建产物...")

	// 查找 JAR 文件
	jarFiles, err := g.findJarFiles()
	if err != nil {
		return "", nil, 0, fmt.Errorf("查找 JAR 文件失败: %w", err)
	}

	if len(jarFiles) == 0 {
		return "", nil, 0, fmt.Errorf("未找到 JAR 文件")
	}

	// 选择主要的 JAR 文件
	mainJar := g.selectMainJar(jarFiles)

	// 创建输出目录
	outputDir := g.options.OutputPath
	if outputDir == "" {
		outputDir = "./build"
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", nil, 0, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成版本号
	version := g.options.Version
	if version == "" {
		version = time.Now().Format("20060102-150405")
	}

	// 复制 JAR 文件到输出目录
	artifactName := fmt.Sprintf("%s-%s.jar", g.config.Project.Name, version)
	artifactPath := filepath.Join(outputDir, artifactName)

	if err := g.copyFile(mainJar, artifactPath); err != nil {
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
func (g *GradleBuilder) findJarFiles() ([]string, error) {
	buildDir := "build/libs"

	var jarFiles []string

	err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
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
func (g *GradleBuilder) selectMainJar(jarFiles []string) string {
	// 优先选择不包含 sources、javadoc、tests 的 JAR 文件
	for _, jar := range jarFiles {
		name := filepath.Base(jar)
		if !strings.Contains(name, "sources") &&
			!strings.Contains(name, "javadoc") &&
			!strings.Contains(name, "tests") &&
			!strings.Contains(name, "plain") { // Gradle 有时会生成 plain JAR
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
func (g *GradleBuilder) copyFile(src, dst string) error {
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
