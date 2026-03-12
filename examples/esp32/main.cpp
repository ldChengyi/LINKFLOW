// =============================================================================
// LinkFlow IoT Platform — ESP32 Client Example
// =============================================================================
// 功能覆盖:
//   - WiFi 连接 + 自动重连
//   - MQTT 认证 (device_id + device_secret)
//   - 遥测上报 (telemetry/up)   — DHT11 温湿度 + WS2812 灯带状态
//   - 属性下发 (telemetry/down) — 接收并执行属性设置，回传确认 + 确认音
//   - 服务调用 (service/invoke) — 接收并执行服务，回传 reply
//   - OTA 升级 (ota/down)       — HTTP 下载固件 + 进度上报 + 自动刷写
//   - 语音结果 (voice/down)     — 接收语音执行结果 + 成功/失败音效
//   - LED 模拟控制              — 板载 LED 演示 bool 属性开关
//   - WS2812 灯带               — 8色 + 彩虹动画，支持语音/平台控制
//   - 音频反馈                  — ES8311 + NS4150B 确认音/成功音/失败音
//
// 需要的库:
//   PubSubClient, ArduinoJson (v7), DHT sensor library, Adafruit Unified Sensor,
//   Adafruit NeoPixel
//
// 使用方式:
//   1. 在 LinkFlow 平台创建物模型（属性: temperature, humidity, led_switch）
//   2. 创建设备并绑定物模型
//   3. 复制 device_id 和 device_secret 填入下方配置
//   4. 修改 WiFi 和 MQTT Broker 地址
//   5. 烧录到 ESP32
// =============================================================================

#include <Arduino.h>
#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <DHT.h>
#include <DHT_U.h>
#include <HTTPClient.h>
#include <Update.h>
#include "audio.h"
#include "ws2812.h"
#include "tts_player.h"

// ======================== 用户配置区 ========================

// WiFi
const char *WIFI_SSID = "hqm";
const char *WIFI_PASS = "chengs666";

// MQTT Broker (LinkFlow 后端地址)
const char *MQTT_HOST = "111.228.58.69"; // 改为你的 LinkFlow 服务器 IP
const uint16_t MQTT_PORT = 1883;

// 设备凭证 (从 LinkFlow 平台复制)
const char *DEVICE_ID = "8943954d-4c53-4b84-a1fe-c553c980e782";                                 // 设备 ID (UUID)
const char *DEVICE_SECRET = "3c1804107e2a9f82ea3230ca1db43d4977cd8632a93232413f6e0aadef2d4be4"; // 设备密钥 (64字符)

// DHT 传感器
#define DHT_PIN 4
#define DHT_TYPE DHT11

// LED (模拟开关属性)
#define LED_PIN 2 // ESP32 板载 LED，部分板子是 GPIO2

// 遥测上报间隔 (毫秒)
#define TELEMETRY_INTERVAL 10000 // 10 秒

// MQTT 重连间隔 (毫秒)
#define MQTT_RECONNECT_INTERVAL 5000

// ======================== 全局变量 ========================

WiFiClient wifiClient;
PubSubClient mqtt(wifiClient);
DHT_Unified dht(DHT_PIN, DHT_TYPE);
AudioModule audio;
WS2812Module ledStrip;
TTSPlayer ttsPlayer;

// Topic 缓存 (避免反复拼接)
String topicTelemetryUp;
String topicTelemetryDown;
String topicServiceInvoke;
String topicServiceReply;
String topicVoiceDown;
String topicOtaDown;
String topicOtaUp;

// 设备状态
bool ledState = false;
bool buzzerEnabled = true;
float lastTemperature = NAN;
float lastHumidity = NAN;
String firmwareVersion = "1.0.0";

// 计时器
unsigned long lastTelemetryTime = 0;
unsigned long lastMqttReconnect = 0;

// OTA 状态
bool otaInProgress = false;
String otaTaskId;

// ======================== 函数声明 ========================

void setupWiFi();
void setupTopics();
void setupMQTT();
void mqttCallback(char *topic, byte *payload, unsigned int length);
void mqttReconnect();
void readSensors();
void publishTelemetry();
void handleTelemetryDown(JsonDocument &doc);
void handleServiceInvoke(JsonDocument &doc);
void handleOtaDown(JsonDocument &doc);
void handleVoiceDown(JsonDocument &doc);
void reportOtaProgress(const char *status, int progress, const char *error, const char *version);
void performOTA(const char *url, const char *checksum, const char *taskId, const char *version);

// ======================== Setup & Loop ========================

