# LinkFlow IoT Platform — Claude Code 项目说明

## 项目概述

**LinkFlow** 是一个基于 Go 的轻量级物联网云平台毕业设计项目。 平台负责接入模拟 IoT 设备，采集传感器数据，提供实时监控、语音控制、视频接入功能。

------

## 技术栈

### 后端

| 用途           | 技术                           |
| -------------- | ------------------------------ |
| 语言           | Go 1.24+                       |
| Web 框架       | Gin                            |
| MQTT Broker    | Mochi MQTT v2（内嵌，非外部服务）|
| 数据库驱动     | pgx v5（不使用 ORM，手写 SQL） |
| 环境变量       | godotenv                       |
| JWT 认证       | golang-jwt/jwt v5              |
| WebSocket      | gorilla/websocket              |
| 日志           | zap + lumberjack               |

### 数据存储

| 用途     | 技术                             |
| -------- | -------------------------------- |
| 主数据库 | PostgreSQL 15 + TimescaleDB 扩展 |
| 缓存     | Redis 7                          |

### 消息协议

| 用途        | 技术                                    |
| ----------- | --------------------------------------- |
| MQTT Broker | Mochi MQTT v2（Go 内嵌，无需额外容器） |
| 协议版本    | MQTT 3.1.1                              |

### 扩展功能

| 用途         | 技术                                     |
| ------------ | ---------------------------------------- |
| 语音控制     | 本地关键词匹配 NLP（内置于 MQTT Broker） |
| 视频流服务器 | Nginx-RTMP（待开发）                     |
| 视频转码     | FFmpeg（待开发）                         |

### 前端

| 用途        | 技术                       |
| ----------- | -------------------------- |
| 框架        | React 18 + TypeScript      |
| UI 组件库   | shadcn/ui + Tailwind CSS   |
| 通知提示    | sonner                     |
| 图标        | lucide-react               |
| HTTP 客户端 | Axios                      |
| 构建工具    | Vite                       |

### 部署

| 用途     | 技术                    |
| -------- | ----------------------- |
| 容器化   | Docker + Docker Compose |
| 反向代理 | Nginx                   |

------

## 项目结构

```
linkflow/
├── cmd/
│   └── server/main.go              # 服务入口
├── internal/
│   ├── cache/redis.go              # Redis 客户端
│   ├── config/config.go            # 配置加载
│   ├── database/database.go        # 多连接池 + RLS 支持
│   ├── handler/
│   │   ├── auth.go                 # 认证处理器
│   │   ├── thing_model.go          # 物模型处理器
│   │   ├── device.go               # 设备处理器
│   │   ├── module.go               # 功能模块处理器
│   │   └── stats.go                # 仪表盘统计处理器
│   ├── logger/logger.go            # zap 日志封装
│   ├── middleware/auth.go          # JWT 认证中间件
│   ├── mqtt/
│   │   ├── broker.go               # MQTT Broker 生命周期
│   │   ├── auth.go                 # 设备认证 + ACL
│   │   ├── hooks.go                # 连接管理 + 消息处理
│   │   ├── validator.go            # 物模型属性校验器
│   │   └── voice.go                # 语音指令处理器（本地 NLP）
│   ├── model/
│   │   ├── user.go                 # 用户模型
│   │   ├── thing_model.go          # 物模型（属性/事件/服务/模块）
│   │   ├── device.go               # 设备模型
│   │   └── module.go               # 功能模块模型 + 语音指令结构
│   ├── repository/
│   │   ├── user.go                 # 用户数据访问
│   │   ├── thing_model.go          # 物模型数据访问
│   │   ├── device.go               # 设备数据访问
│   │   ├── device_data.go          # 遥测数据存储 + 查询
│   │   └── module.go               # 功能模块数据访问
│   └── service/
│       ├── auth.go                 # 认证业务逻辑
│       └── jwt.go                  # JWT + Redis 校验
├── migrations/
│   ├── 001_init_roles.sql          # 角色 + 扩展初始化
│   ├── 002_users_table.sql         # 用户表
│   ├── 003_devices_table_rls.sql   # 设备表 + RLS 策略
│   ├── 004_device_data_hypertable.sql # 遥测数据时序表
│   └── 005_modules.sql             # 功能模块表 + 物模型 modules 字段
├── web/                             # 前端项目
│   ├── src/
│   │   ├── api/index.ts            # Axios API 客户端
│   │   ├── components/ui/          # shadcn/ui 组件
│   │   ├── pages/
│   │   │   ├── Login.tsx           # 登录/注册页
│   │   │   ├── Dashboard.tsx       # 仪表盘（侧边栏布局）
│   │   │   ├── ThingModelList.tsx  # 物模型列表
│   │   │   ├── ThingModelForm.tsx  # 物模型创建/编辑（含模块配置）
│   │   │   ├── DeviceList.tsx      # 设备列表
│   │   │   ├── DeviceForm.tsx      # 设备创建/编辑
│   │   │   ├── DeviceData.tsx      # 设备数据查看（属性/物模型/接口详情）
│   │   │   └── ModuleList.tsx      # 功能模块列表
│   │   ├── App.tsx                 # 路由配置
│   │   └── main.tsx                # 入口文件
│   ├── Dockerfile                   # 前端镜像（Nginx）
│   ├── nginx.conf.template          # Nginx 配置模板
│   └── package.json
├── scripts/
│   ├── init-db.sh                   # 数据库初始化入口
│   └── init-db.sql.template         # SQL 模板（支持环境变量）
├── .env                              # 环境配置（不提交）
├── .env.example                      # 配置模板
├── docker-compose.yml                # 一键部署
├── Dockerfile                        # 后端镜像（多阶段构建）
└── go.mod
```

