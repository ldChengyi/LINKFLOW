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
│   │   ├── stats.go                # 仪表盘统计处理器
│   │   ├── ws.go                   # WebSocket 升级处理器
│   │   ├── alert_rule.go           # 告警规则 CRUD 处理器
│   │   ├── alert_log.go            # 告警日志查询处理器
│   │   ├── scheduled_task.go       # 定时任务 CRUD 处理器
│   │   ├── debug.go                # 设备在线调试处理器（模拟上下线 + 指令下发）
│   │   ├── firmware.go             # 固件管理处理器（上传/下载/列表/删除）
│   │   └── ota_task.go             # OTA升级任务处理器（创建/列表/取消）
│   ├── logger/logger.go            # zap 日志封装
│   ├── middleware/auth.go          # JWT 认证中间件
│   ├── mqtt/
│   │   ├── broker.go               # MQTT Broker 生命周期 + WS推送 + 告警缓存
│   │   ├── auth.go                 # 设备认证 + ACL + 上线WS推送
│   │   ├── hooks.go                # 连接管理 + 消息处理 + 告警评估
│   │   ├── validator.go            # 物模型属性校验器
│   │   └── voice.go                # 语音指令处理器（本地 NLP）
│   ├── scheduler/
│   │   └── scheduler.go            # 定时任务调度引擎（Cron 评估 + MQTT 下发）
│   ├── ws/
│   │   └── hub.go                  # WebSocket Hub + Client 管理
│   ├── model/
│   │   ├── user.go                 # 用户模型
│   │   ├── thing_model.go          # 物模型（属性/事件/服务/模块）
│   │   ├── device.go               # 设备模型
│   │   ├── module.go               # 功能模块模型 + 语音指令结构
│   │   ├── alert.go                # 告警规则 + 告警日志模型
│   │   ├── scheduled_task.go       # 定时任务模型
│   │   ├── debug_log.go            # 设备调试日志模型
│   │   └── ota.go                  # 固件 + OTA任务模型
│   ├── repository/
│   │   ├── user.go                 # 用户数据访问
│   │   ├── thing_model.go          # 物模型数据访问
│   │   ├── device.go               # 设备数据访问
│   │   ├── device_data.go          # 遥测数据存储 + 查询
│   │   ├── module.go               # 功能模块数据访问
│   │   ├── alert_rule.go           # 告警规则数据访问
│   │   ├── alert_log.go            # 告警日志数据访问
│   │   ├── scheduled_task.go       # 定时任务数据访问
│   │   ├── debug_log.go            # 设备调试日志数据访问
│   │   ├── firmware.go             # 固件数据访问
│   │   └── ota_task.go             # OTA任务数据访问
│   ├── service/
│   │   ├── auth.go                 # 认证业务逻辑
│   │   ├── jwt.go                  # JWT + Redis 校验
│   │   ├── interfaces.go           # 可 mock 接口（UserRepo/TokenStore/TokenGenerator）
│   │   └── mock/                   # mockgen 生成的 mock
│   └── testutil/
│       └── config.go               # 测试配置加载工具
├── migrations/
│   ├── 001_init_roles.sql          # 角色 + 扩展初始化
│   ├── 002_users_table.sql         # 用户表
│   ├── 003_devices_table_rls.sql   # 设备表 + RLS 策略
│   ├── 004_device_data_hypertable.sql # 遥测数据时序表
│   ├── 005_modules.sql             # 功能模块表 + 物模型 modules 字段
│   ├── 007_alert_system.sql        # 告警规则表 + 告警日志表 + RLS
│   ├── 008_scheduled_tasks.sql     # 定时任务表 + RLS
│   ├── 009_alert_enhancements.sql  # 告警冷却字段 + 告警确认字段
│   ├── 010_ota_system.sql          # 固件表 + OTA任务表 + RLS + 设备firmware_version字段
│   ├── 011_platform_settings.sql  # 平台全局设置表（key-value）
│   └── 012_debug_logs.sql         # 设备调试日志表 + RLS
├── web/                             # 前端项目
│   ├── src/
│   │   ├── api/index.ts            # Axios API 客户端
│   │   ├── components/ui/          # shadcn/ui 组件
│   │   ├── hooks/
│   │   │   ├── useWebSocket.ts     # WebSocket hook（连接/重连/消息分发）
│   │   │   └── useTheme.ts         # 主题切换 hook（localStorage 持久化）
│   │   ├── pages/
│   │   │   ├── Login.tsx           # 登录/注册页
│   │   │   ├── Dashboard.tsx       # 仪表盘（侧边栏 + WS连接 + 全局事件总线）
│   │   │   ├── ThingModelList.tsx  # 物模型列表
│   │   │   ├── ThingModelForm.tsx  # 物模型创建/编辑（含模块配置）
│   │   │   ├── DeviceList.tsx      # 设备列表（实时状态更新）
│   │   │   ├── DeviceForm.tsx      # 设备创建/编辑
│   │   │   ├── DeviceData.tsx      # 设备数据查看（实时遥测 + 历史趋势图表）
│   │   │   ├── ModuleList.tsx      # 功能模块列表
│   │   │   ├── AlertRuleList.tsx   # 告警规则管理
│   │   │   ├── AlertRuleForm.tsx   # 告警规则创建/编辑
│   │   │   ├── AlertLogList.tsx    # 告警历史（实时新告警推送）
│   │   │   ├── ScheduledTaskList.tsx # 定时任务列表
│   │   │   ├── ScheduledTaskForm.tsx # 定时任务创建/编辑
│   │   │   ├── DeviceDebug.tsx      # 设备在线调试（模拟上下线 + 属性下发 + 服务调用）
│   │   │   ├── FirmwareList.tsx    # 固件管理（上传/列表/删除）
│   │   │   ├── OTATaskList.tsx     # OTA升级任务（创建/列表/取消 + 实时进度）
│   │   │   ├── LandingPage.tsx     # 落地页
│   │   │   └── Settings.tsx        # 系统设置（语音模式 + Dify 连接配置）
│   │   ├── App.tsx                 # 路由配置
│   │   └── main.tsx                # 入口文件
│   ├── Dockerfile                   # 前端镜像（Nginx）
│   ├── nginx.conf.template          # Nginx 配置模板
│   └── package.json
├── scripts/
│   ├── init-db.sh                   # 数据库初始化入口
│   └── init-db.sql.template         # SQL 模板（支持环境变量）
├── tests/                            # 统一测试包
│   ├── setup_test.go                # 测试配置加载（所有测试共享）
│   ├── setup_integration_test.go    # 集成测试 PG/Redis 初始化（integration tag）
│   ├── auth_test.go                 # 认证单元测试
│   ├── jwt_test.go                  # JWT 单元测试
│   └── auth_integration_test.go     # 认证集成测试（integration tag）
├── testdata/
│   └── auth_test.json               # 测试账号配置
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
   - 登录/注册页面（内联表单验证 + 错误/成功反馈）
   - 多主题支持（翠绿深色 / 天蓝浅色），CSS 变量 + `data-theme` 切换
   - 仪表盘（可折叠侧边栏 + 顶部用户菜单 + 主题切换）
   - 导航：仪表盘 / 物模型 / 设备管理 / 设备数据 / 功能模块 / 告警规则 / 告警历史 / 定时任务 / 在线调试
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

