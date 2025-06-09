package detector

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectType 项目类型
type ProjectType string

const (
	ProjectTypeNPM     ProjectType = "npm"
	ProjectTypeMaven   ProjectType = "maven"
	ProjectTypeGradle  ProjectType = "gradle"
	ProjectTypeUnknown ProjectType = "unknown"
)

// ProjectInfo 项目信息
type ProjectInfo struct {
	Type         ProjectType
	Name         string
	Version      string
	BuildCommand string
	ArtifactPath string
}

// DetectProject 检测项目类型
func DetectProject(projectPath string) (*ProjectInfo, error) {
	if projectPath == "" {
		projectPath = "."
	}

	// 检测 NPM 项目
	if info, err := detectNPMProject(projectPath); err == nil {
		return info, nil
	}

	// 检测 Maven 项目
	if info, err := detectMavenProject(projectPath); err == nil {
		return info, nil
	}

	// 检测 Gradle 项目
	if info, err := detectGradleProject(projectPath); err == nil {
		return info, nil
	}

	return &ProjectInfo{
		Type: ProjectTypeUnknown,
		Name: filepath.Base(projectPath),
	}, fmt.Errorf("无法识别项目类型")
}

// detectNPMProject 检测 NPM 项目
func detectNPMProject(projectPath string) (*ProjectInfo, error) {
	packageJsonPath := filepath.Join(projectPath, "package.json")

	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package.json 不存在")
	}

	// 这里可以解析 package.json 获取更多信息
	// 为了简化，先返回基本信息
	return &ProjectInfo{
		Type:         ProjectTypeNPM,
		Name:         filepath.Base(projectPath),
		BuildCommand: "npm run build",
		ArtifactPath: "dist",
	}, nil
}

// detectMavenProject 检测 Maven 项目
func detectMavenProject(projectPath string) (*ProjectInfo, error) {
	pomXmlPath := filepath.Join(projectPath, "pom.xml")

	if _, err := os.Stat(pomXmlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("pom.xml 不存在")
	}

	return &ProjectInfo{
		Type:         ProjectTypeMaven,
		Name:         filepath.Base(projectPath),
		BuildCommand: "mvn clean package -DskipTests",
		ArtifactPath: "target/*.jar",
	}, nil
}

// detectGradleProject 检测 Gradle 项目
func detectGradleProject(projectPath string) (*ProjectInfo, error) {
	buildGradlePath := filepath.Join(projectPath, "build.gradle")
	buildGradleKtsPath := filepath.Join(projectPath, "build.gradle.kts")

	if _, err := os.Stat(buildGradlePath); err != nil {
		if _, err := os.Stat(buildGradleKtsPath); err != nil {
			return nil, fmt.Errorf("build.gradle 或 build.gradle.kts 不存在")
		}
	}

	return &ProjectInfo{
		Type:         ProjectTypeGradle,
		Name:         filepath.Base(projectPath),
		BuildCommand: "./gradlew build",
		ArtifactPath: "build/libs/*.jar",
	}, nil
}

// IsNPMProject 检查是否为 NPM 项目
func IsNPMProject(projectPath string) bool {
	packageJsonPath := filepath.Join(projectPath, "package.json")
	_, err := os.Stat(packageJsonPath)
	return err == nil
}

// IsMavenProject 检查是否为 Maven 项目
func IsMavenProject(projectPath string) bool {
	pomXmlPath := filepath.Join(projectPath, "pom.xml")
	_, err := os.Stat(pomXmlPath)
	return err == nil
}

// IsGradleProject 检查是否为 Gradle 项目
func IsGradleProject(projectPath string) bool {
	buildGradlePath := filepath.Join(projectPath, "build.gradle")
	buildGradleKtsPath := filepath.Join(projectPath, "build.gradle.kts")

	_, err1 := os.Stat(buildGradlePath)
	_, err2 := os.Stat(buildGradleKtsPath)

	return err1 == nil || err2 == nil
}