------

## 已完成功能

### 第一阶段：基础设施 ✅

1. **用户认证系统**
   - 邮箱注册/登录
   - JWT Token 生成与验证
   - Redis 存储 Token（支持撤销/登出）
   - 密码 bcrypt 加密

2. **双层权限模型**
   - 应用层：JWT 角色（admin/user/viewer）
   - 数据库层：PostgreSQL Roles（linkflow_admin/app/read）
   - RLS 行级安全策略（用户只能访问自己的数据）

3. **基础设施**
   - Docker Compose 一键部署
   - 多数据库连接池（admin/app/read 三角色）
   - zap 结构化日志（控制台 + 文件轮转）
   - 环境变量配置

4. **Web 前端**
   - React 18 + TypeScript + shadcn/ui
   - 登录/注册页面（深色主题）
   - 仪表盘（可折叠侧边栏 + 顶部用户菜单）
   - 导航：仪表盘 / 物模型 / 设备管理 / 设备数据 / 功能模块
   - Nginx 反向代理（80 端口）
   - `/api/*` 请求转发至后端

5. **物模型管理**
   - 类阿里云物模型结构（属性/事件/服务）
   - 支持多种数据类型（int/float/bool/string/enum）
   - 模块绑定（物模型可启用功能模块并配置暴露的属性/服务）
   - RLS 行级安全（用户只能访问自己的物模型）
   - 前端 CRUD 完整实现

6. **设备管理**
   - 设备 CRUD（创建/查询/更新/删除）
   - 物模型绑定（可选关联）
   - 自动生成 64 字符设备密钥
   - 设备在线状态：Redis 实时源 + PG 持久化（MQTT 上下线触发）
   - API 返回设备列表/详情时合并 Redis 实时在线状态
   - RLS 行级安全（用户只能访问自己的设备）
   - 前端：设备列表（密钥脱敏）、创建/编辑表单、仪表盘统计概览

### 第二阶段：核心数据链路 ✅

7. **MQTT 数据接入**
   - 内嵌 Mochi MQTT v2 Broker（TCP 1883）
   - 设备认证（clientID + device_secret）
   - ACL 控制（设备只能访问自己的 topic）
   - 物模型属性校验（类型 + 范围 + 枚举）
   - 不合法数据标记 `valid=false` 后仍存储

8. **TimescaleDB 时序存储**
   - device_data hypertable 存储遥测数据
   - 支持查询设备最新一条数据

9. **设备数据查看页面**
   - 选择设备查看最新上报数据
   - 属性卡片展示（正常/异常/未上报状态）
   - 物模型详情 Tab（属性表格 + 上报/下发格式示例）
   - 事件/服务定义展示 + Payload 示例
   - 接口详情 Tab（MQTT Topic + 认证参数 + Payload 示例）
   - 模块 Tab（展示已启用模块的配置和 Topic）

10. **功能模块系统**
    - 平台级模块表（modules），支持配置 schema
    - 物模型通过 `modules` JSONB 字段绑定模块 + 配置
    - 前端模块列表页（展示模块说明、配置项、使用方式）
    - 内置语音控制模块（voice）