12. **WebSocket 实时推送**
    - 全局 Hub 管理所有 WebSocket 连接，按 `userID` 分组
    - 连接入口: `GET /api/ws?token=JWT`，JWT query param 认证
    - 心跳: 30s ping / 60s pong timeout
    - 消息格式: `{ "type": "telemetry|device_status|alert|stats", "data": {...} }`
    - 推送触发点：设备上线/离线（device_status + stats）、遥测上报（telemetry）、告警触发（alert）
    - 前端全局事件总线：Dashboard 建立 WS 连接，子页面通过 `onWSMessage()` 订阅
    - 自动重连（指数退避 3s→30s）
    - Nginx WebSocket 代理配置（Upgrade + Connection headers）

13. **告警系统**
    - 阈值告警规则：属性值 > / >= / < / <= / == / != 阈值时触发
    - 告警级别：info / warning / critical
    - 规则 CRUD + RLS 行级安全
    - 规则缓存：Broker `sync.Map`，CRUD 操作时失效，下次遥测时重新加载
    - 评估位置：`OnPublished` hook 中，数据存储后评估规则
    - 触发后：写入 `alert_logs` + WebSocket 推送 `alert` 消息
    - 前端：规则管理页（CRUD）+ 告警历史页（实时新告警插入）
    - 设备筛选：规则列表和告警历史均支持按设备筛选（下拉选择设备或查看全部）

