package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 主配置结构
type Config struct {
	Project      ProjectConfig                `yaml:"project"`
	NPM          NPMConfig                    `yaml:"npm"`
	Java         JavaConfig                   `yaml:"java"`
	Scripts      ScriptsConfig                `yaml:"scripts"`
	Environments map[string]EnvironmentConfig `yaml:"environments"`
	Deploy       DeployConfig                 `yaml:"deploy"`
}

// ProjectConfig 项目配置
type ProjectConfig struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // auto, npm, maven, gradle
}

// NPMConfig NPM项目配置
type NPMConfig struct {
	BuildCommand        string `yaml:"build_command"`
	BuildDir            string `yaml:"build_dir"`
	InstallCommand      string `yaml:"install_command"`
	NodeVersion         string `yaml:"node_version"`
	DefaultStartCommand string `yaml:"default_start_command"`
	DefaultStopCommand  string `yaml:"default_stop_command"`
}

// JavaConfig Java项目配置
type JavaConfig struct {
	BuildTool            string      `yaml:"build_tool"`
	BuildCommand         string      `yaml:"build_command"`
	ArtifactPath         string      `yaml:"artifact_path"`
	JavaVersion          string      `yaml:"java_version"`
	Runtime              JavaRuntime `yaml:"runtime"`
	DefaultStartCommand  string      `yaml:"default_start_command"`
	DefaultStopCommand   string      `yaml:"default_stop_command"`
	DefaultStatusCommand string      `yaml:"default_status_command"`
}

// JavaRuntime Java运行时配置
type JavaRuntime struct {
	HeapSize   HeapSize `yaml:"heap_size"`
	JvmOptions []string `yaml:"jvm_options"`
	AppOptions []string `yaml:"app_options"`
}

// HeapSize 堆内存配置
type HeapSize struct {
	Min string `yaml:"min"`
	Max string `yaml:"max"`
}

// ScriptsConfig 脚本配置
type ScriptsConfig struct {
	Global GlobalScriptConfig `yaml:"global"`
	Custom map[string]string  `yaml:"custom"`
	Hooks  map[string]string  `yaml:"hooks"`
}

// GlobalScriptConfig 全局脚本配置
type GlobalScriptConfig struct {
	Timeout    int    `yaml:"timeout"`
	Shell      string `yaml:"shell"`
	WorkingDir string `yaml:"working_dir"`
}

// EnvironmentConfig 环境配置
type EnvironmentConfig struct {
	Servers        []ServerConfig     `yaml:"servers"`
	DeployPath     string             `yaml:"deploy_path"`
	ServiceName    string             `yaml:"service_name"`
	ServicePort    int                `yaml:"service_port"`
	HealthCheckURL string             `yaml:"health_check_url"`
	Scripts        EnvironmentScripts `yaml:"scripts"`
	Java           *JavaConfig        `yaml:"java,omitempty"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host    string `yaml:"host"`
	User    string `yaml:"user"`
	Port    int    `yaml:"port"`
	KeyFile string `yaml:"key_file"`
}

// EnvironmentScripts 环境脚本配置
type EnvironmentScripts struct {
	Deploy    string            `yaml:"deploy"`
	Variables map[string]string `yaml:"variables"`
}

// DeployConfig 部署配置
type DeployConfig struct {
	BackupCount        int  `yaml:"backup_count"`
	Timeout            int  `yaml:"timeout"`
	RestartDelay       int  `yaml:"restart_delay"`
	HealthCheckTimeout int  `yaml:"health_check_timeout"`
	UseScripts         bool `yaml:"use_scripts"`
	FallbackToDefault  bool `yaml:"fallback_to_default"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "deploy.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			Name: "my-app",
			Type: "auto",
		},
		NPM: NPMConfig{
			BuildCommand:        "npm run build",
			BuildDir:            "dist",
			InstallCommand:      "npm ci",
			NodeVersion:         "18",
			DefaultStartCommand: "pm2 restart ecosystem.config.js",
			DefaultStopCommand:  "pm2 stop my-app",
		},
		Java: JavaConfig{
			BuildTool:    "maven",
			BuildCommand: "mvn clean package -DskipTests",
			ArtifactPath: "target/*.jar",
			JavaVersion:  "11",
			Runtime: JavaRuntime{
				HeapSize: HeapSize{
					Min: "512m",
					Max: "2g",
				},
				JvmOptions: []string{
					"-XX:+UseG1GC",
					"-XX:+HeapDumpOnOutOfMemoryError",
					"-XX:HeapDumpPath=/opt/app/logs",
					"-Dfile.encoding=UTF-8",
					"-Duser.timezone=Asia/Shanghai",
				},
				AppOptions: []string{
					"--server.port=8080",
					"--spring.profiles.active=prod",
				},
			},
			DefaultStartCommand:  "nohup java -Xms{{.HeapMin}} -Xmx{{.HeapMax}} {{.JvmOptions}} -jar {{.JarFile}} {{.AppOptions}} > {{.LogFile}} 2>&1 & echo $! > {{.PidFile}}",
			DefaultStopCommand:   "kill -TERM $(cat {{.PidFile}}) && rm -f {{.PidFile}}",
			DefaultStatusCommand: "ps -p $(cat {{.PidFile}} 2>/dev/null) > /dev/null 2>&1",
		},
		Scripts: ScriptsConfig{
			Global: GlobalScriptConfig{
				Timeout:    300,
				Shell:      "/bin/bash",
				WorkingDir: "/opt/app",
			},
			Custom: make(map[string]string),
			Hooks:  make(map[string]string),
		},
		Environments: map[string]EnvironmentConfig{
			"dev": {
				Servers: []ServerConfig{
					{
						Host:    "dev.example.com",
						User:    "deploy",
						Port:    22,
						KeyFile: "~/.ssh/id_rsa",
					},
				},
				DeployPath:  "/opt/app",
				ServiceName: "my-app",
				ServicePort: 8080,
			},
		},
		Deploy: DeployConfig{
			BackupCount:        5,
			Timeout:            300,
			RestartDelay:       10,
			HealthCheckTimeout: 60,
			UseScripts:         true,
			FallbackToDefault:  true,
		},
	}
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = "deploy.yaml"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	viper.Set("project", config.Project)
	viper.Set("npm", config.NPM)
	viper.Set("java", config.Java)
	viper.Set("scripts", config.Scripts)
	viper.Set("environments", config.Environments)
	viper.Set("deploy", config.Deploy)

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}