11. **语音控制（本地 NLP）**
    - 设备通过 `voice/up` topic 上报语音文本
    - 本地关键词匹配解析意图（设备名 + 属性名/服务名）
    - 支持 bool 开关（打开/关闭）、数值设置（调到X）、相对调节（调高/调低）、枚举匹配、服务调用
    - 通过物模型 voice 模块的 `exposed_properties` / `exposed_services` 控制可操控范围
    - 执行结果通过 `voice/down` topic 回传
    - 属性设置时同步写入 device_data（合并已有属性）

------

## API 端点

| 方法   | 路径                        | 说明               | 认证 |
| ------ | --------------------------- | ------------------ | ---- |
| GET    | /health                     | 健康检查           | 否   |
| POST   | /api/auth/register          | 用户注册           | 否   |
| POST   | /api/auth/login             | 用户登录           | 否   |
| POST   | /api/auth/logout            | 用户登出           | 是   |
| GET    | /api/me                     | 获取当前用户       | 是   |
| POST   | /api/thing-models           | 创建物模型         | 是   |
| GET    | /api/thing-models           | 物模型列表         | 是   |
| GET    | /api/thing-models/:id       | 物模型详情         | 是   |
| PUT    | /api/thing-models/:id       | 更新物模型         | 是   |
| DELETE | /api/thing-models/:id       | 删除物模型         | 是   |
| POST   | /api/devices                | 创建设备           | 是   |
| GET    | /api/devices                | 设备列表           | 是   |
| GET    | /api/devices/:id            | 设备详情           | 是   |
| GET    | /api/devices/:id/data/latest| 设备最新遥测数据   | 是   |
| PUT    | /api/devices/:id            | 更新设备           | 是   |
| DELETE | /api/devices/:id            | 删除设备           | 是   |
| GET    | /api/modules                | 功能模块列表       | 是   |
| GET    | /api/modules/:id            | 功能模块详情       | 是   |
| GET    | /api/stats/overview         | 仪表盘统计概览     | 是   |

------

## 快速启动

```bash
# 1. 复制配置
cp .env.example .env

# 2. 修改 .env 中的密码和密钥

# 3. 一键启动
docker-compose up -d --build

# 4. 访问应用
# Web 界面: http://localhost
# API: http://localhost/api/...

# 5. 查看日志
docker-compose logs -f app
docker-compose logs -f web
```

## Docker 服务

| 服务     | 容器名            | 端口  | 说明                     |
| -------- | ----------------- | ----- | ------------------------ |
| web      | linkflow-web      | 80    | Nginx 反向代理 + 前端    |
| app      | linkflow-app      | 8080  | Go 后端（内部暴露）      |
| postgres | linkflow-postgres | 5432  | TimescaleDB 数据库       |
| redis    | linkflow-redis    | 6379  | Redis 缓存               |

------

## 待开发功能

### 第二阶段：核心数据链路
- [x] 设备管理（CRUD + 物模型绑定）
- [x] MQTT 数据接入（Mochi MQTT 内嵌 Broker）
- [x] TimescaleDB 时序存储（device_data hypertable）
- [x] 设备数据查看页面（最新数据 + 物模型详情 + 接口详情）
- [ ] WebSocket 实时推送

### 第三阶段：业务增强
- [x] 功能模块系统（平台级模块 + 物模型绑定）
- [x] 语音控制（本地关键词匹配 NLP，通过 MQTT voice topic）
- [ ] 告警系统（阈值规则、通知）
- [ ] 视频接入（RTMP）

------

## MQTT 接入

### 架构
- 内嵌 Mochi MQTT v2 Broker，无需额外容器
- TCP 端口 1883（通过 docker-compose 暴露）
- InlineClient 模式（服务端可直接 Publish 消息）

### Topic 结构
```
devices/{device_id}/telemetry/up      # 设备上报遥测（设备 PUB）
devices/{device_id}/telemetry/down    # 服务端下发属性设置（设备 SUB）
devices/{device_id}/event             # 设备事件上报（设备 PUB）
devices/{device_id}/service/invoke    # 服务调用下发（设备 SUB）
devices/{device_id}/service/reply     # 服务执行结果回传（设备 PUB）
devices/{device_id}/voice/up          # 语音文本上报（设备 PUB）
devices/{device_id}/voice/down        # 语音执行结果回传（设备 SUB）
```

### 设备认证
- clientID = device_id（UUID）
- password = device_secret（64字符）
- ACL：设备只能访问自己的 topic，不能发布到 /down 和 /invoke