14. **数据可视化（历史趋势图表）**
    - 后端 API：`GET /api/devices/:id/data/history?start=&end=&limit=`
    - TimescaleDB 时间范围查询，返回 time + payload + valid
    - 前端 recharts LineChart，支持 1h/6h/24h/7d 时间范围切换
    - 自动提取数值型属性（int/float）绘制多条折线

15. **定时任务系统**
    - Cron 表达式驱动，支持属性设置（property_set）和服务调用（service_invoke）
    - Scheduler 引擎：goroutine 每分钟 tick，评估 cron 表达式，通过 MQTT Publish 下发
    - 任务 CRUD + RLS 行级安全
    - 前端：任务列表（按设备筛选）+ 创建/编辑表单（Cron 预设 + 物模型属性/服务选择）
    - 内置 scheduler 功能模块，可在物模型中启用

### 第四阶段：质量保障 ✅

16. **测试体系**
    - Service 层接口抽象（UserRepo / TokenStore / TokenGenerator）支持 mock
    - mockgen 生成 mock，单元测试覆盖认证 + JWT 逻辑
    - 集成测试连接真实 PG + Redis，覆盖注册/登录/登出 + 全部 API 端点
    - JSON 配置文件（`testdata/auth_test.json`）管理测试账号和连接信息
    - `go test ./tests/ -v`（单元测试）/ `go test -tags integration ./tests/ -v`（集成测试）

17. **多主题系统**
    - CSS 变量主题：翠绿深色（默认）+ 天蓝浅色（`[data-theme="blue"]`）
    - `useTheme` hook + localStorage 持久化 + 顶部 Palette 按钮切换
    - 登录页 / 仪表盘 / Toast 通知均跟随主题

18. **登录反馈增强**
    - 内联错误/成功消息（表单内 banner，替代纯 toast）
    - 输入框字段级错误高亮（红色边框 + 图标）
    - 输入时自动清除错误状态，Tab 切换清除所有提示

19. **设备在线调试**
    - 连接类型识别：区分真实 MQTT 连接（real）和模拟上线（simulated）
    - 模拟上下线：通过 Redis 短 TTL（5分钟）模拟设备在线，前端心跳续期（每2分钟）
    - 自动过期：用户离开页面后心跳停止，Redis key 5分钟后自动过期，设备自动下线
    - 属性下发：通过 MQTT `telemetry/down` 下发属性设置
    - 服务调用：通过 MQTT `service/invoke` 下发服务调用
    - 智能回传：模拟设备自动回传 `telemetry/up`（闭合数据链路）；真实设备由硬件自行回传
    - 真实连接保护：真实 MQTT 连接的设备禁止模拟上下线操作
    - 前端：连接类型 Badge（MQTT连接/模拟在线/离线）+ 属性/服务控件 + WS实时数据更新

20. **告警冷却**
    - `alert_rules` 表新增 `cooldown_minutes` 字段（默认 0，不冷却）
    - MQTT 评估时记录 `last_triggered_at`（内存 sync.Map + DB 持久化）
    - 冷却期内跳过告警触发，避免频繁告警

21. **修改密码**
    - 后端 API：`PUT /api/auth/password`（需验证旧密码）
    - 前端：Dashboard 顶部用户菜单 ChangePassword Dialog

22. **仪表盘今日告警**
    - 统计概览新增第 4 张卡片：今日告警数
    - `GET /api/stats/overview` 返回 `today_alerts` 字段（按自然日统计）

23. **CSV 数据导出**
    - 后端 API：`GET /api/devices/:id/data/export?start=&end=`，返回 CSV 文件流
    - 前端：DeviceData 历史趋势图表页新增「导出 CSV」按钮

