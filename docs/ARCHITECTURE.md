# Chirp 后端架构与模块说明

本文档概述当前 Chirp 后端的整体架构、关键模块与运行配置，便于新同学快速上手与排障。

## 总览
- 语言/框架：Go 1.20+，Gorilla Mux。
- 架构风格：分层（Domain/Service/Repository/Handler），依赖倒置。
- 运行模式：可切换数据库（MySQL | SQLite）、存储（Local | Aliyun OSS）、短信（Mock | Aliyun SMS）。
- 配置来源：`config.json`（默认） + 环境变量覆盖，优先级：环境变量 > config.json > 默认值。

## 目录结构（关键部分）
```
cmd/server/main.go      # 入口与依赖注入
internal/config        # 配置加载
internal/domain        # 领域模型与仓库接口
internal/service       # 业务逻辑（Auth/Resource/Storage/SMS 调用）
internal/repository    # 数据访问实现（mysql, sqlite）
internal/handler/http  # HTTP 路由与中间件
pkg/logger             # 日志初始化（stdout+logs/）
pkg/sms                # 短信 Sender（Mock/Aliyun）
pkg/limiter            # 简单限流（按手机号）
docs/                  # 文档
scripts/               # 启动/测试/迁移脚本
uploads/               # 本地存储目录（local 模式）
logs/                  # 运行日志（已 .gitignore）
```

## 配置与切换
默认配置文件：`config.json`（可用 `CONFIG_FILE` 指定路径）。关键字段：
- `dbDriver`: `mysql` | `sqlite`
- `dbDSN`: MySQL DSN（mysql 模式）
- `sqlitePath`: SQLite 文件路径（sqlite 模式）
- `storageBackend`: `local` | `oss`
- `uploadDir`: 本地存储目录（local 模式使用）
- `aliyunEndpoint` / `aliyunBucketName` / `aliyunAccessKeyID` / `aliyunAccessKeySecret`（OSS）
- `aliyunSignName` / `aliyunTemplateCode`（短信）
- `jwtSecret`, `port`
环境变量可覆盖同名字段，便于生产注入敏感信息（AccessKey、模板等）。

## 各层职责
- **Domain (`internal/domain`)**：领域模型（User/Resource/Code/Notification*预留*）与仓库接口。无外部依赖。
- **Repository (`internal/repository`)**：
  - MySQL 实现：`mysql/*`，建表在 `mysql/db.go`（包含 `users/resources/notifications/verification_codes`）。
  - SQLite 实现：`sqlite/*`，建表在 `sqlite/db.go`。
- **Service (`internal/service`)**：
  - `auth_service.go`：注册/登录、短信验证码发送与校验、JWT 签发，依赖用户仓库、验证码仓库、短信 Sender、限流。
  - `resource_service.go`：资源上传/下载/审核/查重，依赖资源仓库与存储实现。
  - `storage.go` / `oss_storage.go`：本地与 OSS 存储实现。
- **Handler (`internal/handler/http`)**：
  - 路由与控制器：`user_handler.go`, `resource_handler.go`。
  - 中间件：认证/可选认证/管理员校验，`LoggingMiddleware`（请求日志）、`RecoverMiddleware`（panic 捕获）。
- **Pkg**：
  - `pkg/sms`：ConsoleSender（Mock）与 AliyunSender。
  - `pkg/logger`：日志输出到 stdout+`logs/server-YYYYMMDD-HHMMSS.log`。
  - `pkg/limiter`：按 key 窗口计数限流（短信 1 次/分钟）。

## 运行与脚本
- 启动：`./scripts/run_server.sh`（默认使用 `config.json`，可设 `CONFIG_FILE`）。
- 测试：
  - `scripts/test_api.sh`：MVP 基础流程（注册/登录/匿名上传/列表）。
  - `scripts/test_admin.sh`：管理员流程（需 MySQL；DB_DRIVER!=mysql 时跳过提权与审核）。
  - `scripts/test_oss.sh`：上传并检查响应是否包含 OSS 域名。
- 迁移：`scripts/run_migration.sh`（仅 MySQL 的角色列迁移）。
- 提权：`scripts/promote_admin.sh`（仅 MySQL）。

## 短信通道
- Aliyun 实机：配置 `aliyunAccessKeyID/Secret`、`aliyunSignName`、`aliyunTemplateCode`。启动日志会打印 `Using Aliyun SMS Sender`。
- Mock：当 `aliyunAccessKeyID` 为空时自动回退，日志打印 `Using Console SMS Sender (Mock)`，验证码仅写日志，不下发。
- 模板变量名：代码使用 `{"code":"<验证码>"}`，模板需匹配变量名 `code`。
- 限频：每手机号 1 分钟 1 次（超限返回 500，日志有 `too many requests`）。

## 存储通道
- Local：`storageBackend=local`，文件写入 `uploadDir`，对外 URL `/uploads/<filename>`。
- OSS：`storageBackend=oss`，需配置 Endpoint/Bucket/AK。对外 URL 形如 `https://<bucket>.<endpoint>/<key>`。

## 日志
- 位置：`logs/server-YYYYMMDD-HHMMSS.log`（已加入 .gitignore），同时输出到 stdout。
- 中间件：记录 method/path/status/耗时；panic 记录 stack；短信/OSS 初始化日志可用于排障。

## 已知预留/未启用
- `Notification` 模型与表已建，但当前未在业务中使用（可后续扩展站内通知）。
- User 扩展字段（school/student_id 等）未在接口中使用，前端可忽略。

## 常见排障
- **短信 500**：检查阿里云错误码（日志 `aliyun sms error`），或 AK 被风控（Forbidden）。
- **OSS 未生效**：确认 `storageBackend=oss`，并在启动日志查看是否打印 `Using Aliyun OSS Storage`；若仍返回本地 URL，检查 AK/Endpoint/Bucket 是否为空。
- **登录/认证失败**：确认 `JWT_SECRET` 一致；Header 为 `Authorization: Bearer <token>`。
- **频率限制**：短信接口 1 分钟内重复会被拒绝，日志提示 `too many requests`。

## 安全与提交
- `config.json` 已在 `.gitignore`，不要提交真实密钥。
- 生产环境使用环境变量覆盖敏感配置（AK/Secret/JWT）。
