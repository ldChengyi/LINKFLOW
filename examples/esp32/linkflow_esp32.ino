/**
 * LinkFlow IoT Platform — ESP32 接入示例
 *
 * 演示功能：
 *   - WiFi 连接 + MQTT 接入（自动重连）
 *   - 遥测数据定时上报（温度 / 湿度 / 开关 / 亮度 / 模式）
 *   - 属性下发处理（平台 → 设备）
 *   - 服务调用处理 + 结果回传（reboot / reset）
 *   - 语音控制结果接收
 *
 * 硬件依赖（示例，可按实际修改）：
 *   - LED    → GPIO 2（内置 LED，模拟开关状态）
 *   - Relay  → GPIO 26（继电器，真实开关控制）
 *   - DHT22  → GPIO 4（温湿度传感器，可选）
 *
 * Arduino 库依赖（库管理器安装）：
 *   - PubSubClient  by Nick O'Leary  (MQTT)
 *   - ArduinoJson   by Benoit Blanchon
 *   - DHT sensor library by Adafruit（使用真实传感器时安装）
 *
 * 快速开始：
 *   1. 在 LinkFlow 平台创建物模型（参考下方「物模型属性定义」）
 *   2. 创建设备并绑定物模型，复制设备 ID 和密钥
 *   3. 修改下方「配置区」的 WiFi 和设备信息
 *   4. 烧录到 ESP32，打开串口监视器（115200 波特率）
 */

#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>

// ================================================================
// 配置区 — 必须修改
// ================================================================

// WiFi
#define WIFI_SSID       "your_wifi_ssid"
#define WIFI_PASSWORD   "your_wifi_password"

// LinkFlow 服务器（填写运行平台的机器 IP，不能用 localhost）
#define MQTT_HOST       "192.168.1.100"
#define MQTT_PORT       1883

// 设备凭证（从 LinkFlow 设备管理页面复制）
#define DEVICE_ID       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
#define DEVICE_SECRET   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

// ================================================================
// 物模型属性定义（与平台保持一致）
// ================================================================
//
//  标识符        名称      类型    读写  范围          单位
//  switch        开关      bool    rw
//  brightness    亮度      int     rw    0 ~ 100       %
//  mode          模式      enum    rw    0=正常 1=节能 2=强光
//  temperature   温度      float   r                   ℃
//  humidity      湿度      float   r     0 ~ 100       %
//
// 服务：
//  reboot        重启设备
//  reset         重置所有属性到默认值

// ================================================================
// GPIO 定义
// ================================================================

#define PIN_LED    2    // 内置 LED，反映 switch 状态
#define PIN_RELAY  26   // 继电器，实际开关控制
// #define PIN_DHT  4   // 取消注释以启用 DHT22

// ================================================================
// 遥测上报间隔
// ================================================================

const unsigned long TELEMETRY_INTERVAL_MS  = 5000;   // 5 秒上报一次
const unsigned long RECONNECT_INTERVAL_MS  = 5000;   // 断线重连间隔

// ================================================================
// 全局对象与状态
// ================================================================

WiFiClient   wifiClient;
PubSubClient mqtt(wifiClient);

// 设备属性状态（与平台同步）
bool  propSwitch     = false;
int   propBrightness = 50;
int   propMode       = 0;
float propTemperature = 25.0f;
float propHumidity    = 60.0f;

// MQTT Topic 缓存
char topicTelemetryUp[128];
char topicTelemetryDown[128];
char topicServiceInvoke[128];
char topicServiceReply[128];
char topicVoiceUp[128];
char topicVoiceDown[128];

unsigned long lastTelemetryMs  = 0;
unsigned long lastReconnectMs  = 0;

// ================================================================
// 工具：生成 Topic 字符串
// ================================================================

void initTopics() {
    snprintf(topicTelemetryUp,   sizeof(topicTelemetryUp),   "devices/%s/telemetry/up",   DEVICE_ID);
    snprintf(topicTelemetryDown, sizeof(topicTelemetryDown), "devices/%s/telemetry/down", DEVICE_ID);
    snprintf(topicServiceInvoke, sizeof(topicServiceInvoke), "devices/%s/service/invoke", DEVICE_ID);
    snprintf(topicServiceReply,  sizeof(topicServiceReply),  "devices/%s/service/reply",  DEVICE_ID);
    snprintf(topicVoiceUp,       sizeof(topicVoiceUp),       "devices/%s/voice/up",       DEVICE_ID);
    snprintf(topicVoiceDown,     sizeof(topicVoiceDown),     "devices/%s/voice/down",     DEVICE_ID);
}

// ================================================================
// 读取传感器（替换为真实传感器驱动）
// ================================================================

