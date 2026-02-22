/**
 * ╔══════════════════════════════════════════════════╗
 * ║   LinkFlow IoT Platform — Smart Bulb Simulator  ║
 * ║   Wokwi ESP32 模拟智能灯泡                        ║
 * ╚══════════════════════════════════════════════════╝
 *
 * 依赖库（通过 Arduino Library Manager 安装）：
 *   - PubSubClient   by Nick O'Leary
 *   - ArduinoJson    by Benoit Blanchon
 *
 * 【第一步：在平台创建物模型「智能灯泡」】
 *
 * 属性 (Properties):
 *   ID           名称    类型    读写  范围/枚举                      单位
 *   switch       开关    bool   rw    -                              -
 *   brightness   亮度    int    rw    min=1, max=100                  %
 *   color_temp   色温    enum   rw    0=暖白2700K, 1=自然白4000K,
 *                                    2=冷白6500K                      -
 *   power        功耗    float  r     min=0, max=10                   W
 *
 * 服务 (Services):
 *   ID      名称      输入参数                                  输出参数
 *   blink   闪烁测试  times(int,1~10), interval_ms(int,200~2000)  done(bool)
 *   reset   恢复出厂  -                                           -
 *
 * 事件 (Events):
 *   ID      名称  参数
 *   boot    开机  version(string)
 *
 * 【第二步：创建设备，绑定上述物模型，复制 Device ID 和 Device Secret】
 * 【第三步：修改下方配置区域，在 Wokwi 中运行】
 *
 * 引脚说明：
 *   GPIO  4 → 黄色 LED（主灯泡，PWM 亮度）
 *   GPIO 16 → 橙色 LED（暖白色温，PWM）
 *   GPIO 17 → 蓝色 LED（冷白色温，PWM）
 *   GPIO  2 → 内置 LED（MQTT 在线时常亮）
 */

#include <WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>

// ────────────────────────────────────────────────────
// ★ 配置区域 — 仅需修改这里
// ────────────────────────────────────────────────────

// WiFi（Wokwi 内置无密码热点）
#define WIFI_SSID      "Wokwi-GUEST"
#define WIFI_PASSWORD  ""

// LinkFlow 服务器（需公网可达，端口 1883）
// 本地开发可用 ngrok / frp 做 TCP 穿透：ngrok tcp 1883
#define MQTT_HOST      "your.server.ip.or.domain"
#define MQTT_PORT      1883

// 从平台「设备管理」复制以下两项
#define DEVICE_ID      "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"  // 设备 ID（UUID）
#define DEVICE_SECRET  "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"  // 64 字符密钥

// 遥测上报间隔（毫秒）
#define TEL_INTERVAL_MS  10000UL  // 10 秒

// ────────────────────────────────────────────────────

// ---- 引脚 ----
#define PIN_BULB    4   // 主灯泡亮度 (PWM · 黄色 LED)
#define PIN_WARM   16   // 暖白色温   (PWM · 橙色 LED)
#define PIN_COOL   17   // 冷白色温   (PWM · 蓝色 LED)
#define PIN_STATUS  2   // 状态指示   (内置 LED · MQTT 在线时常亮)

// ---- MQTT Topics（运行时构建）----
static char T_TEL_UP[100],    T_TEL_DOWN[100];
static char T_SVC_INVOKE[100], T_SVC_REPLY[100];
static char T_VOICE_DOWN[100], T_EVENT[100];
static char T_OTA_DOWN[100],   T_OTA_UP[100];

// ---- 灯泡状态 ----
struct BulbState {
  bool  sw         = false;  // 开关
  int   brightness = 50;     // 亮度 1–100
  int   colorTemp  = 0;      // 色温: 0=暖白 1=自然白 2=冷白
  float power      = 0.0f;   // 功耗 W（只读，自动计算）
} bulb;

// ---- MQTT ----
WiFiClient   net;
PubSubClient mqtt(net);

// ---- 计时 ----
unsigned long lastTelMs = 0;
unsigned long bootMs    = 0;
bool          bootSent  = false;

// ════════════════════════════════════════════════════
// 工具函数
// ════════════════════════════════════════════════════

