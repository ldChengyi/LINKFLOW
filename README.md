# LinkFlow IoT Platform

LinkFlow 是一个基于 Go 的轻量级物联网云平台，负责接入 IoT 设备、采集传感器数据，提供实时监控、告警、语音控制和定时任务等功能。

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

## 核心功能

### 设备接入与管理
- 设备 CRUD，自动生成 64 字符密钥
- 物模型管理（属性/事件/服务），支持 int/float/bool/string/enum 数据类型
- MQTT 内嵌 Broker（Mochi MQTT v2），设备通过 clientID + secret 认证
- ACL 控制，设备只能访问自己的 Topic
- 物模型属性校验（类型 + 范围 + 枚举），不合法数据标记后仍存储

### 实时数据链路
- 设备遥测数据通过 MQTT 上报，存入 TimescaleDB 时序表
- WebSocket 实时推送（遥测数据、设备上下线、告警通知、统计更新）
- Redis 维护设备实时在线状态，API 返回时合并

### 告警系统
- 阈值告警规则（> / >= / < / <= / == / !=）
- 三级告警：info / warning / critical
- 遥测数据入库后实时评估，触发后 WebSocket 推送 + 日志记录

### 定时任务
- Cron 表达式驱动，支持属性设置和服务调用
- 调度引擎每分钟 tick，通过 MQTT 下发指令到设备

### 语音控制
- 设备上报语音文本，本地关键词匹配 NLP 解析意图
- 支持 bool 开关、数值设置、相对调节、枚举匹配、服务调用
- 通过物模型 voice 模块配置可操控的属性和服务范围

### 数据可视化
- 设备最新数据卡片展示（正常/异常/未上报状态）
- 历史趋势图表（recharts），支持 1h/6h/24h/7d 时间范围切换
- 仪表盘统计概览（设备总数、在线数、物模型数）

### 设备在线调试
- 模拟上下线：无需真实硬件，通过 Redis 模拟设备在线状态
- 自动过期：模拟上线使用短 TTL（5分钟），前端心跳续期，离开页面后自动下线
- 属性下发 & 服务调用：通过 MQTT 下发指令，模拟设备自动回传数据
- 连接类型识别：区分真实 MQTT 连接和模拟上线，真实连接设备禁止模拟操作

### 安全体系
- JWT 认证 + Redis Token 存储（支持撤销/登出）
- PostgreSQL RLS 行级安全策略（用户只能访问自己的数据）
- 双层权限：应用层 JWT 角色 + 数据库层 PG Roles

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.24 / Gin / pgx v5 / Mochi MQTT v2 |
| 前端 | React 18 / TypeScript / shadcn/ui / Tailwind CSS / Vite |
| 数据库 | PostgreSQL 15 + TimescaleDB |
| 缓存 | Redis 7 |
| 认证 | JWT (golang-jwt/v5) + bcrypt |
| WebSocket | gorilla/websocket |
| 日志 | zap + lumberjack |
| 部署 | Docker + Docker Compose + Nginx |

## MQTT Topic 结构

```
devices/{device_id}/telemetry/up       # 设备上报遥测
devices/{device_id}/telemetry/down     # 服务端下发属性设置
devices/{device_id}/event              # 设备事件上报
devices/{device_id}/service/invoke     # 服务调用下发
devices/{device_id}/service/reply      # 服务执行结果回传
devices/{device_id}/voice/up           # 语音文本上报
devices/{device_id}/voice/down         # 语音执行结果回传
```

## 快速部署

### 环境要求

- Docker 20.10+
- Docker Compose v2+

### 一键启动

```bash
# 1. 克隆项目
git clone https://github.com/ldChengyi/LINKFLOW.git
cd LINKFLOW

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env，修改密码和 JWT 密钥

# 3. 启动所有服务
docker-compose up -d --build

# 4. 访问
# Web 界面: http://localhost
# MQTT Broker: localhost:1883
```

### 服务说明

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| web | linkflow-web | 80 | Nginx 反向代理 + 前端 |
| app | linkflow-app | 8080 (内部) / 1883 | Go 后端 + MQTT Broker |
| postgres | linkflow-postgres | 5432 | TimescaleDB 数据库 |
| redis | linkflow-redis | 6379 | Redis 缓存 |

### 关键环境变量

```bash
# 数据库角色密码
DB_ADMIN_PASSWORD=change_me_admin
DB_APP_PASSWORD=change_me_app
DB_READ_PASSWORD=change_me_read

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRE_HOURS=24

# MQTT
MQTT_PORT=1883
```

## 项目结构

```
linkflow/
├── cmd/server/main.go           # 服务入口
├── internal/
│   ├── cache/                   # Redis 客户端
│   ├── config/                  # 配置加载
│   ├── database/                # 多连接池 + RLS
│   ├── handler/                 # HTTP 处理器
│   ├── middleware/              # JWT 认证中间件
│   ├── model/                   # 数据模型
│   ├── mqtt/                    # MQTT Broker + 认证 + 校验 + 语音
│   ├── repository/              # 数据访问层
│   ├── scheduler/               # 定时任务调度引擎
│   ├── service/                 # 业务逻辑层
│   └── ws/                      # WebSocket Hub
├── web/                         # React 前端
├── migrations/                  # 数据库迁移脚本
├── scripts/                     # 初始化脚本
├── tests/                       # 测试
├── docker-compose.yml
├── Dockerfile
└── .env.example
```


```bash
# MQTT 连接参数
# clientID: 设备UUID（创建设备时生成）
# password: 设备密钥（64字符，创建设备时生成）

# 上报遥测数据
mosquitto_pub -h localhost -p 1883 \
  -i "device-uuid" -P "device-secret" \
  -t "devices/device-uuid/telemetry/up" \
  -m '{"temperature": 25.6, "humidity": 60}'

# 订阅属性下发
mosquitto_sub -h localhost -p 1883 \
  -i "device-uuid" -P "device-secret" \
  -t "devices/device-uuid/telemetry/down"
```

## License

MIT