void readSensors() {
#ifdef PIN_DHT
    // 使用真实 DHT22：
    // #include <DHT.h>
    // DHT dht(PIN_DHT, DHT22);
    // float t = dht.readTemperature();
    // float h = dht.readHumidity();
    // if (!isnan(t)) propTemperature = t;
    // if (!isnan(h)) propHumidity    = h;
#else
    // 模拟传感器数据（用于测试）
    propTemperature = 22.0f + (float)(random(0, 60)) / 10.0f;  // 22.0 ~ 28.0
    propHumidity    = 55.0f + (float)(random(0, 100)) / 10.0f; // 55.0 ~ 65.0
#endif
}

// ================================================================
// 遥测上报
// ================================================================

void publishTelemetry() {
    readSensors();

    StaticJsonDocument<256> doc;
    doc["switch"]      = propSwitch;
    doc["brightness"]  = propBrightness;
    doc["mode"]        = propMode;
    doc["temperature"] = round(propTemperature * 10) / 10.0;  // 保留 1 位小数
    doc["humidity"]    = round(propHumidity    * 10) / 10.0;

    char buf[256];
    serializeJson(doc, buf);

    if (mqtt.publish(topicTelemetryUp, buf, false)) {
        Serial.printf("[遥测] 上报成功: %s\n", buf);
    } else {
        Serial.println("[遥测] 上报失败（MQTT 未连接？）");
    }
}

// ================================================================
// 应用属性值到硬件
// ================================================================

void applyProperties() {
    digitalWrite(PIN_LED,   propSwitch ? HIGH : LOW);
    digitalWrite(PIN_RELAY, propSwitch ? HIGH : LOW);

    // 若有 PWM 调光：
    // ledcWrite(0, map(propBrightness, 0, 100, 0, 255));

    switch (propMode) {
        case 0: Serial.println("[模式] 正常");  break;
        case 1: Serial.println("[模式] 节能");  break;
        case 2: Serial.println("[模式] 强光");  break;
        default: break;
    }
}

// ================================================================
// 处理属性下发（平台 → 设备）
// ================================================================

void handlePropertySet(JsonDocument& doc) {
    bool changed = false;

    if (doc.containsKey("switch")) {
        propSwitch = doc["switch"].as<bool>();
        Serial.printf("[属性] switch = %s\n", propSwitch ? "true" : "false");
        changed = true;
    }
    if (doc.containsKey("brightness")) {
        propBrightness = constrain(doc["brightness"].as<int>(), 0, 100);
        Serial.printf("[属性] brightness = %d\n", propBrightness);
        changed = true;
    }
    if (doc.containsKey("mode")) {
        propMode = doc["mode"].as<int>();
        Serial.printf("[属性] mode = %d\n", propMode);
        changed = true;
    }

    if (changed) {
        applyProperties();
        // 属性变更后立即上报，让平台数据与设备实际状态一致
        publishTelemetry();
    }
}

// ================================================================
// 服务调用回复
// ================================================================

void replyService(const char* id, const char* service, bool success, const char* message) {
    StaticJsonDocument<256> doc;
    doc["id"]      = id;
    doc["service"] = service;
    doc["code"]    = success ? 200 : 500;
    doc["message"] = message;

    char buf[256];
    serializeJson(doc, buf);
    mqtt.publish(topicServiceReply, buf, false);
    Serial.printf("[服务] 回复: %s\n", buf);
}

// ================================================================
// 处理服务调用（平台 → 设备）
// ================================================================

void handleServiceInvoke(JsonDocument& doc) {
    const char* id      = doc["id"]      | "unknown";
    const char* service = doc["service"] | "";

    Serial.printf("[服务] 收到调用: id=%s, service=%s\n", id, service);

    if (strcmp(service, "reboot") == 0) {
        replyService(id, service, true, "设备即将重启");
        delay(300);
        ESP.restart();

    } else if (strcmp(service, "reset") == 0) {
        propSwitch     = false;
        propBrightness = 50;
        propMode       = 0;
        applyProperties();
        publishTelemetry();
        replyService(id, service, true, "属性已重置为默认值");

    } else {
        replyService(id, service, false, "未知服务");
    }
}

// ================================================================
// 处理语音控制结果（平台 → 设备，可选）
// ================================================================

void handleVoiceResult(JsonDocument& doc) {
    bool        success = doc["success"] | false;
    const char* message = doc["message"] | "";
    const char* action  = doc["action"]  | "";

    if (success) {
        Serial.printf("[语音] 执行成功: %s\n", action);
        // 语音成功执行后，平台会通过 telemetry/down 推送新属性值
        // 此处可做 UI 反馈（如蜂鸣器提示音）
    } else {
        Serial.printf("[语音] 执行失败: %s\n", message);
    }
}

// ================================================================
// 可选：发送语音文本（若设备集成了麦克风 + 语音识别模组）
// ================================================================

void publishVoiceText(const char* text) {
    StaticJsonDocument<128> doc;
    doc["text"] = text;

    char buf[128];
    serializeJson(doc, buf);
    mqtt.publish(topicVoiceUp, buf, false);
    Serial.printf("[语音] 发送文本: %s\n", text);
}

