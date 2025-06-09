# Deploy

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一个专门为 NPM 和 Java 项目设计的命令行部署工具。

## 🚀 特性

- **多项目类型支持**：自动检测并构建 NPM、Maven、Gradle 项目
- **智能项目识别**：自动检测项目类型或手动指定
- **多环境配置**：支持开发、测试、生产等多环境配置管理
- **自定义部署脚本**：支持自定义部署脚本和默认启动命令
- **Java 运行时优化**：智能配置 JVM 参数和堆大小
- **版本管理**：支持版本标记和快速回滚
- **友好的用户界面**：清晰的进度提示和错误信息

## 📦 安装

### 从源码构建

```bash
git clone <repository-url>
cd deploy
go build -o deploy .
```

### 直接下载

从 [Releases](releases) 页面下载适合您操作系统的预编译二进制文件。

## 🛠️ 使用方法

### 快速开始

1. **初始化配置文件**

```bash
# 在当前目录初始化
./deploy init

# 在指定目录初始化
./deploy init ./my-project

# 强制覆盖已存在的配置
./deploy init --force
```

2. **检测项目类型**

```bash
# 检测当前目录
./deploy detect

# 检测指定目录
./deploy detect ./my-project
```

3. **构建项目**

```bash
# 自动检测并构建
./deploy build

# 构建指定目录的项目
./deploy build ./my-project

# 指定项目类型构建
./deploy build --type maven
./deploy build --type npm
./deploy build --type gradle
```

### 命令详解

#### `deploy init` - 初始化配置

```bash
deploy init [项目路径] [flags]

Flags:
  -f, --force         强制覆盖已存在的配置文件
  -p, --path string   项目路径 (default ".")
```

#### `deploy detect` - 检测项目类型

```bash
deploy detect [项目路径] [flags]

Flags:
  -p, --path string   项目路径 (default ".")
```

支持检测的项目类型：
- **NPM 项目**：检测 `package.json` 文件
- **Maven 项目**：检测 `pom.xml` 文件
- **Gradle 项目**：检测 `build.gradle` 或 `build.gradle.kts` 文件

#### `deploy build` - 构建项目

```bash
deploy build [项目路径] [flags]

Flags:
  -o, --output string    输出目录 (默认为 ./build)
  -p, --path string      项目路径 (default ".")
      --skip-tests       跳过测试
  -t, --type string      项目类型 (npm, maven, gradle, auto) (default "auto")
      --version string   版本号 (默认为时间戳)
```

**构建示例：**

```bash
# NPM 项目构建
./deploy build --type npm --output ./dist

# Maven 项目构建（跳过测试）
./deploy build --type maven --skip-tests

# Gradle 项目构建（指定版本）
./deploy build --type gradle --version 1.0.0

# 详细输出模式
./deploy build --verbose
```

### 全局选项

```bash
Global Flags:
      --config string   配置文件路径 (默认为 deploy.yaml)
  -v, --verbose         显示详细输出
```

## ⚙️ 配置文件

初始化后会生成 `deploy.yaml` 配置文件，包含以下主要配置：

### 项目配置

```yaml
project:
  name: "my-app"
  type: "auto"  # auto, npm, maven, gradle
```

### NPM 项目配置

```yaml
npm:
  build_command: "npm run build"
  build_dir: "dist"
  install_command: "npm ci"
  node_version: "18"
  default_start_command: "pm2 restart ecosystem.config.js"
  default_stop_command: "pm2 stop my-app"
```

### Java 项目配置

```yaml
java:
  build_tool: "maven"  # maven, gradle
  build_command: "mvn clean package -DskipTests"
  artifact_path: "target/*.jar"
  java_version: "11"
  
  # Java 运行时配置
  runtime:
    heap_size:
      min: "512m"
      max: "2g"
    jvm_options:
      - "-XX:+UseG1GC"
      - "-XX:+HeapDumpOnOutOfMemoryError"
      - "-XX:HeapDumpPath=/opt/app/logs"
      - "-Dfile.encoding=UTF-8"
      - "-Duser.timezone=Asia/Shanghai"
    app_options:
      - "--server.port=8080"
      - "--spring.profiles.active=prod"
  
  # 默认启动命令（支持模板变量）
  default_start_command: "nohup java -Xms{{.HeapMin}} -Xmx{{.HeapMax}} {{.JvmOptions}} -jar {{.JarFile}} {{.AppOptions}} > {{.LogFile}} 2>&1 & echo $! > {{.PidFile}}"
```

### 环境配置

```yaml
environments:
  dev:
    servers:
      - host: "dev.example.com"
        user: "deploy"
        port: 22
        key_file: "~/.ssh/id_rsa"
    deploy_path: "/opt/app"
    service_name: "my-app"
    service_port: 8080
    
  prod:
    servers:
      - host: "prod1.example.com"
        user: "deploy"
        port: 22
      - host: "prod2.example.com"
        user: "deploy"
        port: 22
    deploy_path: "/opt/app"
    service_name: "my-app"
    service_port: 8080
    health_check_url: "http://localhost:8080/health"
```

## 🏗️ 项目结构