24. **告警确认 + 未读角标**
    - 后端 API：`PUT /api/alert-logs/:id/acknowledge`，标记告警为已确认
    - `alert_logs` 表新增 `acknowledged` / `acknowledged_at` / `acknowledged_by` 字段
    - 前端：告警历史列表新增确认按钮；侧边栏导航「告警历史」显示未读告警数角标

25. **审计日志**
    - 中间件自动记录 API 操作到 `audit_logs` 表（资源类型 + 操作类型 + 操作结果）
    - 覆盖 alert_rule / alert_log / auth 等资源类型的 actionMap
    - 前端：AuditLogList 页面展示审计日志（用户、操作、资源、时间、IP）

26. **OTA 固件升级**
    - 固件管理：上传固件文件（.bin/.hex/.elf）、SHA256 校验、本地文件存储（uploads/firmwares/）
    - OTA 任务：创建升级任务 → MQTT 推送命令 → 设备下载固件 → 上报进度 → 完成/失败
    - MQTT Topic：`devices/{id}/ota/down`（服务端→设备）、`devices/{id}/ota/up`（设备→服务端进度）
    - 固件下载：HTTP Basic Auth（device_id:device_secret），设备通过 HTTP 下载固件文件
    - 设备上线自动推送：设备连接时检查 pending OTA 任务并自动推送
    - 设备 firmware_version 字段：OTA 完成后自动更新设备固件版本
    - WebSocket 实时进度推送（ota_progress 消息类型）
    - 前端：固件管理页（上传/列表/删除）+ OTA任务页（创建/列表/取消 + 实时进度条）

27. **数据聚合查询**
    - History API 根据查询时长自动选择聚合粒度（无需前端传 interval）：
      - ≤ 2h → 原始数据（返回 `aggregated: false`）
      - ≤ 12h → 5 分钟窗口（≤ 144 点）
      - ≤ 48h → 15 分钟窗口（≤ 192 点）
      - ≤ 7d → 1 小时窗口（≤ 168 点）
      - > 7d → 6 小时窗口（≤ 120 点）
    - 统一响应格式：`{ aggregated, interval, data }`；聚合时每点包含 `payload`(avg) / `max_payload` / `min_payload`
    - SQL：`time_bucket($1::interval, time)` + `LATERAL jsonb_each_text(payload)` 过滤数值字段，Go 侧 pivot 到结构体
    - 前端：新增 30d 时间范围选项；聚合模式每属性渲染 3 条 Line（均值实线 + 上/下边界虚线）；底部显示聚合粒度文字
    - 关键文件：`internal/repository/device_data.go`（`GetDataHistoryAggregated`）、`internal/handler/device.go`（`History`）、`web/src/pages/DeviceData.tsx`

28. **设备调试日志**
    - 记录所有设备调试操作（属性设置/服务调用）到 `debug_logs` 表
    - 记录内容：连接类型（real/simulated）、操作类型、请求参数、响应结果、成功状态、错误信息
    - 后端 API：`GET /api/devices/:id/debug-logs?limit=&offset=`，支持分页查询
    - RLS 行级安全：用户只能查看自己的调试日志
    - 前端：DeviceDebug 页面展示调试历史记录

------

## API 端点