void setup()
{
    Serial.begin(115200);
    Serial.println();
    Serial.println(F("=== LinkFlow ESP32 Client ==="));
    Serial.printf("Firmware: v%s\n", firmwareVersion.c_str());

    // 打印模块 & GPIO 分配
    Serial.println(F(""));
    Serial.println(F("========== Modules & GPIO =========="));
    Serial.println(F(" [Board LED]    GPIO 2   — 板载 LED"));
    Serial.println(F(" [DHT11]        GPIO 4   — 温湿度传感器"));
    Serial.println(F(" [ES8311 I2C]   GPIO 14  — I2C SDA"));
    Serial.println(F("                GPIO 27  — I2C SCL"));
    Serial.println(F(" [ES8311 I2S]   GPIO 22  — I2S DIN (数据)"));
    Serial.println(F("                GPIO 25  — I2S LRCK (左右声道)"));
    Serial.println(F("                GPIO 26  — I2S SCLK (时钟)"));
    Serial.println(F(" [WS2812]       GPIO 15  — 灯带 DIN"));
    Serial.println(F("===================================="));
    Serial.println(F(" Libraries required:"));
    Serial.println(F("   - PubSubClient"));
    Serial.println(F("   - ArduinoJson (v7)"));
    Serial.println(F("   - DHT sensor library"));
    Serial.println(F("   - Adafruit Unified Sensor"));
    Serial.println(F("   - Adafruit NeoPixel"));
    Serial.println(F("===================================="));
    Serial.println(F(""));

    // LED
    pinMode(LED_PIN, OUTPUT);
    digitalWrite(LED_PIN, LOW);

    // DHT
    dht.begin();
    sensor_t sensor;
    dht.temperature().getSensor(&sensor);
    Serial.printf("DHT Sensor: %s, min_delay=%ldus\n", sensor.name, sensor.min_delay);

    // Audio — 初始化并默认启用
    if (audio.begin())
    {
        audio.enable();
        audio.playBeep(); // 开机提示音
        Serial.println(F("[Audio] Module ready, enabled by default"));
    }
    else
    {
        Serial.println(F("[Audio] Module init FAILED — check I2C/I2S wiring"));
    }

    // TTS Player
    ttsPlayer.begin(&audio);

    // WS2812 LED Strip — 初始化并默认点亮
    if (ledStrip.begin())
    {
        ledStrip.setSwitch(true);
        ledStrip.setBrightness(128);
        ledStrip.setColor(COLOR_CYAN);
        Serial.println(F("[WS2812] Strip ON, color=CYAN, brightness=128"));
    }
    else
    {
        Serial.println(F("[WS2812] Strip init FAILED — check DIN wiring on GPIO 15"));
    }

    // WiFi
    setupWiFi();

    // Topic
    setupTopics();

    // MQTT
    setupMQTT();
}

void loop()
{
    // WiFi 断线重连
    if (WiFi.status() != WL_CONNECTED)
    {
        Serial.println(F("[WiFi] Disconnected, reconnecting..."));
        setupWiFi();
    }

    // MQTT 断线重连
    if (!mqtt.connected())
    {
        unsigned long now = millis();
        if (now - lastMqttReconnect >= MQTT_RECONNECT_INTERVAL)
        {
            lastMqttReconnect = now;
            mqttReconnect();
        }
    }

    mqtt.loop();

    // WS2812 彩虹动画更新
    ledStrip.update();

    // OTA 进行中不上报遥测
    if (otaInProgress)
        return;

    // 定时上报遥测
    unsigned long now = millis();
    if (now - lastTelemetryTime >= TELEMETRY_INTERVAL)
    {
        lastTelemetryTime = now;
        readSensors();
        publishTelemetry();
    }
}

// ======================== WiFi ========================

void setupWiFi()
{
    Serial.printf("[WiFi] Connecting to %s", WIFI_SSID);
    WiFi.mode(WIFI_STA);
    WiFi.begin(WIFI_SSID, WIFI_PASS);

    int attempts = 0;
    while (WiFi.status() != WL_CONNECTED && attempts < 40)
    {
        delay(500);
        Serial.print(".");
        attempts++;
    }

    if (WiFi.status() == WL_CONNECTED)
    {
        Serial.printf("\n[WiFi] Connected! IP: %s\n", WiFi.localIP().toString().c_str());
    }
    else
    {
        Serial.println(F("\n[WiFi] Connection failed, will retry in loop"));
    }
}

// ======================== MQTT Topics ========================

