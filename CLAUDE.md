# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

UPortal API — 基于 Go 的用户门户后端服务，面向微信小程序客户端。提供用户管理、代币经济、任务奖励、支付集成、邀请系统等功能。

- 模块路径: `github.com/reusedev/uportal-api`
- Go 版本: 1.22.1
- 代码注释和错误信息均为中文

## Build & Run

无 Makefile，使用标准 Go 命令。

```bash
# 构建三个独立二进制
go build -o bin/api     ./cmd/api/
go build -o bin/manager ./cmd/manager/
go build -o bin/backend ./cmd/backend/

# 运行（均接受 -config 参数，默认 config/config.yaml）
go run ./cmd/api/main.go -config config/config.yaml          # 用户端 API
go run ./cmd/manager/main.go -config config/config.yaml      # 管理后台 API
go run ./cmd/backend/main.go -config config/config.yaml      # 后台定时任务（通知处理）

# 数据库迁移（api 和 manager 支持）
go run ./cmd/api/main.go -config config/config.yaml -migrate

# 测试
go test ./...
go test ./pkg/jwt/...          # 单个包
go test -run TestXxx ./pkg/... # 单个测试

# 格式化 & 静态检查
gofmt -w .
go vet ./...
```

## Architecture

三层架构: **Handler → Service → Model**

```
cmd/
  api/       → 用户端 HTTP 服务 (路由前缀 /api)
  manager/   → 管理后台 HTTP 服务 (路由前缀 /admin)
  backend/   → 后台 Worker (cron 每5分钟处理通知)
internal/
  handler/   → Gin 请求处理，参数绑定，调用 service，返回 response
  service/   → 业务逻辑，事务管理，跨服务调用
  model/     → GORM 模型定义 + 数据库/Redis 初始化 + 数据访问
  middleware/→ Auth, AdminAuth, CORS, Logger, Recovery
pkg/
  config/    → YAML 配置加载，全局单例 config.Get()
  errors/    → 统一错误码 + Error 类型
  response/  → 统一 JSON 响应 (Success/Error/ListResponse)
  logs/      → Zap 日志 (Business + DB 两个 logger)
  jwt/       → JWT 生成/解析
  consts/    → 常量定义（用户角色、代币类型、Redis key 前缀等）
  utils/     → 分页、参数提取工具
```

### 关键约定

- **配置**: `config/config.yaml`（从 `config.example.yaml` 复制），通过 `config.Get()` 全局访问
- **响应格式**: 所有接口统一返回 HTTP 200，业务错误通过 JSON body 中的 `code` 区分
  ```json
  {"code": 0, "message": "success", "data": {...}}
  {"code": 1001, "message": "无效的请求参数"}
  ```
- **认证**: JWT Bearer Token，用户 ID 存入 Gin Context `consts.UserId`("user_id")
- **用户 ID**: Snowflake 生成 → Base58 编码 → VARCHAR(13)
- **分布式锁**: Redis SetNX（任务完成防重复）
- **Service 方法**: 第一个参数为 `context.Context`，Handler 传入 `c.Request.Context()`

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| Handler | `*Handler` | `NewAuthHandler(svc)` |
| Service | `*Service` | `NewAuthService(db, wechatSvc)` |
| 路由注册 | `Register*Routes(group, handler)` | `RegisterTaskRoutes(tasks, taskHandler)` |
| 请求体 | `*Request` | `CreateTaskRequest` |

### 错误码范围

| 范围 | 类别 |
|------|------|
| 0 | 成功 |
| 1000-1005 | 系统错误（内部、参数、未授权、禁止、不存在、不可用）|
| 2000-2007 | 用户错误 |
| 3000-3001 | 微信错误 |
| -2, 4001-4003 | 代币错误（余额不足 = -2）|
| 10000-10007 | 底层系统错误（DB、Redis、第三方）|

### Handler 标准模式

```go
func (h *XxxHandler) DoSomething(c *gin.Context) {
    var req DoSomethingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, errors.New(errors.ErrCodeInvalidParams, "无效的请求参数", err))
        return
    }
    result, err := h.service.DoSomething(c.Request.Context(), &req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, result)
}
```

## Key Dependencies

| 库 | 用途 |
|----|------|
| gin | HTTP 框架 |
| gorm + mysql driver | ORM |
| go-redis/v8 | Redis |
| zap + lumberjack | 日志 + 轮转 |
| golang-jwt/v5 | JWT |
| bwmarrin/snowflake | 分布式 ID |
| robfig/cron | 定时任务 |
| wechatpay-go | 微信支付 v3 |
| smartwalle/alipay/v3 | 支付宝 |

## Git Commit Convention

```
feat: 新功能
fix: 修复问题
docs: 文档修改
style: 代码格式
refactor: 重构
test: 测试
chore: 其他
```