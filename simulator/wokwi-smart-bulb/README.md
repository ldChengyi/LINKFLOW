# Wokwi 智能灯泡模拟器

用 Wokwi 在线模拟 ESP32 智能灯泡，通过 MQTT 接入 LinkFlow 平台。

## 文件说明

| 文件 | 说明 |
|------|------|
| `sketch.ino` | ESP32 主程序（Arduino 框架） |
| `diagram.json` | Wokwi 电路图（ESP32 + 3 个 LED + 电阻） |

## 引脚说明

| GPIO | 颜色 | 功能 |
|------|------|------|
| 4 | 黄色 LED | 主灯泡亮度（PWM） |
| 16 | 橙色 LED | 暖白色温（PWM） |
| 17 | 蓝色 LED | 冷白色温（PWM） |
| 2 | 内置 LED | MQTT 连接状态（在线=常亮） |

## 物模型配置

在平台创建物模型「智能灯泡」，按以下配置：

**属性：**

| ID | 类型 | 读写 | 范围 / 枚举 | 单位 |
|----|------|------|------------|------|
| switch | bool | rw | — | — |
| brightness | int | rw | min=1, max=100 | % |
| color_temp | enum | rw | 0=暖白2700K, 1=自然白4000K, 2=冷白6500K | — |
| power | float | r | min=0, max=10 | W |

**服务：**

| ID | 输入参数 | 输出参数 |
|----|---------|---------|
| blink | times(int,1~10), interval_ms(int,200~2000) | done(bool) |
| reset | — | — |

**事件：**

| ID | 参数 |
|----|------|
| boot | version(string) |

## 使用步骤

1. 在平台「设备管理」创建设备，绑定上述物模型
2. 复制 **Device ID** 和 **Device Secret**
3. 修改 `sketch.ino` 顶部配置区：
   ```cpp
   #define MQTT_HOST      "your.server.ip"   // 服务器 IP / 域名
   #define DEVICE_ID      "xxx-xxx-xxx"       // 设备 ID
   #define DEVICE_SECRET  "64字符密钥"         // 设备密钥
   ```
4. 在 [wokwi.com](https://wokwi.com) 新建 ESP32 项目
5. 将 `sketch.ino` 和 `diagram.json` 内容分别粘贴到对应文件
6. 点击 Run，查看串口输出

> **本地开发网络穿透**：Wokwi 的 ESP32 可以访问公网，若 MQTT Broker 在本机，
> 需用 `ngrok tcp 1883` 或 frp 暴露到公网后填入 `MQTT_HOST`。

## 数据流

```
Wokwi ESP32                         LinkFlow 平台
    │── CONNECT(clientID, secret) ──>│ 设备上线 + WS 推送
    │── telemetry/up (每10s) ────────>│ 存库 + 告警评估 + WS 推送
    │<── telemetry/down ─────────────│ 下发属性（调试页/语音/定时任务）
    │── telemetry/up (立即确认) ─────>│
    │<── service/invoke ─────────────│ blink / reset
    │── service/reply ───────────────>│ 执行结果
    │── event/boot (首次) ───────────>│ 开机事件
    │<── ota/down ───────────────────│ OTA 命令（上线时自动推送）
    │── ota/up (0%→100%) ────────────>│ 进度 + WS 实时进度条
```