void setupTopics()
{
    String prefix = "devices/" + String(DEVICE_ID);
    topicTelemetryUp = prefix + "/telemetry/up";
    topicTelemetryDown = prefix + "/telemetry/down";
    topicServiceInvoke = prefix + "/service/invoke";
    topicServiceReply = prefix + "/service/reply";
    topicVoiceDown = prefix + "/voice/down";
    topicOtaDown = prefix + "/ota/down";
    topicOtaUp = prefix + "/ota/up";
}

// ======================== MQTT ========================

void setupMQTT()
{
    mqtt.setServer(MQTT_HOST, MQTT_PORT);
    mqtt.setCallback(mqttCallback);
    mqtt.setBufferSize(1024); // OTA 消息可能较大
    mqttReconnect();
}

void mqttReconnect()
{
    if (mqtt.connected())
        return;

    Serial.printf("[MQTT] Connecting as %s ...\n", DEVICE_ID);

    // LinkFlow 认证: clientID = device_id, password = device_secret
    // username 不使用，传空字符串
    if (mqtt.connect(DEVICE_ID, "", DEVICE_SECRET))
    {
        Serial.println(F("[MQTT] Connected!"));

        // 订阅下行 Topic
        mqtt.subscribe(topicTelemetryDown.c_str());
        mqtt.subscribe(topicServiceInvoke.c_str());
        mqtt.subscribe(topicVoiceDown.c_str());
        mqtt.subscribe(topicOtaDown.c_str());

        Serial.println(F("[MQTT] Subscribed to downlink topics"));
    }
    else
    {
        Serial.printf("[MQTT] Failed, rc=%d\n", mqtt.state());
        // PubSubClient 状态码:
        // -4 连接超时, -3 连接丢失, -2 连接失败
        // -1 断开, 4 认证失败, 5 未授权
    }
}

// ======================== MQTT 消息回调 ========================

void mqttCallback(char *topic, byte *payload, unsigned int length)
{
    // 安全拷贝 payload
    char *buf = (char *)malloc(length + 1);
    if (!buf)
        return;
    memcpy(buf, payload, length);
    buf[length] = '\0';

    Serial.printf("[MQTT] << %s : %s\n", topic, buf);

    JsonDocument doc;
    DeserializationError err = deserializeJson(doc, buf);
    free(buf);

    if (err)
    {
        Serial.printf("[MQTT] JSON parse error: %s\n", err.c_str());
        return;
    }

    String topicStr(topic);

    if (topicStr == topicTelemetryDown)
    {
        handleTelemetryDown(doc);
    }
    else if (topicStr == topicServiceInvoke)
    {
        handleServiceInvoke(doc);
    }
    else if (topicStr == topicOtaDown)
    {
        handleOtaDown(doc);
    }
    else if (topicStr == topicVoiceDown)
    {
        handleVoiceDown(doc);
    }
}

// ======================== 遥测上报 ========================

void readSensors()
{
    sensors_event_t event;

    dht.temperature().getEvent(&event);
    if (!isnan(event.temperature))
    {
        lastTemperature = event.temperature;
    }
    else
    {
        Serial.println(F("[DHT] Temperature read error"));
    }

    dht.humidity().getEvent(&event);
    if (!isnan(event.relative_humidity))
    {
        lastHumidity = event.relative_humidity;
    }
    else
    {
        Serial.println(F("[DHT] Humidity read error"));
    }
}

void publishTelemetry()
{
    if (!mqtt.connected())
        return;

    JsonDocument doc;

    // 上报传感器数据
    if (!isnan(lastTemperature))
        doc["temperature"] = lastTemperature;
    if (!isnan(lastHumidity))
        doc["humidity"] = lastHumidity;

    // 上报 LED 开关状态
    doc["led_switch"] = ledState;
    doc["buzzer_enabled"] = buzzerEnabled;

    // 上报灯带状态
    doc["strip_switch"] = ledStrip.isOn();
    doc["led_color"] = ledStrip.getColor();
    doc["strip_brightness"] = ledStrip.getBrightness();

    char payload[384];
    serializeJson(doc, payload, sizeof(payload));

    mqtt.publish(topicTelemetryUp.c_str(), payload);
    Serial.printf("[MQTT] >> telemetry/up : %s\n", payload);
}

// ======================== 属性下发处理 ========================

