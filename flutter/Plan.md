# LinkFlow Flutter 语音助手 APP

## Context

ESP32 硬件麦克风和音频输出有问题，开发 Flutter APP 替代。采用**聊天式交互界面**，像聊天软件一样与设备对话。

## 用户流程

```
首次启动 → 填写后端 IP 地址 → 登录 → 选择设备 → 聊天式语音/文字交互
```

## 界面设计

### 1. 服务器配置页（首次启动）
- 输入 LinkFlow 后端 IP 地址（如 `192.168.1.100`）
- 可选端口（默认 8080）
- "连接" 按钮 → 测试连通性（调 /health）
- 连接成功后保存到本地，跳转登录页
- 后续启动自动跳过（设置页可修改）

### 2. 登录页
- 邮箱 + 密码
- 登录按钮
- 底部小字显示当前服务器地址（可点击返回修改）

### 3. 主页（聊天界面）
- **顶部 AppBar**: 当前设备名 + 在线状态指示灯，点击切换设备
- **聊天消息列表**（类似微信/Telegram）:
  - 右侧气泡：用户发出的语音/文字指令
  - 左侧气泡：设备回复（成功/失败 + 消息 + 执行动作）
  - 支持播放 TTS 音频（气泡内播放按钮）
  - 设备数据卡片（遥测数据、设备状态等嵌入聊天流）
- **底部输入栏**:
  - 文字输入框
  - 麦克风按钮（长按说话，松开发送）
  - 发送按钮
- **快捷操作菜单**（输入栏左侧 + 按钮展开）:
  - 查看设备数据
  - 下发属性
  - 调用服务

## 项目结构

```
examples/flutter_app/
├── Plan.md                          # 本计划文档
├── pubspec.yaml
├── lib/
│   ├── main.dart                    # 入口 + 主题 + Provider
│   ├── config.dart                  # 常量定义
│   ├── models/
│   │   ├── user.dart                # 用户模型
│   │   ├── device.dart              # 设备模型
│   │   ├── voice_result.dart        # 语音结果模型
│   │   └── chat_message.dart        # 聊天消息模型（用户指令/设备回复/数据卡片）
│   ├── services/
│   │   ├── api_service.dart         # HTTP 客户端
│   │   ├── auth_service.dart        # 认证 + Token 管理（ChangeNotifier）
│   │   └── audio_service.dart       # TTS 音频播放
│   ├── screens/
│   │   ├── server_config_screen.dart # 服务器配置页（首次启动）
│   │   ├── login_screen.dart        # 登录页
│   │   └── chat_screen.dart         # 主聊天页
│   └── widgets/
│       ├── chat_bubble.dart         # 聊天气泡（用户/设备两种样式）
│       ├── voice_input_bar.dart     # 底部输入栏（文字+麦克风+发送）
│       ├── device_picker.dart       # 设备选择弹窗
│       ├── device_data_card.dart    # 设备数据卡片（嵌入聊天流）
│       └── quick_actions.dart       # 快捷操作菜单（查数据/下发属性/调服务）
├── android/app/src/main/AndroidManifest.xml
└── ios/Runner/Info.plist
```

## 依赖

```yaml
dependencies:
  flutter:
    sdk: flutter
  http: ^1.2.0               # HTTP 客户端
  speech_to_text: ^6.6.0     # 语音识别（中文）
  just_audio: ^0.9.36        # TTS 播放
  shared_preferences: ^2.2.0 # 持久化（服务器地址/Token）
  provider: ^6.1.0           # 状态管理
```

## 聊天消息模型

```dart
enum MessageType {
  userVoice,      // 用户语音指令
  userText,       // 用户文字指令
  deviceReply,    // 设备语音回复（含TTS）
  deviceData,     // 设备数据卡片
  systemInfo,     // 系统提示（上线/离线/错误）
}

class ChatMessage {
  final String id;
  final MessageType type;
  final String content;       // 文字内容
  final String? audioUrl;     // TTS 音频地址
  final String? action;       // 执行动作描述
  final bool? success;        // 设备回复成功/失败
  final Map<String, dynamic>? data;  // 设备数据/属性
  final DateTime timestamp;
}
```

## API 对接

所有响应格式：`{"code": int, "msg": string, "data": ...}`

认证头：`Authorization: Bearer {jwt_token}`

