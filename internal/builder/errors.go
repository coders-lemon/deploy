package builder

import "errors"

var (
	// ErrUnsupportedProjectType 不支持的项目类型
	ErrUnsupportedProjectType = errors.New("不支持的项目类型")

	// ErrBuildFailed 构建失败
	ErrBuildFailed = errors.New("构建失败")

	// ErrValidationFailed 验证失败
	ErrValidationFailed = errors.New("构建环境验证失败")

	// ErrArtifactNotFound 构建产物未找到
	ErrArtifactNotFound = errors.New("构建产物未找到")
)
