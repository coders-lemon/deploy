package builder

import (
	"deploy/internal/config"
	"deploy/internal/detector"
)

// Builder 构建器接口
type Builder interface {
	// Build 执行构建
	Build() (*BuildResult, error)
	// GetType 获取构建器类型
	GetType() detector.ProjectType
	// Validate 验证构建环境
	Validate() error
}

// BuildResult 构建结果
type BuildResult struct {
	Success      bool     `json:"success"`
	ArtifactPath string   `json:"artifact_path"`
	Version      string   `json:"version"`
	BuildTime    string   `json:"build_time"`
	Files        []string `json:"files"`
	Size         int64    `json:"size"`
	Message      string   `json:"message"`
}

// BuildOptions 构建选项
type BuildOptions struct {
	ProjectPath string
	Environment string
	OutputPath  string
	Version     string
	Verbose     bool
	SkipTests   bool
}

// NewBuilder 创建构建器
func NewBuilder(projectType detector.ProjectType, config *config.Config, options *BuildOptions) (Builder, error) {
	switch projectType {
	case detector.ProjectTypeNPM:
		return NewNPMBuilder(config, options), nil
	case detector.ProjectTypeMaven:
		return NewMavenBuilder(config, options), nil
	case detector.ProjectTypeGradle:
		return NewGradleBuilder(config, options), nil
	default:
		return nil, ErrUnsupportedProjectType
	}
}

// BuildProject 构建项目的便捷函数
func BuildProject(config *config.Config, options *BuildOptions) (*BuildResult, error) {
	// 检测项目类型
	projectInfo, err := detector.DetectProject(options.ProjectPath)
	if err != nil {
		return nil, err
	}

	// 创建构建器
	builder, err := NewBuilder(projectInfo.Type, config, options)
	if err != nil {
		return nil, err
	}

	// 验证构建环境
	if err := builder.Validate(); err != nil {
		return nil, err
	}

	// 执行构建
	return builder.Build()
}