void buildTopics() {
  snprintf(T_TEL_UP,     sizeof(T_TEL_UP),     "devices/%s/telemetry/up",   DEVICE_ID);
  snprintf(T_TEL_DOWN,   sizeof(T_TEL_DOWN),   "devices/%s/telemetry/down", DEVICE_ID);
  snprintf(T_SVC_INVOKE, sizeof(T_SVC_INVOKE), "devices/%s/service/invoke", DEVICE_ID);
  snprintf(T_SVC_REPLY,  sizeof(T_SVC_REPLY),  "devices/%s/service/reply",  DEVICE_ID);
  snprintf(T_VOICE_DOWN, sizeof(T_VOICE_DOWN), "devices/%s/voice/down",     DEVICE_ID);
  snprintf(T_EVENT,      sizeof(T_EVENT),      "devices/%s/event",          DEVICE_ID);
  snprintf(T_OTA_DOWN,   sizeof(T_OTA_DOWN),   "devices/%s/ota/down",       DEVICE_ID);
  snprintf(T_OTA_UP,     sizeof(T_OTA_UP),     "devices/%s/ota/up",         DEVICE_ID);
}

// 将当前灯泡状态写入 LED 引脚
void applyLED() {
  if (!bulb.sw) {
    ledcWrite(0, 0); ledcWrite(1, 0); ledcWrite(2, 0);
    Serial.println("[LED] ■ 关灯");
    return;
  }
  // 亮度 1–100 → PWM 5–255
  int pwm  = map(bulb.brightness, 1, 100, 5, 255);
  int warm = 0, cool = 0;
  switch (bulb.colorTemp) {
    case 0: warm = pwm;       cool = pwm / 6;    break;  // 暖白 2700K
    case 1: warm = pwm * 2/3; cool = pwm * 2/3;  break;  // 自然白 4000K
    case 2: warm = pwm / 6;   cool = pwm;         break;  // 冷白 6500K
  }
  ledcWrite(0, pwm); ledcWrite(1, warm); ledcWrite(2, cool);
  Serial.printf("[LED] ● 开灯  brightness=%d%%  color_temp=%d  pwm=%d/warm=%d/cool=%d\n",
                bulb.brightness, bulb.colorTemp, pwm, warm, cool);
}

// 按亮度线性估算功耗（最大 10W）
float calcPower() {
  return bulb.sw ? 10.0f * (bulb.brightness / 100.0f) : 0.0f;
}

// ════════════════════════════════════════════════════
// MQTT 发布
// ════════════════════════════════════════════════════

// ↑ 上报遥测数据
void publishTelemetry() {
  if (!mqtt.connected()) return;

  bulb.power = calcPower();

  StaticJsonDocument<256> doc;
  doc["switch"]     = bulb.sw;
  doc["brightness"] = bulb.brightness;
  doc["color_temp"] = bulb.colorTemp;
  // 保留 1 位小数
  doc["power"]      = (float)((int)(bulb.power * 10)) / 10.0f;

  char buf[256];
  serializeJson(doc, buf);
  bool ok = mqtt.publish(T_TEL_UP, buf, false);
  Serial.printf("[↑ telemetry/up] %s  %s\n", buf, ok ? "✓" : "✗FAIL");
}

// ↑ 上报 boot 开机事件
void publishBootEvent() {
  if (!mqtt.connected()) return;

  StaticJsonDocument<128> doc;
  doc["event_id"]          = "boot";
  doc["params"]["version"] = "1.0.0";

  char buf[128];
  serializeJson(doc, buf);
  bool ok = mqtt.publish(T_EVENT, buf, false);
  Serial.printf("[↑ event/boot] %s  %s\n", buf, ok ? "✓" : "✗FAIL");
}

// ↑ 回传服务执行结果
void publishServiceReply(const char* id, const char* svc,
                         int code, const char* msg,
                         const char* outputJson = nullptr) {
  StaticJsonDocument<256> doc;
  doc["id"]      = id;
  doc["service"] = svc;
  doc["code"]    = code;
  doc["message"] = msg;

  if (outputJson) {
    StaticJsonDocument<128> out;
    if (!deserializeJson(out, outputJson))
      doc["output_params"] = out.as<JsonObject>();
  } else {
    doc.createNestedObject("output_params");
  }

  char buf[256];
  serializeJson(doc, buf);
  bool ok = mqtt.publish(T_SVC_REPLY, buf, false);
  Serial.printf("[↑ service/reply] %s  %s\n", buf, ok ? "✓" : "✗FAIL");
}

