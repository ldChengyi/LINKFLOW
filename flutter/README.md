# LinkFlow Flutter 语音助手 APP

## 快速开始

### 1. 安装依赖
```bash
cd flutter
flutter pub get
```

### 2. 运行项目
```bash
# Android
flutter run

# iOS (需要 macOS)
flutter run -d ios
```

### 3. 使用流程
1. 首次启动输入后端服务器 IP 地址（如 `192.168.1.100`）
2. 登录 LinkFlow 账号
3. 选择设备开始聊天式语音/文字交互

## 功能特性

- ✅ 聊天式交互界面
- ✅ 语音识别（长按麦克风说话）
- ✅ 文字输入
- ✅ TTS 语音播放
- ✅ 设备选择与状态显示
- ✅ 快捷操作（查看数据/下发属性/调用服务）

## 项目结构

```
lib/
├── main.dart                 # 应用入口
├── config.dart               # 配置常量
├── models/                   # 数据模型
├── services/                 # API 服务层
├── screens/                  # 页面
└── widgets/                  # 组件
```

## API 对接

后端 API 基础地址：`http://{IP}:{PORT}`

主要端点：
- `POST /api/auth/login` - 登录
- `GET /api/devices` - 设备列表
- `POST /api/devices/:id/voice-debug` - 语音指令

详见 `Plan.md`
