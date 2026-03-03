# LinkFlow — 轻量级物联网云平台 | Lightweight IoT Platform

基于 Go + React 的开源物联网（IoT）平台，支持 MQTT 设备接入、实时数据监控、告警、定时任务、语音控制、OTA 升级。

[English](#english) | [中文](#中文)

---
<img width="1920" height="911" alt="1" src="https://github.com/user-attachments/assets/8c79f618-3063-4a56-bb33-a0f42c4ee4ef" />
<img width="1920" height="911" alt="2" src="https://github.com/user-attachments/assets/8ab7a1b3-f29a-447b-9db0-e100659b6caa" />
<img width="1920" height="911" alt="3" src="https://github.com/user-attachments/assets/f645e885-9cff-41c2-bd46-04d0b15ee036" />



<a id="中文"></a>

LinkFlow 是一个基于 Go 的**轻量级物联网云平台**，适合作为学习项目或中小型 IoT 场景的起点。它可以接入模拟/真实 IoT 设备，采集传感器数据，并提供实时监控、告警、语音控制、OTA 升级等功能。

> 本项目最初是一个毕业设计，现已开源。Docker 一键部署，开箱即用。欢迎 Star、Fork、提 Issue 或 PR。

## 功能一览

| 模块 | 说明 |
|------|------|
| 设备管理 | 设备 CRUD、自动生成密钥、物模型绑定（属性/事件/服务） |
| MQTT 接入 | 内嵌 Mochi MQTT v2 Broker，设备认证 + ACL + 物模型校验 |
| 实时数据 | 遥测数据通过 MQTT 上报 → TimescaleDB 时序存储 → WebSocket 实时推送到前端 |
| 数据可视化 | 实时属性卡片 + 历史趋势图表（自动聚合，支持 1h/6h/24h/7d/30d） |
| 告警系统 | 阈值规则、三级告警（info/warning/critical）、冷却机制、WebSocket 实时通知 |
| 定时任务 | Cron 表达式驱动，支持属性设置和服务调用，MQTT 下发 |
| 语音控制 | 本地 NLP 关键词匹配 或 Dify Workflow API，支持开关/数值/枚举/服务调用 |
| OTA 升级 | 固件上传 → MQTT 推送升级命令 → 设备 HTTP 下载 → 进度实时上报 |
| 在线调试 | 无需真实硬件，浏览器内模拟设备上下线、属性下发、服务调用 |
| 安全体系 | JWT 认证 + PostgreSQL RLS 行级安全 + 审计日志 |
| CSV 导出 | 一键导出设备历史遥测数据 |

## 技术架构

```
┌──────────────┐     ┌──────────────────────────────────────────────┐
│   Browser    │────▶│  Nginx (反向代理 + 静态资源)                  │
│  React SPA   │◀────│  :80                                         │
└──────┬───────┘     └──────────┬───────────────────────────────────┘
       │ WebSocket              │ /api/*
       ▼                        ▼
┌──────────────────────────────────────────────────────────────────┐
│                     Go Backend (Gin)  :8080                      │
│  ┌─────────┐ ┌───────────┐ ┌───────────┐ ┌──────────────────┐   │
│  │ JWT Auth │ │ REST API  │ │ WebSocket │ │ Scheduler (Cron) │   │
│  │Middleware│ │ Handlers  │ │   Hub     │ │                  │   │
│  └─────────┘ └───────────┘ └───────────┘ └──────────────────┘   │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │          Mochi MQTT v2 Broker (内嵌)  :1883               │   │
│  │  设备认证 │ ACL │ 物模型校验 │ 语音NLP │ 告警评估          │   │
│  └───────────────────────────────────────────────────────────┘   │
└──────────┬────────────────────────────────┬──────────────────────┘
           │                                │
           ▼                                ▼
┌─────────────────────┐          ┌─────────────────────┐
│ TimescaleDB (PG 15) │          │    Redis 7          │
│ 时序数据 + RLS 行级  │          │ Token + 设备在线状态 │
│ 安全策略             │          │                     │
└─────────────────────┘          └─────────────────────┘
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24+ / Gin / pgx v5（手写 SQL，无 ORM） / Mochi MQTT v2 |
| 前端 | React 18 / TypeScript / shadcn/ui / Tailwind CSS / Vite / recharts |
| 数据库 | PostgreSQL 15 + TimescaleDB（时序扩展） |
| 缓存 | Redis 7 |
| 认证 | JWT (golang-jwt/v5) + bcrypt |
| 实时通信 | gorilla/websocket |
| 日志 | zap + lumberjack（结构化日志 + 文件轮转） |
| 部署 | Docker + Docker Compose + Nginx |

## 快速开始

### 前提条件

只需要安装以下两个工具：

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

> 不需要单独安装 Go、Node.js、PostgreSQL 或 Redis，所有依赖都在 Docker 容器中运行。

### 第一步：克隆项目

```bash
git clone https://github.com/ldChengyi/LINKFLOW.git
cd LINKFLOW
```

### 第二步：配置环境变量

```bash
cp .env.example .env
```

用文本编辑器打开 `.env` 文件，**必须修改以下配置**（否则有安全风险）：

```bash
# ========== 必须修改 ==========

# 数据库角色密码（三个角色各设一个强密码）
DB_ADMIN_PASSWORD=your_admin_password_here
DB_APP_PASSWORD=your_app_password_here
DB_READ_PASSWORD=your_read_password_here

# JWT 签名密钥（用于用户登录 Token，务必随机生成）
JWT_SECRET=your-random-secret-key-at-least-32-chars

# PostgreSQL root 密码
POSTGRES_ROOT_PASSWORD=your_postgres_root_password
```

其他配置项可保持默认值，详见 `.env.example` 中的注释说明。

<details>
<summary>完整环境变量说明（点击展开）</summary>

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `COMPOSE_PROJECT_NAME` | `linkflow` | Docker 容器名前缀，多实例部署时需不同 |
| `WEB_PORT` | `8081` | Web 界面对外端口 |
| `SERVER_PORT` | `8080` | Go 后端端口（容器内部） |
| `POSTGRES_ROOT_PASSWORD` | `postgres` | PostgreSQL root 密码，**生产环境必改** |
| `DB_HOST` | `postgres` | 数据库主机（Docker 内部网络名） |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_NAME` | `linkflow` | 数据库名 |
| `DB_ADMIN_USER` / `DB_ADMIN_PASSWORD` | `linkflow_admin` / `change_me_admin` | 管理员角色凭证，**必改** |
| `DB_APP_USER` / `DB_APP_PASSWORD` | `linkflow_app` / `change_me_app` | 应用角色凭证，**必改** |
| `DB_READ_USER` / `DB_READ_PASSWORD` | `linkflow_read` / `change_me_read` | 只读角色凭证，**必改** |
| `REDIS_ADDR` | `redis:6379` | Redis 地址（Docker 内部网络名） |
| `REDIS_PASSWORD` | 空 | Redis 密码（可选） |
| `JWT_SECRET` | `your-super-secret-key-...` | JWT 签名密钥，**必改** |
| `JWT_EXPIRE_HOURS` | `24` | Token 过期时间（小时） |
| `LOG_LEVEL` | `info` | 日志级别：debug / info / warn / error |
| `MQTT_HOST` | `0.0.0.0` | MQTT Broker 监听地址 |
| `MQTT_PORT` | `1883` | MQTT Broker 对外端口 |

</details>

### 第三步：一键启动

```bash
docker-compose up -d --build
```

首次启动需要下载镜像和编译，大约需要几分钟。启动完成后：

- Web 界面：`http://localhost:8081`（或你在 `.env` 中配置的 `WEB_PORT`）
- MQTT Broker：`localhost:1883`

> 数据库和表会自动创建，无需手动执行迁移脚本。

### 第四步：开始使用

1. 在浏览器打开 Web 界面，**注册一个账号**
2. 创建**物模型**（定义设备的属性、事件、服务）
3. 创建**设备**，绑定物模型
4. 进入**在线调试**页面，模拟设备上线并测试数据上报
5. 在**设备数据**页面查看实时数据和历史图表

### 验证服务状态

```bash
# 查看所有容器是否正常运行
docker-compose ps

# 查看后端日志
docker-compose logs -f app

# 查看前端日志
docker-compose logs -f web

# 健康检查
curl http://localhost:8081/health
```

### Docker 服务说明

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| web | linkflow-web | 8081 (可配置) | Nginx 反向代理 + React 前端 |
| app | linkflow-app | 8080 (内部) / 1883 | Go 后端 + MQTT Broker |
| postgres | linkflow-postgres | 5432 (内部) | TimescaleDB 数据库 |
| redis | linkflow-redis | 6379 (内部) | Redis 缓存 |

> 注意：PostgreSQL 和 Redis 默认不暴露端口到宿主机，仅在 Docker 内部网络中通信。如需本地调试访问，可在 `docker-compose.yml` 中添加 `ports` 映射。

## 设备接入指南

### MQTT 连接参数

设备通过 MQTT 协议接入平台。在平台创建设备后，会获得以下凭证：

| 参数 | 值 |
|------|------|
| Broker 地址 | `你的服务器IP:1883` |
| Client ID | 设备 ID（UUID，创建设备时自动生成） |
| 用户名 | 不需要 |
| 密码 | 设备密钥（64 字符，创建设备时自动生成） |

### Topic 结构

```
devices/{device_id}/telemetry/up       # 设备 → 平台：上报遥测数据
devices/{device_id}/telemetry/down     # 平台 → 设备：下发属性设置
devices/{device_id}/event              # 设备 → 平台：事件上报
devices/{device_id}/service/invoke     # 平台 → 设备：服务调用
devices/{device_id}/service/reply      # 设备 → 平台：服务调用结果
devices/{device_id}/voice/up           # 设备 → 平台：语音文本上报
devices/{device_id}/voice/down         # 平台 → 设备：语音执行结果
devices/{device_id}/ota/down           # 平台 → 设备：OTA 升级命令
devices/{device_id}/ota/up             # 设备 → 平台：OTA 进度上报
```

### 快速测试（使用 mosquitto 命令行）

```bash
# 上报遥测数据
mosquitto_pub -h localhost -p 1883 \
  -i "你的设备ID" -P "你的设备密钥" \
  -t "devices/你的设备ID/telemetry/up" \
  -m '{"temperature": 25.6, "humidity": 60}'

# 订阅属性下发指令
mosquitto_sub -h localhost -p 1883 \
  -i "你的设备ID" -P "你的设备密钥" \
  -t "devices/你的设备ID/telemetry/down"
```

### ESP32 示例

`examples/esp32/` 目录下提供了完整的 ESP32 客户端示例（基于 PlatformIO），涵盖：
- WiFi 连接 + MQTT 认证
- 遥测上报（DHT11 温湿度）
- 属性下发接收 + 服务调用响应
- OTA 固件升级
- LED 模拟控制

## 项目结构

```
linkflow/
├── cmd/server/main.go           # 服务入口
├── internal/
│   ├── ai/                      # Dify Workflow API 集成
│   ├── cache/                   # Redis 客户端（在线状态、Token 缓存）
│   ├── config/                  # 环境变量配置加载
│   ├── database/                # PostgreSQL 多连接池（admin/app/read）+ RLS
│   ├── handler/                 # HTTP 处理器（REST API）
│   ├── logger/                  # zap 结构化日志
│   ├── middleware/              # JWT 认证 + 审计日志中间件
│   ├── model/                   # 数据模型定义
│   ├── mqtt/                    # MQTT Broker + 设备认证 + ACL + 数据校验 + 语音
│   ├── repository/              # 数据访问层（手写 SQL）
│   ├── scheduler/               # 定时任务 Cron 调度引擎
│   ├── service/                 # 业务逻辑层 + 接口定义 + Mock
│   └── ws/                      # WebSocket Hub（实时推送）
├── web/                         # React 前端项目
│   ├── src/pages/               # 页面组件
│   ├── src/hooks/               # 自定义 Hook（WebSocket、主题）
│   └── src/components/ui/       # shadcn/ui 组件
├── migrations/                  # 数据库迁移 SQL（自动执行）
├── examples/esp32/              # ESP32 客户端示例（PlatformIO）
├── scripts/                     # 数据库初始化脚本
├── tests/                       # 单元测试 + 集成测试
├── testdata/                    # 测试配置文件
├── docker-compose.yml           # Docker 编排
├── Dockerfile                   # 后端多阶段构建
└── .env.example                 # 环境变量模板
```

## 本地开发

如果需要在本地开发而非使用 Docker，需要以下环境：

### 环境要求

- Go 1.24+
- Node.js 18+ / npm
- PostgreSQL 15 + [TimescaleDB 扩展](https://docs.timescale.com/install/)
- Redis 7

### 后端

```bash
# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env，将 DB_HOST 改为 localhost，REDIS_ADDR 改为 localhost:6379

# 手动执行数据库迁移（按顺序执行 migrations/ 下的 SQL 文件）
# 或者先用 Docker 启动一次让它自动迁移

# 运行后端
go run cmd/server/main.go
```

### 前端

```bash
cd web
npm install
npm run dev
# 访问 http://localhost:5173
```

### 运行测试

```bash
# 单元测试（不需要数据库）
go test ./tests/ -v

# 集成测试（需要本地 PostgreSQL + Redis）
go test -tags integration ./tests/ -v
```

## 常见问题

**Q: 启动后访问页面显示 502 Bad Gateway？**
A: 后端服务可能还在启动中。运行 `docker-compose logs -f app` 查看后端日志，等待 `Server started` 日志出现。

**Q: 设备连接 MQTT 失败？**
A: 确认使用设备 ID 作为 Client ID、设备密钥作为密码。MQTT 不需要用户名字段。确认 1883 端口在防火墙中开放。

**Q: 如何重置数据库？**
A: 运行 `docker-compose down -v` 删除所有数据卷，然后重新 `docker-compose up -d --build`。

**Q: 如何查看数据库内容？**
A: 在 `docker-compose.yml` 的 postgres 服务中添加端口映射 `ports: ["5432:5432"]`，然后使用数据库客户端连接。

## 贡献

欢迎提交 Issue 和 Pull Request。

1. Fork 本项目
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 提交更改：`git commit -m "feat: add your feature"`
4. 推送分支：`git push origin feature/your-feature`
5. 创建 Pull Request

## License

[MIT](LICENSE)

---

<a id="english"></a>

## English

LinkFlow is a lightweight **IoT cloud platform** built with Go, designed for learning and small-to-medium IoT scenarios. It connects simulated or real IoT devices, collects sensor data, and provides real-time monitoring, alerts, voice control, OTA upgrades, and more.

### Key Features

- **Device Management** — CRUD, auto-generated secrets, Thing Model binding (properties/events/services)
- **MQTT Broker** — Embedded Mochi MQTT v2, device auth + ACL + Thing Model validation
- **Real-time Data** — MQTT telemetry → TimescaleDB storage → WebSocket push to browser
- **Visualization** — Live data cards + historical trend charts with auto-aggregation (1h to 30d)
- **Alerts** — Threshold rules, 3 severity levels, cooldown, real-time WebSocket notifications
- **Scheduled Tasks** — Cron-driven property setting and service invocation via MQTT
- **Voice Control** — Local NLP keyword matching or Dify Workflow API integration
- **OTA Upgrades** — Firmware upload → MQTT push → HTTP download → progress reporting
- **Online Debug** — Simulate device online/offline in browser without real hardware
- **Security** — JWT auth + PostgreSQL Row-Level Security + audit logs

### Quick Start

**Prerequisites:** Docker 20.10+ and Docker Compose v2+

```bash
git clone https://github.com/ldChengyi/LINKFLOW.git
cd LINKFLOW
cp .env.example .env
# Edit .env — change all passwords and JWT_SECRET
docker-compose up -d --build
```

Access the web UI at `http://localhost:8081`, register an account, and start creating devices.

See the Chinese documentation above for detailed setup instructions and configuration reference.

### Tech Stack

Go 1.24+ / Gin / pgx v5 / Mochi MQTT v2 / React 18 / TypeScript / shadcn/ui / Tailwind CSS / PostgreSQL 15 + TimescaleDB / Redis 7 / Docker

### License

[MIT](LICENSE)