// ════════════════════════════════════════════════════
// MQTT 下行处理
// ════════════════════════════════════════════════════

// ↓ telemetry/down — 平台下发属性设置
void onTelemetryDown(const byte* data, unsigned int len) {
  StaticJsonDocument<512> doc;
  if (deserializeJson(doc, data, len)) {
    Serial.println("[↓ telemetry/down] JSON 解析失败");
    return;
  }

  Serial.print("[↓ telemetry/down] ");
  serializeJson(doc, Serial);
  Serial.println();

  bool changed = false;
  if (doc.containsKey("switch")) {
    bulb.sw = doc["switch"].as<bool>();
    changed = true;
  }
  if (doc.containsKey("brightness")) {
    bulb.brightness = constrain(doc["brightness"].as<int>(), 1, 100);
    changed = true;
  }
  if (doc.containsKey("color_temp")) {
    bulb.colorTemp = constrain(doc["color_temp"].as<int>(), 0, 2);
    changed = true;
  }

  if (changed) {
    applyLED();
    publishTelemetry();   // 立即上报确认执行结果
    lastTelMs = millis();
  }
}

// ↓ service/invoke — 服务调用
void onServiceInvoke(const byte* data, unsigned int len) {
  StaticJsonDocument<512> doc;
  if (deserializeJson(doc, data, len)) {
    Serial.println("[↓ service/invoke] JSON 解析失败");
    return;
  }

  const char* id  = doc["id"]      | "?";
  const char* svc = doc["service"] | "";
  Serial.printf("[↓ service/invoke] id=%s  service=%s\n", id, svc);

  // ---- blink：闪烁测试 ----
  if (strcmp(svc, "blink") == 0) {
    int times = constrain((int)(doc["params"]["times"]       | 3), 1, 10);
    int ms    = constrain((int)(doc["params"]["interval_ms"] | 500), 200, 2000);
    Serial.printf("[Blink] %d 次  间隔 %dms\n", times, ms);

    bool orig = bulb.sw;
    for (int i = 0; i < times; i++) {
      bulb.sw = true;  applyLED(); delay(ms / 2);
      bulb.sw = false; applyLED(); delay(ms / 2);
    }
    bulb.sw = orig;
    applyLED();
    publishServiceReply(id, svc, 200, "blink done", "{\"done\":true}");

  // ---- reset：恢复出厂 ----
  } else if (strcmp(svc, "reset") == 0) {
    Serial.println("[Reset] 恢复出厂默认");
    bulb.sw = false; bulb.brightness = 50; bulb.colorTemp = 0;
    applyLED();
    publishTelemetry();
    lastTelMs = millis();
    publishServiceReply(id, svc, 200, "reset done");

  } else {
    Serial.printf("[SvcInvoke] 未知服务: %s\n", svc);
    publishServiceReply(id, svc, 404, "unknown service");
  }
}

// ↓ ota/down — OTA 升级命令（Wokwi 中模拟进度上报）
void onOtaDown(const byte* data, unsigned int len) {
  StaticJsonDocument<512> doc;
  deserializeJson(doc, data, len);

  const char* taskId  = doc["task_id"] | "";
  const char* version = doc["version"] | "";
  const char* url     = doc["url"]     | "";

  Serial.printf("[↓ ota/down] task=%s  ver=%s\n", taskId, version);
  Serial.printf("[OTA] 固件地址: %s\n", url);

  // 模拟 downloading → verifying → installing → completed 各阶段
  struct { const char* status; int progress; } stages[] = {
    {"downloading", 20}, {"downloading", 50}, {"downloading", 80},
    {"verifying",   90}, {"installing",  95}, {"completed",  100}
  };

  for (auto& s : stages) {
    StaticJsonDocument<256> rpt;
    rpt["task_id"]  = taskId;
    rpt["status"]   = s.status;
    rpt["progress"] = s.progress;
    rpt["error"]    = "";
    if (s.progress == 100) rpt["version"] = version;  // 完成时携带新版本号

    char buf[256];
    serializeJson(rpt, buf);
    mqtt.publish(T_OTA_UP, buf, false);
    Serial.printf("[↑ ota/up] %s\n", buf);
    delay(800);
  }
  Serial.println("[OTA] 模拟升级完成 ✓");
}

