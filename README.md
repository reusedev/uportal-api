# UPortal API

UPortal API 是一个基于 Go 语言开发的用户门户系统后端服务，提供用户管理、任务系统、代币管理等功能。

## 功能特性

### 用户系统
- 微信小程序登录
- 第三方登录（支持微信、Apple、Google、Twitter）
- 用户信息管理（昵称、头像等）
- 登录日志记录
- JWT 认证

### 任务系统
- 任务管理（创建、更新、删除、查询）
- 任务完成与奖励发放
- 任务统计与记录
- 任务完成通知

### 代币系统
- 代币余额管理
- 代币消费规则管理
- 代币交易记录
- 基于服务类型的代币消费

### 支付系统
- 微信支付集成
- 充值订单管理
- 退款处理
- 支付记录查询

## 技术栈

- 编程语言：Go 1.21+
- Web 框架：Gin
- 数据库：MySQL 8.0+
- 缓存：Redis 6.0+
- ORM：GORM
- 日志：Zap
- 配置管理：Yaml
- 认证：JWT
- 第三方集成：微信小程序

## 项目结构

```
.
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动
├── internal/              # 内部包
│   ├── app/              # 应用初始化
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   └── router/          # 路由配置
├── pkg/                  # 公共包
│   ├── config/          # 配置管理
│   ├── consts/          # 常量定义
│   ├── errors/          # 错误处理
│   ├── logs/            # 日志处理
│   ├── response/        # 响应处理
│   └── utils/           # 工具函数
├── script/              # 脚本文件
│   └── schema.sql      # 数据库表结构
├── config/             # 配置文件
├── docs/              # 文档
├── go.mod             # Go 模块文件
├── go.sum             # Go 依赖版本锁定
└── README.md         # 项目说明
```

## 快速开始

### 环境要求

- Go 1.21 或更高版本
- MySQL 8.0 或更高版本
- Redis 6.0 或更高版本
- 微信小程序账号（用于登录功能）

### 安装

1. 克隆项目
```bash
git clone https://github.com/reusedev/uportal-api.git
cd uportal-api
```

2. 安装依赖
```bash
go mod download
```

3. 配置数据库
```bash
# 创建数据库
mysql -u root -p < script/schema.sql
```

4. 修改配置
```bash
# 复制配置文件模板
cp config/config.example.yaml config/config.yaml
# 编辑配置文件
vim config/config.yaml
```

配置文件需要设置以下关键项：
```yaml
wechat:
  miniprogram:
    appId: "your-app-id"
    appSecret: "your-app-secret"
```

5. 运行服务
```bash
go run cmd/api/main.go
```

## API 文档

### 认证 API

#### 用户接口

- `POST /api/v1/login` - 微信小程序登录
  ```json
  {
    "code": "string",           // 微信登录code
    "nickname": "string",       // 可选，用户昵称
    "avatar_url": "string",     // 可选，头像URL
    "encrypted_data": "string", // 可选，加密数据
    "iv": "string"             // 可选，加密算法的初始向量
  }
  ```

- `POST /api/v1/third-party-login` - 第三方登录
  ```json
  {
    "provider": "string",       // 登录提供商：wechat/apple/google/twitter
    "provider_user_id": "string", // 提供商用户ID
    "nickname": "string",       // 可选，用户昵称
    "avatar_url": "string"      // 可选，头像URL
  }
  ```

- `PUT /api/v1/update` - 更新用户信息
  ```json
  {
    "nickname": "string",    // 可选，用户昵称
    "avatar_url": "string"   // 可选，头像URL
  }
  ```

### 任务系统 API

#### 管理员接口

- `POST /api/v1/tasks/admin` - 创建任务
- `PUT /api/v1/tasks/admin/:task_id` - 更新任务
- `DELETE /api/v1/tasks/admin/:task_id` - 删除任务
- `GET /api/v1/tasks/admin/:task_id` - 获取任务详情
- `GET /api/v1/tasks/admin` - 获取任务列表
- `GET /api/v1/tasks/admin/statistics/:task_id` - 获取任务统计信息

#### 用户接口

- `GET /api/v1/tasks/available` - 获取可用任务列表
- `POST /api/v1/tasks/complete` - 完成任务
- `GET /api/v1/tasks/records` - 获取用户任务记录
- `GET /api/v1/tasks/statistics` - 获取用户任务统计

### 代币系统 API

#### 管理员接口

- `POST /api/v1/tokens/rules` - 创建代币消费规则
- `PUT /api/v1/tokens/rules/:rule_id` - 更新代币消费规则
- `DELETE /api/v1/tokens/rules/:rule_id` - 删除代币消费规则
- `GET /api/v1/tokens/rules` - 获取代币消费规则列表

#### 用户接口

- `GET /api/v1/tokens/balance` - 获取代币余额
- `GET /api/v1/tokens/records` - 获取代币交易记录

## 开发指南

### 代码规范

- 遵循 [Go 代码规范](https://golang.org/doc/effective_go)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 进行代码检查
- 编写单元测试，保持测试覆盖率

### 提交规范

- feat: 新功能
- fix: 修复问题
- docs: 文档修改
- style: 代码格式修改
- refactor: 代码重构
- test: 测试用例修改
- chore: 其他修改

## 部署

### Docker 部署

1. 构建镜像
```bash
docker build -t uportal-api .
```

2. 运行容器
```bash
docker run -d \
  --name uportal-api \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  uportal-api
```

### 系统服务部署

1. 编译
```bash
go build -o uportal-api cmd/server/main.go
```

2. 创建系统服务
```bash
# 创建服务文件
sudo vim /etc/systemd/system/uportal-api.service

[Unit]
Description=UPortal API Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/uportal-api
ExecStart=/path/to/uportal-api/uportal-api
Restart=always

[Install]
WantedBy=multi-user.target
```

3. 启动服务
```bash
sudo systemctl enable uportal-api
sudo systemctl start uportal-api
```

## 监控与日志

- 使用 Zap 进行日志记录
- 日志文件位于 `logs/` 目录
- 支持日志轮转
- 支持不同级别的日志记录

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License