| 端点 | 方法 | 用途 | 认证 |
|------|------|------|------|
| `/health` | GET | 测试服务器连通性 | 否 |
| `/api/auth/login` | POST | 登录，返回 `{token, user}` | 否 |
| `/api/devices` | GET | 设备列表 | 是 |
| `/api/devices/:id` | GET | 设备详情 | 是 |
| `/api/devices/:id/connection-type` | GET | 连接类型（real/simulated/offline） | 是 |
| `/api/devices/:id/simulate/online` | POST | 模拟上线 | 是 |
| `/api/devices/:id/simulate/heartbeat` | POST | 心跳续期 | 是 |
| `/api/devices/:id/simulate/offline` | POST | 模拟下线 | 是 |
| `/api/devices/:id/voice-debug` | POST | 语音指令，body: `{"text": "打开灯"}` | 是 |
| `/api/devices/:id/data/latest` | GET | 最新遥测数据 | 是 |
| `/api/thing-models/:id` | GET | 物模型详情（属性/服务定义） | 是 |
| `/api/tts/:filename` | GET | TTS 音频文件下载 | 否 |

### API 响应示例

**登录响应**:
```json
{"code": 200, "msg": "success", "data": {"token": "eyJ...", "user": {"id": "uuid", "email": "a@b.com", "role": "user"}}}
```

**语音指令响应**:
```json
{"code": 200, "msg": "success", "data": {"success": true, "message": "指令已执行", "action": "设置属性 switch=true", "audio_url": "/api/tts/abc123.wav"}}
```

**设备列表响应**:
```json
{"code": 200, "msg": "success", "data": {"list": [{"id": "uuid", "name": "客厅灯", "model_id": "uuid", "model_name": "智能灯", "status": "online", ...}], "total": 5, "page": 1, "page_size": 50}}
```

**最新遥测数据响应**:
```json
{"code": 200, "msg": "success", "data": {"time": "2026-03-12T10:30:00Z", "payload": {"temperature": 25.5, "switch": true}, "valid": true}}
```

**连接类型响应**:
```json
{"code": 200, "msg": "success", "data": {"device_id": "uuid", "connection_type": "offline"}}
```

## 关键逻辑

### 首次启动判断
```
启动 → SharedPreferences 读取 serverUrl
  → 有值 → 读取 token → 有效 → chat_screen
                       → 无效 → login_screen
  → 无值 → server_config_screen
```

### 设备上线管理
- 选设备时：查 connection-type → offline 则 simulate/online → 启动 2min 心跳定时器
- 切换设备时：清除旧心跳定时器（旧设备 5min 后自动过期下线）
- APP 退出/暂停时：清除心跳定时器

### 聊天交互流程
1. 用户发语音/文字 → 右侧气泡显示
2. 调 voice-debug API → 等待中显示 typing 动画
3. 收到回复 → 左侧气泡显示结果（成功绿色/失败红色）
4. 有 audio_url → 自动播放 TTS 或显示播放按钮
5. 快捷操作（查数据）→ 插入数据卡片到聊天流

### 快捷操作
- **查看设备数据**: 调 `/devices/:id/data/latest`，结果作为 `deviceData` 卡片插入聊天
- **下发属性**: 弹窗选属性+输入值 → 调 `/devices/:id/debug` (property_set) → 结果作为回复气泡
- **调用服务**: 弹窗选服务 → 调 `/devices/:id/debug` (service_invoke) → 结果作为回复气泡

### 主题
暗色，匹配 LinkFlow：
- Primary: `#2DD4BF`（Teal）
- Background: `#0F172A`
- Surface: `#1E293B`
- 用户气泡: Teal 色
- 设备气泡: Surface 色
- 成功文字: `#4ADE80`
- 失败文字: `#F87171`
- 系统提示: `#94A3B8`

## 实现步骤

1. Flutter 项目脚手架 + pubspec.yaml
2. config.dart + models（user, device, voice_result, chat_message）
3. services（api_service, auth_service, audio_service）
4. server_config_screen（首次启动配置 IP）
5. login_screen
6. chat_screen 骨架 + 消息列表 + 输入栏
7. chat_bubble + device_reply 气泡
8. voice_input_bar（speech_to_text 集成）
9. voice-debug API 对接 + TTS 播放
10. device_picker + 设备上线管理
11. quick_actions + device_data_card
12. 打磨 UI + 错误处理

## 参考文件

- `internal/handler/debug.go` — VoiceDebug 端点
- `internal/model/module.go:43-48` — VoiceResult 结构
- `web/src/api/index.ts` — API 调用参考
- `cmd/server/main.go` — 路由定义
