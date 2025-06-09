package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists 检查目录是否存在
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(path string) error {
	if !DirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GetProjectName 从路径获取项目名称
func GetProjectName(projectPath string) string {
	if projectPath == "" || projectPath == "." {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Base(wd)
		}
		return "unknown"
	}
	return filepath.Base(projectPath)
}

// GenerateVersion 生成版本号
func GenerateVersion() string {
	return time.Now().Format("20060102-150405")
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// SanitizeFileName 清理文件名，移除不安全字符
func SanitizeFileName(name string) string {
	// 替换不安全字符
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// IsValidProjectType 检查项目类型是否有效
func IsValidProjectType(projectType string) bool {
	validTypes := []string{"npm", "maven", "gradle", "auto"}
	for _, t := range validTypes {
		if t == projectType {
			return true
		}
	}
	return false
}

// PrintSuccess 打印成功消息
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintError 打印错误消息
func PrintError(message string) {
	fmt.Printf("❌ %s\n", message)
}

// PrintWarning 打印警告消息
func PrintWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// PrintInfo 打印信息消息
func PrintInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}