// ================================================================
// MQTT 消息统一回调
// ================================================================

void onMqttMessage(char* topic, byte* payload, unsigned int length) {
    // 解析 JSON
    StaticJsonDocument<512> doc;
    DeserializationError err = deserializeJson(doc, payload, length);
    if (err) {
        Serial.printf("[MQTT] JSON 解析失败 (%s): %s\n", topic, err.c_str());
        return;
    }

    if (strcmp(topic, topicTelemetryDown) == 0) {
        handlePropertySet(doc);
    } else if (strcmp(topic, topicServiceInvoke) == 0) {
        handleServiceInvoke(doc);
    } else if (strcmp(topic, topicVoiceDown) == 0) {
        handleVoiceResult(doc);
    }
}

// ================================================================
// WiFi 连接
// ================================================================

void connectWiFi() {
    if (WiFi.status() == WL_CONNECTED) return;

    Serial.printf("\n[WiFi] 连接 %s ...", WIFI_SSID);
    WiFi.mode(WIFI_STA);
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

    unsigned long start = millis();
    while (WiFi.status() != WL_CONNECTED && millis() - start < 15000) {
        delay(500);
        Serial.print(".");
    }
    Serial.println();

    if (WiFi.status() == WL_CONNECTED) {
        Serial.printf("[WiFi] 已连接，IP: %s\n", WiFi.localIP().toString().c_str());
    } else {
        Serial.println("[WiFi] 连接超时，将在下一轮重试");
    }
}

// ================================================================
// MQTT 连接
// ================================================================

bool connectMQTT() {
    if (mqtt.connected()) return true;

    Serial.printf("[MQTT] 连接 %s:%d ...\n", MQTT_HOST, MQTT_PORT);

    // clientID = DEVICE_ID, username = DEVICE_ID, password = DEVICE_SECRET
    if (!mqtt.connect(DEVICE_ID, DEVICE_ID, DEVICE_SECRET)) {
        Serial.printf("[MQTT] 连接失败，state=%d\n", mqtt.state());
        // state 说明：
        //  -4  连接超时
        //  -3  连接断开
        //  -2  连接失败
        //  5   未授权（DEVICE_ID / DEVICE_SECRET 错误）
        return false;
    }

    Serial.println("[MQTT] 连接成功");

    // 订阅平台下发 Topic（QoS 1）
    mqtt.subscribe(topicTelemetryDown, 1);
    mqtt.subscribe(topicServiceInvoke, 1);
    mqtt.subscribe(topicVoiceDown,     1);

    Serial.printf("[MQTT] 已订阅:\n  %s\n  %s\n  %s\n",
        topicTelemetryDown,
        topicServiceInvoke,
        topicVoiceDown
    );

    // 上线后立即上报一次当前状态
    publishTelemetry();
    return true;
}

// ================================================================
// Arduino 入口
// ================================================================

void setup() {
    Serial.begin(115200);
    delay(200);

    Serial.println("\n========================================");
    Serial.println("  LinkFlow ESP32 示例");
    Serial.printf("  设备 ID: %s\n", DEVICE_ID);
    Serial.println("========================================");

    // GPIO 初始化
    pinMode(PIN_LED,   OUTPUT);
    pinMode(PIN_RELAY, OUTPUT);
    digitalWrite(PIN_LED,   LOW);
    digitalWrite(PIN_RELAY, LOW);

    // 生成 Topic 字符串
    initTopics();

    // MQTT 配置
    mqtt.setServer(MQTT_HOST, MQTT_PORT);
    mqtt.setCallback(onMqttMessage);
    mqtt.setBufferSize(1024);          // 增大缓冲区防止大 JSON 被截断
    mqtt.setKeepAlive(60);             // 60 秒心跳

    // 初始连接
    connectWiFi();
    connectMQTT();
}

void loop() {
    unsigned long now = millis();

    // 维持连接
    if (!mqtt.connected()) {
        if (now - lastReconnectMs > RECONNECT_INTERVAL_MS) {
            lastReconnectMs = now;
            if (WiFi.status() != WL_CONNECTED) connectWiFi();
            connectMQTT();
        }
    } else {
        mqtt.loop();  // 处理 MQTT 收发

        // 定时上报遥测
        if (now - lastTelemetryMs > TELEMETRY_INTERVAL_MS) {
            lastTelemetryMs = now;
            publishTelemetry();
        }
    }

    // ────────────────────────────────────────────────
    // 可在此添加本地逻辑，例如：
    //   - 读取按键，手动切换 propSwitch 并调用 publishTelemetry()
    //   - 麦克风触发后调用 publishVoiceText("打开灯")
    //   - 告警阈值检测（温度超过 35℃ 时做本地告警）
    // ────────────────────────────────────────────────
}