| 方法   | 路径                        | 说明               | 认证 |
| ------ | --------------------------- | ------------------ | ---- |
| GET    | /health                     | 健康检查           | 否   |
| POST   | /api/auth/register          | 用户注册           | 否   |
| POST   | /api/auth/login             | 用户登录           | 否   |
| POST   | /api/auth/logout            | 用户登出           | 是   |
| PUT    | /api/auth/password          | 修改密码           | 是   |
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
| GET    | /api/ws?token=JWT           | WebSocket 连接     | 是(query) |
| POST   | /api/alert-rules            | 创建告警规则       | 是   |
| GET    | /api/alert-rules            | 告警规则列表（?device_id 筛选） | 是   |
| GET    | /api/alert-rules/:id        | 告警规则详情       | 是   |
| PUT    | /api/alert-rules/:id        | 更新告警规则       | 是   |
| DELETE | /api/alert-rules/:id        | 删除告警规则       | 是   |
| GET    | /api/devices/:id/data/history | 设备历史遥测数据（自动聚合，返回 `{aggregated,interval,data}`） | 是   |
| GET    | /api/alert-logs             | 告警历史（?device_id 筛选） | 是   |
| PUT    | /api/alert-logs/:id/acknowledge | 确认告警           | 是   |
| GET    | /api/devices/:id/data/export | 设备历史数据 CSV 导出 | 是  |
| POST   | /api/scheduled-tasks        | 创建定时任务       | 是   |
| GET    | /api/scheduled-tasks        | 定时任务列表（?device_id 筛选） | 是   |
| GET    | /api/scheduled-tasks/:id    | 定时任务详情       | 是   |
| PUT    | /api/scheduled-tasks/:id    | 更新定时任务       | 是   |
| DELETE | /api/scheduled-tasks/:id    | 删除定时任务       | 是   |
| POST   | /api/devices/:id/debug      | 设备调试下发（属性设置/服务调用） | 是   |
| GET    | /api/devices/:id/connection-type | 查询设备连接类型（real/simulated/offline） | 是   |
| POST   | /api/devices/:id/simulate/online | 模拟设备上线       | 是   |
| POST   | /api/devices/:id/simulate/offline | 模拟设备下线      | 是   |
| POST   | /api/devices/:id/simulate/heartbeat | 模拟上线心跳续期 | 是   |
| GET    | /api/devices/:id/debug-logs | 设备调试日志列表（?limit&offset） | 是   |
| POST   | /api/firmwares              | 上传固件（multipart）| 是   |
| GET    | /api/firmwares              | 固件列表           | 是   |
| DELETE | /api/firmwares/:id          | 删除固件           | 是   |
| GET    | /api/firmwares/:id/download | 固件下载（Basic Auth）| 否(Basic) |
| POST   | /api/ota-tasks              | 创建OTA升级任务    | 是   |
| GET    | /api/ota-tasks              | OTA任务列表（?device_id） | 是   |
| GET    | /api/ota-tasks/:id          | OTA任务详情        | 是   |
| PUT    | /api/ota-tasks/:id/cancel   | 取消OTA任务        | 是   |

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
- [x] WebSocket 实时推送

### 第三阶段：业务增强
- [x] 功能模块系统（平台级模块 + 物模型绑定）
- [x] 语音控制（本地关键词匹配 NLP，通过 MQTT voice topic）
- [x] 告警系统（阈值规则 + WebSocket 通知 + 告警历史）
- [x] 数据可视化（历史趋势图表，recharts LineChart）
- [x] 定时任务系统（Cron 调度 + MQTT 下发 + 前端管理）
- [x] 设备在线调试（模拟上下线 + 属性下发 + 服务调用 + 智能回传）
- [x] 告警冷却（cooldown_minutes 字段 + last_triggered_at 冷却跳过）
- [x] 修改密码（PUT /api/auth/password + Dashboard Dialog）
- [x] 仪表盘今日告警（today_alerts 第 4 张统计卡）
- [x] CSV 数据导出（历史遥测数据导出按钮）
- [x] 告警确认 + 未读角标（acknowledge API + 侧边栏角标）
- [x] 审计日志（操作记录中间件 + 前端查询页面）
- [ ] 视频接入（RTMP）
- [x] OTA 固件升级（固件管理 + MQTT推送 + HTTP下载 + 进度上报 + 设备上线自动推送）
- [x] Dify 语音模式（platform_settings 表 + SettingsRepo/Handler + ai/dify.go + Broker 缓存 + Settings 页面）
- [x] 语音指令修复：真实 MQTT 设备不直接写 device_data，由设备 telemetry/up 回传更新
- [ ] 设备影子（Device Shadow，desired/reported 双状态 + 离线同步 delta）
- [x] 数据聚合查询（TimescaleDB time_bucket，自动选粒度，avg/max/min 三线图表，30d 范围支持）
- [ ] 多渠道告警通知（SMTP 邮件 + 钉钉/企业微信 Webhook）
- [ ] Webhook 数据转发（规则引擎轻量版，遥测数据 POST 转发到第三方）

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
devices/{device_id}/ota/down          # OTA升级命令下发（设备 SUB）
devices/{device_id}/ota/up            # OTA进度上报（设备 PUB）
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
2. `VoiceHandler` 根据 `platform_settings.voice_mode` 选择路径：
   - **local**：本地 Pipeline NLP（PreprocessNode → DeviceLoadNode → IntentClassifyNode → EntityExtractNode → SlotValidateNode → ExecuteNode）
   - **dify**：调用 Dify Workflow API（`ai.CallWorkflow`），返回结构化 `DifyCommand`