### 数据校验
- 根据设备绑定的物模型 Properties 校验上报数据
- 校验类型（int/float/bool/string/enum）+ 范围（min/max）+ 枚举值
- 不合法数据标记 `valid=false` + `errors` 字段后仍存储（便于故障排查）
- 未绑定物模型的设备跳过校验

### 语音指令处理流程
1. 设备发布 `voice/up` topic，payload: `{"text": "打开灯"}`
2. `VoiceHandler` 解析意图：匹配目标设备名 → 获取物模型 → 检查 voice 模块 → 匹配属性/服务
3. 执行：属性设置通过 `telemetry/down` 下发，服务调用通过 `service/invoke` 下发
4. 结果通过 `voice/down` 回传：`{"success": true, "message": "指令已执行", "action": "..."}`

### 关键文件
- `internal/mqtt/broker.go` — Broker 生命周期管理 + Publish 方法
- `internal/mqtt/auth.go` — 认证 + ACL Hook
- `internal/mqtt/hooks.go` — 连接管理 + 消息处理 Hook（分发 telemetry/event/voice）
- `internal/mqtt/validator.go` — 物模型属性校验器
- `internal/mqtt/voice.go` — 语音指令处理器（意图解析 + 执行 + 相对值计算）

------

## 设备在线状态（Redis）

### 架构
- Redis 作为实时在线状态数据源，PG 保留持久化记录
- MQTT Broker 上下线事件同时写 Redis 和 PG，Redis 失败只 log 不阻断

### Redis Key 设计
- `device:online:{device_id}` — SET，值为 userID，TTL 24h
- `user:online_devices:{user_id}` — Set 集合，存储该用户所有在线设备 ID

### 数据流
1. 设备 MQTT 连接 → `OnConnect` → Redis `SET` + `SADD` + PG `UPDATE status='online'`
2. 设备 MQTT 断开 → `OnDisconnect` → Redis `DEL` + `SREM` + PG `UPDATE status='offline'`
3. API `GET /devices` / `GET /devices/:id` → PG 查询 + Redis `BatchCheckOnline` 合并状态
4. API `GET /stats/overview` → PG `COUNT` (设备/物模型) + Redis `SCARD` (在线数)

### 关键文件
- `internal/cache/redis.go` — SetDeviceOnline/Offline, BatchCheckOnline, CountUserOnlineDevices
- `internal/mqtt/hooks.go` — OnConnect/OnDisconnect 写 Redis
- `internal/handler/device.go` — mergeOnlineStatus 合并实时状态
- `internal/handler/stats.go` — 仪表盘统计（设备总数/在线数/物模型数）

------

## 功能模块系统

### 架构
- `modules` 表存储平台级模块定义（如 voice），包含 `config_schema` 描述可配置项
- `thing_models.modules` JSONB 字段存储物模型绑定的模块列表及配置
- 模块是平台级资源，所有角色可读，仅 admin 可写

### 物模型模块绑定格式
```json
[{"id": "voice", "config": {"exposed_properties": ["switch", "brightness"], "exposed_services": ["reboot"]}}]
```

### 语音模块（voice）
- 通过 `exposed_properties` 控制哪些 rw 属性可被语音操控
- 通过 `exposed_services` 控制哪些服务可被语音调用
- 支持的语音指令类型：
  - bool 开关：打开/关闭/开启/关掉
  - 数值设置：调到X/设为X
  - 相对调节：调高/调低/增大/减小（使用属性 step 或默认步长 10）
  - 枚举匹配：匹配枚举 label
  - 服务调用：匹配服务名称

------

## 开发规范

- 不使用 ORM，手写 SQL
- 使用 pgx v5 连接池
- 日志使用 `logger.Log.Info/Error/...`
- 配置通过 `.env` 环境变量
- 数据库迁移脚本放 `migrations/`
- RLS 会话变量通过 `database.WithRLS()` 设置（`set_config` 第三参数为 `false`，会话级别）
- RLS 连接用完后必须调用 `database.ReleaseRLSConn(ctx)` 释放
- MQTT 相关 Repository 使用 `db.Admin()` 绕过 RLS（设备认证/语音处理时无用户上下文）
- Repository 中通过 `queryRow`/`query`/`exec` 方法自动判断使用 RLS 连接还是 pool
- 前端 UI 组件使用 shadcn/ui，样式使用 Tailwind CSS
- 前端通知使用 sonner（`toast.success/error`）
- Broker 内部缓存：`devices` sync.Map 缓存已连接设备信息，`models` sync.Map 缓存物模型属性