void handleTelemetryDown(JsonDocument &doc)
{
    Serial.println(F("[CMD] Property set received"));

    bool changed = false;

    // 处理 LED 开关
    if (doc.containsKey("led_switch"))
    {
        ledState = doc["led_switch"].as<bool>();
        digitalWrite(LED_PIN, ledState ? HIGH : LOW);
        Serial.printf("[CMD] led_switch = %s\n", ledState ? "ON" : "OFF");
        changed = true;
    }

    // 处理蜂鸣器
    if (doc.containsKey("buzzer_enabled"))
    {
        buzzerEnabled = doc["buzzer_enabled"].as<bool>();
        if (buzzerEnabled)
        {
            audio.enable();
            audio.playTone(1000, 200); // 1kHz, 200ms
        }
        else
        {
            audio.disable();
        }
        Serial.printf("[CMD] buzzer_enabled = %s\n", buzzerEnabled ? "ON" : "OFF");
        changed = true;
    }

    // 处理灯带开关
    if (doc.containsKey("strip_switch"))
    {
        ledStrip.setSwitch(doc["strip_switch"].as<bool>());
        changed = true;
    }

    // 处理灯带颜色
    if (doc.containsKey("led_color"))
    {
        ledStrip.setColor(doc["led_color"].as<uint8_t>());
        changed = true;
    }

    // 处理灯带亮度
    if (doc.containsKey("strip_brightness"))
    {
        ledStrip.setBrightness(doc["strip_brightness"].as<uint8_t>());
        changed = true;
    }

    // 回传确认: 设备必须通过 telemetry/up 回传当前状态
    // 这样 LinkFlow 才能记录属性已生效
    if (changed)
    {
        audio.playBeep(); // 属性变更确认音
        publishTelemetry();
    }
}

// ======================== 服务调用处理 ========================

void handleServiceInvoke(JsonDocument &doc)
{
    const char *id = doc["id"];
    const char *service = doc["service"];

    Serial.printf("[CMD] Service invoke: %s (id=%s)\n", service, id);

    JsonDocument reply;
    reply["id"] = id;
    reply["service"] = service;

    if (strcmp(service, "reboot") == 0)
    {
        // 先回复再重启
        reply["code"] = 200;
        reply["message"] = "rebooting";

        char payload[256];
        serializeJson(reply, payload, sizeof(payload));
        mqtt.publish(topicServiceReply.c_str(), payload);
        mqtt.loop(); // 确保 reply 发出
        Serial.println(F("[CMD] Rebooting..."));

        mqtt.disconnect(); // 发送 DISCONNECT 包，让 broker 正确处理下线
        delay(500);
        ESP.restart();
    }
    else if (strcmp(service, "get_info") == 0)
    {
        // 返回设备信息
        reply["code"] = 200;
        reply["message"] = "success";
        JsonObject output = reply["output"].to<JsonObject>();
        output["firmware"] = firmwareVersion;
        output["free_heap"] = ESP.getFreeHeap();
        output["uptime_ms"] = millis();
        output["wifi_rssi"] = WiFi.RSSI();

        char payload[512];
        serializeJson(reply, payload, sizeof(payload));
        mqtt.publish(topicServiceReply.c_str(), payload);
        Serial.println(F("[CMD] Device info sent"));
    }
    else if (strcmp(service, "play_tone") == 0)
    {
        // 播放指定频率和时长的提示音
        int frequency = doc["params"]["frequency"] | 1000;
        int duration = doc["params"]["duration"] | 200;

        audio.enable();
        audio.playTone((uint16_t)frequency, (uint16_t)duration);
        audio.disable();

        reply["code"] = 200;
        reply["message"] = "tone played";

        char payload[256];
        serializeJson(reply, payload, sizeof(payload));
        mqtt.publish(topicServiceReply.c_str(), payload);
        Serial.printf("[CMD] play_tone: %dHz %dms\n", frequency, duration);
    }
    else
    {
        // 未知服务
        reply["code"] = 404;
        reply["message"] = "unknown service";

        char payload[256];
        serializeJson(reply, payload, sizeof(payload));
        mqtt.publish(topicServiceReply.c_str(), payload);
        Serial.printf("[CMD] Unknown service: %s\n", service);
    }
}

// ======================== OTA 升级处理 ========================

void handleOtaDown(JsonDocument &doc)
{
    const char *taskId = doc["task_id"];
    const char *version = doc["version"];
    const char *url = doc["url"];
    const char *checksum = doc["checksum"];
    int size = doc["size"] | 0;

    Serial.println(F("=== OTA Update ==="));
    Serial.printf("  Task:     %s\n", taskId);
    Serial.printf("  Version:  %s\n", version);
    Serial.printf("  URL:      %s\n", url);
    Serial.printf("  Checksum: %s\n", checksum);
    Serial.printf("  Size:     %d bytes\n", size);

    // 检查剩余空间
    if (size > 0 && (size_t)size > ESP.getFreeSketchSpace())
    {
        Serial.println(F("[OTA] Not enough flash space!"));
        reportOtaProgress("failed", 0, "insufficient flash space", "");
        return;
    }

    otaInProgress = true;
    otaTaskId = String(taskId);

    performOTA(url, checksum, taskId, version);
}