3. 执行：属性设置通过 `telemetry/down` 下发，服务调用通过 `service/invoke` 下发
4. 结果通过 `voice/down` 回传：`{"success": true, "message": "指令已执行", "action": "..."}`

### 语音指令写入规则
- **真实 MQTT 连接**（`IsClientConnected=true`）：只发 `telemetry/down`，等设备自行回传 `telemetry/up`
- **模拟设备**（`IsClientConnected=false`）：发 `telemetry/down` 后直接写入 device_data（模拟回传）

### 关键文件
- `internal/mqtt/broker.go` — Broker 生命周期管理 + Publish 方法 + WS推送 + 告警规则缓存 + settings缓存
- `internal/mqtt/auth.go` — 认证 + ACL Hook + 上线WS推送
- `internal/mqtt/hooks.go` — 连接管理 + 消息处理 Hook（分发 telemetry/event/voice）+ 告警评估
- `internal/mqtt/validator.go` — 物模型属性校验器
- `internal/mqtt/voice.go` — 语音指令处理器（本地NLP Pipeline + Dify路径 + 执行）
- `internal/ai/dify.go` — Dify Workflow API 调用（CallWorkflow，blocking模式）
- `internal/handler/settings.go` — 平台设置 CRUD + Broker 缓存失效
- `internal/repository/settings.go` — platform_settings 表 UPSERT
- `migrations/011_platform_settings.sql` — 平台设置表

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

## WebSocket 实时推送

### 架构
- 全局 Hub 管理所有连接，按 `userID` 分组（一个用户可多个连接）
- gorilla/websocket 实现，readPump/writePump 双协程模式
- 连接入口: `GET /api/ws?token=JWT`，JWT 通过 query param 认证
- 心跳: 30s ping，60s pong timeout
- 消息格式: `{ "type": "telemetry|device_status|alert|stats", "data": {...} }`

### 推送触发点
| 事件 | 推送类型 | 触发位置 |
|------|---------|---------|
| 设备上报遥测 | `telemetry` | hooks.go OnPublished |
| 设备上线 | `device_status` + `stats` | auth.go OnConnectAuthenticate |
| 设备离线 | `device_status` + `stats` | hooks.go OnDisconnect |
| 告警触发 | `alert` | hooks.go checkAlertRules |

### 前端集成
- `useWebSocket` hook：自动连接/重连（指数退避 3s→30s）
- Dashboard 建立全局 WS 连接，通过 `onWSMessage()` 事件总线分发给子页面
- DeviceList：实时更新设备在线状态
- DeviceData：实时更新属性卡片数据
- AlertLogList：新告警实时插入列表顶部
- Dashboard：统计卡片实时更新 + 告警 toast 通知

### 关键文件
- `internal/ws/hub.go` — Hub + Client + readPump/writePump + SendToUser
- `internal/handler/ws.go` — WebSocket 升级 handler（JWT query param 认证）
- `web/src/hooks/useWebSocket.ts` — React WS hook
- `web/src/pages/Dashboard.tsx` — 全局 WS 连接 + 事件总线（`onWSMessage`）
- `web/nginx.conf.template` — WebSocket 代理配置（Upgrade headers + 3600s timeout）

------

## 告警系统

