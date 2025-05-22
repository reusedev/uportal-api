# UPortal API

UPortal API 是一个基于 Go 语言开发的用户门户系统后端服务，提供用户管理、任务系统、代币管理等功能。

## 功能特性

### 用户系统
- 用户注册与登录
- 用户认证与授权
- 用户信息管理
- 登录日志记录

### 任务系统
- 任务管理（创建、更新、删除、查询）
- 任务完成与奖励发放
- 任务统计与记录
- 任务完成通知

### 代币系统
- 代币余额管理
- 代币充值计划
- 代币消费规则
- 代币交易记录

### 支付系统
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

## 项目结构

```
.
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动
├── internal/              # 内部包
│   ├── handler/          # HTTP 处理器
│   ├── middleware/       # 中间件
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   └── router/          # 路由配置
├── pkg/                  # 公共包
│   ├── errors/          # 错误处理
│   ├── logging/         # 日志处理
│   ├── response/        # 响应处理
│   └── utils/           # 工具函数
├── script/              # 脚本文件
│   └── schema.sql      # 数据库表结构
├── configs/            # 配置文件
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
cp configs/config.example.yaml configs/config.yaml
# 编辑配置文件
vim configs/config.yaml
```

5. 运行服务
```bash
go run cmd/server/main.go
```

## API 文档

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

## 联系方式

- 项目维护者：[Your Name]
- 邮箱：[Your Email]
- 项目地址：[GitHub Repository URL] 