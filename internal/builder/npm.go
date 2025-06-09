package builder

import (
	"archive/tar"
	"compress/gzip"
	"deploy/internal/config"
	"deploy/internal/detector"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NPMBuilder NPM 构建器
type NPMBuilder struct {
	config  *config.Config
	options *BuildOptions
}

// NewNPMBuilder 创建 NPM 构建器
func NewNPMBuilder(config *config.Config, options *BuildOptions) *NPMBuilder {
	return &NPMBuilder{
		config:  config,
		options: options,
	}
}

// GetType 获取构建器类型
func (n *NPMBuilder) GetType() detector.ProjectType {
	return detector.ProjectTypeNPM
}

// Validate 验证构建环境
func (n *NPMBuilder) Validate() error {
	// 检查 Node.js 是否安装
	if err := n.checkNodeJS(); err != nil {
		return fmt.Errorf("Node.js 环境检查失败: %w", err)
	}

	// 检查 npm 是否安装
	if err := n.checkNPM(); err != nil {
		return fmt.Errorf("npm 环境检查失败: %w", err)
	}

	// 检查 package.json 是否存在
	packageJsonPath := filepath.Join(n.options.ProjectPath, "package.json")
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return fmt.Errorf("package.json 不存在: %s", packageJsonPath)
	}

	return nil
}

// Build 执行构建
func (n *NPMBuilder) Build() (*BuildResult, error) {
	startTime := time.Now()

	fmt.Println("🚀 开始构建 NPM 项目...")

	// 切换到项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取当前目录失败: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(n.options.ProjectPath); err != nil {
		return nil, fmt.Errorf("切换到项目目录失败: %w", err)
	}

	// 安装依赖
	if err := n.installDependencies(); err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("安装依赖失败: %v", err),
		}, err
	}

	// 执行构建
	if err := n.runBuild(); err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("构建失败: %v", err),
		}, err
	}

	// 打包构建产物
	artifactPath, files, size, err := n.packageArtifacts()
	if err != nil {
		return &BuildResult{
			Success: false,
			Message: fmt.Sprintf("打包失败: %v", err),
		}, err
	}

	buildTime := time.Since(startTime)

	fmt.Printf("✅ NPM 项目构建完成，耗时: %v\n", buildTime)
	fmt.Printf("📦 构建产物: %s (%.2f MB)\n", artifactPath, float64(size)/(1024*1024))

	return &BuildResult{
		Success:      true,
		ArtifactPath: artifactPath,
		Version:      n.options.Version,
		BuildTime:    buildTime.String(),
		Files:        files,
		Size:         size,
		Message:      "构建成功",
	}, nil
}

// checkNodeJS 检查 Node.js 环境
func (n *NPMBuilder) checkNodeJS() error {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Node.js 未安装或不在 PATH 中")
	}

	version := strings.TrimSpace(string(output))
	fmt.Printf("✓ Node.js 版本: %s\n", version)

	// 可以在这里检查版本是否符合要求
	if n.config.NPM.NodeVersion != "" {
		// 简单的版本检查，实际项目中可能需要更复杂的版本比较
		fmt.Printf("  要求版本: %s\n", n.config.NPM.NodeVersion)
	}

	return nil
}

// checkNPM 检查 npm 环境
func (n *NPMBuilder) checkNPM() error {
	cmd := exec.Command("npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("npm 未安装或不在 PATH 中")
	}

	version := strings.TrimSpace(string(output))
	fmt.Printf("✓ npm 版本: %s\n", version)

	return nil
}

// installDependencies 安装依赖
func (n *NPMBuilder) installDependencies() error {
	fmt.Println("📦 安装依赖...")

	installCmd := n.config.NPM.InstallCommand
	if installCmd == "" {
		installCmd = "npm ci"
	}

	parts := strings.Fields(installCmd)
	cmd := exec.Command(parts[0], parts[1:]...)

	if n.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 %s 失败: %w", installCmd, err)
	}

	fmt.Println("✓ 依赖安装完成")
	return nil
}

// runBuild 执行构建
func (n *NPMBuilder) runBuild() error {
	fmt.Println("🔨 执行构建...")

	buildCmd := n.config.NPM.BuildCommand
	if buildCmd == "" {
		buildCmd = "npm run build"
	}

	parts := strings.Fields(buildCmd)
	cmd := exec.Command(parts[0], parts[1:]...)

	if n.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 %s 失败: %w", buildCmd, err)
	}

	fmt.Println("✓ 构建完成")
	return nil
}

// packageArtifacts 打包构建产物
func (n *NPMBuilder) packageArtifacts() (string, []string, int64, error) {
	fmt.Println("📦 打包构建产物...")

	buildDir := n.config.NPM.BuildDir
	if buildDir == "" {
		buildDir = "dist"
	}

	// 检查构建目录是否存在
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return "", nil, 0, fmt.Errorf("构建目录不存在: %s", buildDir)
	}

	// 创建输出目录
	outputDir := n.options.OutputPath
	if outputDir == "" {
		outputDir = "./build"
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", nil, 0, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成版本号
	version := n.options.Version
	if version == "" {
		version = time.Now().Format("20060102-150405")
	}

	// 创建 tar.gz 文件
	artifactName := fmt.Sprintf("%s-%s.tar.gz", n.config.Project.Name, version)
	artifactPath := filepath.Join(outputDir, artifactName)

	file, err := os.Create(artifactPath)
	if err != nil {
		return "", nil, 0, fmt.Errorf("创建压缩文件失败: %w", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	var files []string
	var totalSize int64

	// 遍历构建目录并添加到 tar 文件
	err = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		files = append(files, relPath)

		// 创建 tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			size, err := io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
			totalSize += size
		}

		return nil
	})

	if err != nil {
		return "", nil, 0, fmt.Errorf("打包文件失败: %w", err)
	}

	// 获取最终文件大小
	if stat, err := os.Stat(artifactPath); err == nil {
		totalSize = stat.Size()
	}

	fmt.Printf("✓ 打包完成: %s\n", artifactPath)
	return artifactPath, files, totalSize, nil
}