void reportOtaProgress(const char *status, int progress, const char *error, const char *version)
{
    if (!mqtt.connected())
        return;

    JsonDocument doc;
    doc["task_id"] = otaTaskId;
    doc["status"] = status;
    doc["progress"] = progress;
    doc["error"] = error;
    doc["version"] = version;

    char payload[256];
    serializeJson(doc, payload, sizeof(payload));
    mqtt.publish(topicOtaUp.c_str(), payload);
    mqtt.loop(); // 确保消息发出
    Serial.printf("[OTA] Progress: %s %d%%\n", status, progress);
}

void performOTA(const char *url, const char *checksum, const char *taskId, const char *version)
{
    HTTPClient http;
    http.begin(url);

    // Basic Auth: device_id:device_secret
    http.setAuthorization(DEVICE_ID, DEVICE_SECRET);
    http.setTimeout(30000);

    reportOtaProgress("downloading", 0, "", "");

    int httpCode = http.GET();
    if (httpCode != HTTP_CODE_OK)
    {
        Serial.printf("[OTA] HTTP error: %d\n", httpCode);
        String err = "HTTP " + String(httpCode);
        reportOtaProgress("failed", 0, err.c_str(), "");
        otaInProgress = false;
        http.end();
        return;
    }

    int contentLength = http.getSize();
    if (contentLength <= 0)
    {
        reportOtaProgress("failed", 0, "empty response", "");
        otaInProgress = false;
        http.end();
        return;
    }

    // 开始 OTA 写入
    if (!Update.begin(contentLength))
    {
        reportOtaProgress("failed", 0, "Update.begin failed", "");
        otaInProgress = false;
        http.end();
        return;
    }

    reportOtaProgress("downloading", 5, "", "");

    WiFiClient *stream = http.getStreamPtr();
    uint8_t buf[1024];
    int totalRead = 0;
    int lastReportedPercent = 5;

    while (http.connected() && totalRead < contentLength)
    {
        int available = stream->available();
        if (available > 0)
        {
            int readLen = stream->readBytes(buf, min((int)sizeof(buf), available));
            Update.write(buf, readLen);
            totalRead += readLen;

            // 每 10% 上报一次进度
            int percent = (totalRead * 100) / contentLength;
            if (percent >= lastReportedPercent + 10)
            {
                lastReportedPercent = (percent / 10) * 10;
                reportOtaProgress("downloading", lastReportedPercent, "", "");
            }
        }
        delay(1);
    }

    reportOtaProgress("verifying", 90, "", "");

    // TODO: 这里可以对比 checksum (SHA256)
    // 需要 mbedtls 库计算下载数据的 SHA256 并与 checksum 参数对比
    // 简化示例跳过校验

    if (Update.end(true))
    {
        reportOtaProgress("flashing", 95, "", "");

        if (Update.isFinished())
        {
            Serial.println(F("[OTA] Success! Rebooting..."));
            reportOtaProgress("completed", 100, "", version);
            mqtt.disconnect(); // 发送 DISCONNECT 包，让 broker 正确处理下线
            delay(500);
            ESP.restart();
        }
        else
        {
            reportOtaProgress("failed", 95, "Update not finished", "");
            otaInProgress = false;
        }
    }
    else
    {
        String err = "Update error: " + String(Update.getError());
        reportOtaProgress("failed", 90, err.c_str(), "");
        otaInProgress = false;
    }

    http.end();
}

// ======================== 语音结果处理 ========================

void handleVoiceDown(JsonDocument &doc)
{
    bool success = doc["success"] | false;
    const char *message = doc["message"] | "no message";
    const char *action = doc["action"] | "";
    const char *audioURL = doc["audio_url"] | "";

    Serial.printf("[Voice] %s: %s\n", success ? "OK" : "FAIL", message);
    if (strlen(action) > 0)
    {
        Serial.printf("[Voice] Action: %s\n", action);
    }

    // TTS 音频播放（有 audio_url 时用 TTS，否则用提示音）
    if (strlen(audioURL) > 0)
    {
        ttsPlayer.play(audioURL);
    }
    else
    {
        // 无 TTS 时用简单音效反馈
        if (success)
        {
            audio.playSuccessTone();
        }
        else
        {
            audio.playErrorTone();
        }
    }
}
