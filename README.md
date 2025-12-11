# Chirp (知了) - Backend Service

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Chirp** 是一个面向大学生的资料共享平台后端服务。本项目采用 Golang 开发，旨在提供高效、安全的文件存储与元数据管理服务。

当前版本为 MVP（Minimum Viable Product），实现了核心的用户认证与资源管理功能。

## 架构设计

本项目遵循 **Clean Architecture**（整洁架构）原则，结合 **Standard Go Project Layout** 进行组织。旨在实现高内聚、低耦合，确保业务逻辑独立于外部框架和驱动。

### 分层说明

*   **Domain Layer (`internal/domain`)**: 核心业务实体与接口定义。不依赖任何外部库。
*   **Service Layer (`internal/service`)**: 具体的业务逻辑实现（如认证流程、文件处理）。依赖于 Domain 层接口。
*   **Repository Layer (`internal/repository`)**: 数据持久化适配器。实现了 Domain 层定义的 Repository 接口。当前提供 MySQL 与 SQLite 两种实现，可通过配置切换。
*   **Handler Layer (`internal/handler`)**: 接口适配层。负责处理 HTTP 请求，解析参数并调用 Service 层。
*   **Config (`internal/config`)**: 集中式配置管理。

### 目录结构

```text
/
├── cmd/
│   └── server/
│       └── main.go           # 应用程序入口，负责依赖注入与服务启动
├── internal/
│   ├── config/               # 配置加载与管理
│   ├── domain/               # 领域模型 (User, Resource) 与 接口定义
│   ├── handler/              # HTTP 处理器 (REST API)
│   │   └── http/             # 具体 HTTP Handler 实现与中间件
│   ├── repository/           # 数据访问层实现
│   │   └── sqlite/           # SQLite 实现
│   └── service/              # 业务逻辑层
├── pkg/                      # 公共库 (可被外部项目复用)
│   └── util/                 # 工具函数 (如 Password Hashing)
├── uploads/                  # 本地文件存储目录 (开发环境)
├── go.mod                    # 依赖管理
└── README.md                 # 项目文档
```

## 快速开始 (Getting Started)

### 前置要求

*   **Go**: 1.20 或更高版本
*   **GCC**: Windows 环境下运行 SQLite 驱动需要安装 GCC (推荐 [MinGW-w64](https://www.mingw-w64.org/))。
*   **PowerShell**: 推荐使用 PowerShell 进行脚本运行。

## API 文档 (API Documentation)

详细的 API 接口文档请参考 [docs/API.md](docs/API.md)。

### 接口概览

*   **用户认证**: 注册、登录、短信验证码、获取用户信息
*   **资源管理**: 上传、下载、搜索资源
*   **管理员**: 资源审核、查重

## 本地开发环境搭建

1.  **克隆仓库**

    ```bash
    git clone https://github.com/zuquanzhi/Chirp.git
    cd Chirp
    ```

2.  **配置（推荐 config.json）**

        复制 `config.example.json` 为 `config.json` 并按需修改，程序启动时默认读取当前目录的 `config.json`。也可通过环境变量 `CONFIG_FILE` 指定路径。

        `config.json` 示例字段说明：
        ```json
        {
            "port": "9527",
            "dbDriver": "mysql",            // mysql | sqlite
            "dbDSN": "chirp:test12345@tcp(127.0.0.1:3306)/chirp?parseTime=true&loc=Local",
            "sqlitePath": "chirp.db",
            "jwtSecret": "dev_secret_key",
            "uploadDir": "uploads",
            "storageBackend": "local",      // local | oss
            "aliyunEndpoint": "oss-cn-hangzhou.aliyuncs.com",
            "aliyunAccessKeyID": "your-access-key",
            "aliyunAccessKeySecret": "your-access-secret",
            "aliyunBucketName": "chirp-oss",
            "aliyunSignName": "your-sms-sign",
            "aliyunTemplateCode": "SMS_xxx"
        }
        ```

        环境变量仍可覆盖同名配置（优先级：环境变量 > config.json > 默认值），便于在生产中通过环境覆盖敏感信息。

3.  **安装依赖**

    ```powershell
    go mod tidy
    ```

4.  **启动服务**

    使用提供的脚本一键启动：
    ```powershell
    .\scripts\run_server.ps1
    ```
    或者手动运行：
    ```powershell
    go run ./cmd/server/main.go
    ```

5.  **运行测试**

    使用测试脚本验证 API：
    ```powershell
    .\scripts\test_api.ps1
    ```

## API 文档

详细的 API 接口文档请参考 [docs/API.md](docs/API.md)。

## 技术栈

*   **Language**: Go (Golang)
*   **Web Framework**: Gorilla Mux
*   **Database**: MySQL / SQLite
*   **Storage**: Local Filesystem / Aliyun OSS
*   **Auth**: JWT (JSON Web Tokens)
*   **Password Hashing**: bcrypt

## 开发规范

*   **代码风格**: 遵循 `go fmt` 标准。
*   **错误处理**: 尽量在 Service 层处理业务错误，Handler 层处理 HTTP 状态码映射。
*   **依赖注入**: 严禁在业务逻辑中直接初始化 Repository，必须通过构造函数注入。

## 路线图 (Roadmap)

*   [x] 基础用户认证 (Signup/Login/JWT)
*   [x] 资源上传与下载 (Local Storage)
*   [x] 架构重构 (Clean Architecture)
*   [x] MVP 1.0 功能 (匿名上传/搜索/审核/查重)
*   [x] 接入 MySQL 数据库
*   [x] 接入阿里云 OSS 对象存储
*   [ ] Docker 容器化部署支持
*   [ ] 单元测试覆盖

---
© 2025 Chirp Team. All Rights Reserved.