### 架构
- 阈值告警：属性值与阈值比较（> / >= / < / <= / == / !=）
- 告警级别：info / warning / critical
- 规则缓存：Broker `alertRules` sync.Map，key 为 deviceID
- 缓存失效：告警规则 CRUD 时通过 `AlertRuleBrokerInvalidator` 接口清除
- 评估位置：`OnPublished` hook 中，遥测数据存储后评估

### 数据流
1. 设备上报遥测 → hooks.go 存储数据
2. `checkAlertRules()` 加载规则（优先缓存，miss 时查 DB）
3. 遍历规则，比较属性值与阈值
4. 触发时：写入 `alert_logs` + WebSocket 推送 `alert` 消息

### 数据库表
```sql
-- alert_rules: 用户定义的告警规则
alert_rules (id UUID PK, user_id, name, device_id, model_id, property_id, operator, threshold, severity, enabled, cooldown_minutes INT DEFAULT 0, last_triggered_at TIMESTAMPTZ, created_at, updated_at)

-- alert_logs: 触发的告警记录
alert_logs (id BIGSERIAL PK, rule_id, user_id, device_id, device_name, property_id, property_name, operator, threshold, actual_value, severity, rule_name, acknowledged BOOL DEFAULT FALSE, acknowledged_at TIMESTAMPTZ, acknowledged_by UUID, created_at)
```

### 关键文件
- `internal/model/alert.go` — AlertRule + AlertLog 模型
- `internal/repository/alert_rule.go` — 规则 CRUD + `ListEnabledByDeviceID`（绕过 RLS）
- `internal/repository/alert_log.go` — 日志写入（绕过 RLS）+ 查询（RLS）
- `internal/handler/alert_rule.go` — 规则 CRUD handler + Broker 缓存失效
- `internal/handler/alert_log.go` — 日志查询 handler
- `internal/mqtt/hooks.go` — `checkAlertRules()` 告警评估方法
- `migrations/007_alert_system.sql` — 表结构 + RLS + 索引

------

## 设备在线调试

### 架构
- 区分真实 MQTT 连接（real）和模拟上线（simulated）两种在线模式
- 真实连接：设备通过 MQTT Broker 认证上线，由 `server.Clients.Get()` 判断
- 模拟上线：通过 Redis 短 TTL（5分钟）标记在线，前端心跳（2分钟）续期
- 调试下发时，模拟设备自动回传 `telemetry/up`；真实设备由硬件自行回传

### 连接类型判断
1. `Broker.IsClientConnected(deviceID)` → Mochi `server.Clients.Get()` + `!cl.Closed()` → **real**
2. Redis `device:online:{device_id}` 存在但无真实连接 → **simulated**
3. 两者都没有 → **offline**

### 模拟上线自动过期
- `SetSimulatedOnline()` 写 Redis，TTL = 5 分钟
- 前端 `setInterval` 每 2 分钟调 `POST /simulate/heartbeat` 续期
- 用户离开页面 / 切换设备 → interval 清理 → 5 分钟后自动过期下线

### 智能回传
- 模拟设备：下发 `telemetry/down` 后自动 Publish `telemetry/up`（InlineClient），触发完整数据链路（存储 + WS 推送 + 告警评估）
- 真实设备：仅下发 `telemetry/down`，等待硬件自行回传
- `OnPublished` hook 通过 `extractDeviceIDFromTopic()` 从 topic 提取设备 ID（兼容 InlineClient，其 `cl.ID` 为 `"inline"`）
- `loadDeviceInfo()` 缓存未命中时回查 DB 并缓存（支持模拟设备场景）

### 真实连接保护
- 真实 MQTT 连接的设备：禁止模拟上线、禁止模拟下线
- 前端：真实连接时模拟上下线按钮 disabled

### 关键文件
- `internal/handler/debug.go` — Debug（调试下发）+ SimulateOnline/Offline + SimulateHeartbeat + ConnectionType
- `internal/mqtt/broker.go` — `IsClientConnected()` 查询真实 MQTT 连接
- `internal/mqtt/hooks.go` — `extractDeviceIDFromTopic()` + `loadDeviceInfo()` 兼容 InlineClient
- `internal/cache/redis.go` — `SetSimulatedOnline()` / `RefreshSimulatedOnline()` 短 TTL 管理
- `web/src/pages/DeviceDebug.tsx` — 连接类型 Badge + 心跳续期 + 属性/服务控件