// ↓ voice/down — 语音执行结果（打印日志）
void onVoiceDown(const byte* data, unsigned int len) {
  StaticJsonDocument<256> doc;
  deserializeJson(doc, data, len);
  Serial.printf("[↓ voice/down] success=%s  msg=%s  action=%s\n",
    doc["success"].as<bool>() ? "true" : "false",
    doc["message"] | "",
    doc["action"]  | "");
}

// ════════════════════════════════════════════════════
// MQTT 回调路由
// ════════════════════════════════════════════════════

void mqttCallback(char* topic, byte* payload, unsigned int len) {
  if      (!strcmp(topic, T_TEL_DOWN))   onTelemetryDown(payload, len);
  else if (!strcmp(topic, T_SVC_INVOKE)) onServiceInvoke(payload, len);
  else if (!strcmp(topic, T_OTA_DOWN))   onOtaDown(payload, len);
  else if (!strcmp(topic, T_VOICE_DOWN)) onVoiceDown(payload, len);
}

// ════════════════════════════════════════════════════
// 连接管理
// ════════════════════════════════════════════════════

void connectWiFi() {
  Serial.printf("[WiFi] 连接 %s ...", WIFI_SSID);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  while (WiFi.status() != WL_CONNECTED) { delay(500); Serial.print("."); }
  Serial.printf("\n[WiFi] 已连接  IP=%s\n", WiFi.localIP().toString().c_str());
}

void connectMQTT() {
  while (!mqtt.connected()) {
    Serial.printf("[MQTT] 连接 %s:%d ...\n", MQTT_HOST, MQTT_PORT);
    digitalWrite(PIN_STATUS, LOW);

    // clientID = DEVICE_ID（Broker 用此作为设备 ID 认证）
    // username  = DEVICE_ID（确保 password 字段被 MQTT 协议携带）
    // password  = DEVICE_SECRET（64 字符密钥）
    if (mqtt.connect(DEVICE_ID, DEVICE_ID, DEVICE_SECRET)) {
      Serial.println("[MQTT] 已连接 ✓");
      digitalWrite(PIN_STATUS, HIGH);

      mqtt.subscribe(T_TEL_DOWN);
      mqtt.subscribe(T_SVC_INVOKE);
      mqtt.subscribe(T_VOICE_DOWN);
      mqtt.subscribe(T_OTA_DOWN);
      Serial.println("[MQTT] 已订阅 4 个下行 Topic");

    } else {
      Serial.printf("[MQTT] 连接失败 state=%d，5s 后重试\n", mqtt.state());
      delay(5000);
    }
  }
}

// ════════════════════════════════════════════════════
// setup / loop
// ════════════════════════════════════════════════════

void setup() {
  Serial.begin(115200);
  delay(200);
  Serial.println(F("\n====================================="));
  Serial.println(F("  LinkFlow Smart Bulb Simulator 1.0 "));
  Serial.println(F("====================================="));

  pinMode(PIN_STATUS, OUTPUT);
  digitalWrite(PIN_STATUS, LOW);

  // 初始化 3 路 PWM（LEDC）
  ledcSetup(0, 5000, 8); ledcAttachPin(PIN_BULB, 0);
  ledcSetup(1, 5000, 8); ledcAttachPin(PIN_WARM, 1);
  ledcSetup(2, 5000, 8); ledcAttachPin(PIN_COOL, 2);

  buildTopics();

  Serial.printf("[Config] Broker = %s:%d\n", MQTT_HOST, MQTT_PORT);
  Serial.printf("[Config] Device = %s\n",   DEVICE_ID);
  Serial.printf("[Config] TelUp  = %s\n",   T_TEL_UP);

  bootMs = millis();

  connectWiFi();

  mqtt.setServer(MQTT_HOST, MQTT_PORT);
  mqtt.setCallback(mqttCallback);
  mqtt.setBufferSize(1024);  // 扩大缓冲区以支持 OTA payload

  connectMQTT();
}

void loop() {
  if (!mqtt.connected()) connectMQTT();
  mqtt.loop();

  // 首次连接后：发送 boot 事件 + 上报初始状态
  if (!bootSent && mqtt.connected()) {
    publishBootEvent();
    publishTelemetry();
    lastTelMs = millis();
    bootSent  = true;
  }

  // 定时遥测上报
  if (millis() - lastTelMs >= TEL_INTERVAL_MS) {
    publishTelemetry();
    lastTelMs = millis();
  }
}