------

## 开发规范

- 不使用 ORM，手写 SQL
- 使用 pgx v5 连接池
- 日志使用 `logger.Log.Info/Error/...`
- 配置通过 `.env` 环境变量
- 数据库迁移脚本放 `migrations/`
- RLS 会话变量通过 `database.WithRLS()` 设置（`set_config` 第三参数为 `false`，会话级别）
- RLS 策略中必须使用 `current_setting('app.current_user_id', true)`，注意变量名是 `app.current_user_id`（非 `app.user_id`）
- RLS 连接用完后必须调用 `database.ReleaseRLSConn(ctx)` 释放
- MQTT 相关 Repository 使用 `db.Admin()` 绕过 RLS（设备认证/语音处理时无用户上下文）
- Repository 中通过 `queryRow`/`query`/`exec` 方法自动判断使用 RLS 连接还是 pool
- 前端 UI 组件使用 shadcn/ui，样式使用 Tailwind CSS
- 前端通知使用 sonner（`toast.success/error`）
- Broker 内部缓存：`devices` sync.Map 缓存已连接设备信息，`models` sync.Map 缓存物模型属性，`alertRules` sync.Map 缓存告警规则
- WebSocket Hub 通过 `SendToUser(userID, msg)` 推送消息，Broker 持有 Hub 引用

------

## 测试规范

### 统一测试包

所有测试统一放在 `tests/` 包（`package tests`），每个功能模块一个测试文件。新增功能时在 `tests/` 下新建对应的 `xxx_test.go`。

### 文件结构

| 文件 | 说明 |
|------|------|
| `tests/setup_test.go` | 共享配置加载（`init()` 读取 `testdata/auth_test.json`） |
| `tests/setup_integration_test.go` | 集成测试 PG/Redis 初始化 + setupRouter（`//go:build integration`） |
| `tests/auth_test.go` | 认证单元测试（mock UserRepo + TokenGenerator） |
| `tests/jwt_test.go` | JWT 单元测试（mock TokenStore） |
| `tests/auth_integration_test.go` | 认证集成测试（真实 PG + Redis） |
| `tests/api_helpers_test.go` | API 测试基础设施（路由初始化 + HTTP 辅助函数） |
| `tests/api_auth_test.go` | 认证 API 测试（注册/登录/登出/me/未授权） |
| `tests/api_thing_model_test.go` | 物模型 API 测试（CRUD 全流程） |
| `tests/api_device_test.go` | 设备 API 测试（CRUD + 最新数据 + 历史数据） |
| `tests/api_stats_test.go` | 仪表盘统计 API 测试 |
| `tests/api_module_test.go` | 功能模块 API 测试（列表 + 详情） |
| `tests/api_alert_test.go` | 告警规则 CRUD + 告警日志列表 API 测试 |
| `tests/api_scheduled_task_test.go` | 定时任务 CRUD API 测试 |
| `tests/api_device_data_test.go` | 历史遥测数据聚合 API 测试（raw/5min/1h/6h 四档 + 空结果） |

### Mock 生成

接口定义在 `internal/service/interfaces.go`，mock 通过 mockgen 生成：

```bash
mockgen -source=internal/service/interfaces.go -destination=internal/service/mock/mock_interfaces.go -package=mock
```

### 运行测试

```bash
# 单元测试
go test ./tests/ -v

# 集成测试（需本地 PG + Redis）
go test -tags integration ./tests/ -v
```

### 测试配置

测试账号存储在 `testdata/auth_test.json`，邮箱统一使用 `@linkflow.dev` 域名。集成测试 Redis 使用 DB 1 与开发环境隔离。

### 新增功能测试约定

1. 在 `tests/` 下新建 `{feature}_test.go`
2. 如需 mock 新接口，在 `internal/service/interfaces.go` 添加接口定义，重新运行 mockgen
3. 如需新测试账号，在 `testdata/auth_test.json` 的 `accounts` 中添加
4. 集成测试文件头部加 `//go:build integration`
5. 更新本节文件结构表
