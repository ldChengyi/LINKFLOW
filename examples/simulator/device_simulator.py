#!/usr/bin/env python3
"""
LinkFlow IoT 设备模拟器
=======================
从设备视角完整模拟一个 IoT 终端，无需真实硬件即可验证平台全部功能。

模拟功能：
  ✅ WiFi 连接（TCP/MQTT 自动重连，模拟掉线恢复）
  ✅ 传感器随机游走（温度 / 湿度 实时变化）
  ✅ 遥测数据定时上报
  ✅ 属性下发处理（switch / brightness / mode）
  ✅ 服务调用处理 + 结果回传（reboot / reset）
  ✅ 语音指令上报 + 执行结果接收
  ✅ 事件主动上报
  ✅ 实时状态面板（Rich TUI）+ 日志滚动

依赖安装：
  pip install -r requirements.txt

运行示例：
  python device_simulator.py
  python device_simulator.py --host 192.168.1.100 --device-id xxx --secret yyy --name 客厅灯
  python device_simulator.py --interval 3 --no-tui
"""

import argparse
import json
import random
import sys
import threading
import time
from collections import deque
from datetime import datetime

# ── 依赖检查 ──────────────────────────────────────────────────────────────────

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("[ERROR] 请先安装依赖: pip install -r requirements.txt")
    sys.exit(1)

try:
    from rich.console import Console
    from rich.layout import Layout
    from rich.live import Live
    from rich.panel import Panel
    from rich.table import Table
    from rich.text import Text
    from rich import box
    RICH_OK = True
except ImportError:
    RICH_OK = False

# ── 默认配置（可通过命令行覆盖）─────────────────────────────────────────────

DEFAULT_HOST     = "localhost"
DEFAULT_PORT     = 1883
DEFAULT_ID       = "YOUR_DEVICE_ID"
DEFAULT_SECRET   = "YOUR_DEVICE_SECRET"
DEFAULT_NAME     = "模拟设备"
DEFAULT_INTERVAL = 5.0   # 遥测上报间隔（秒）

# ── 设备属性模型（与平台物模型保持一致）─────────────────────────────────────
#
#  标识符       名称    类型   读写  范围          单位
#  switch       开关    bool   rw
#  brightness   亮度    int    rw   0 ~ 100       %
#  mode         模式    enum   rw   0=正常/1=节能/2=强光
#  temperature  温度    float  r                  ℃
#  humidity     湿度    float  r    0 ~ 100       %
#
# 服务：reboot（重启）/ reset（重置属性）
# 事件：temperature_alarm（温度告警）/ device_boot（设备启动）


# ── 全局状态 ──────────────────────────────────────────────────────────────────

class DeviceState:
    def __init__(self):
        self._lock = threading.Lock()
        # 属性
        self.switch      = False
        self.brightness  = 50
        self.mode        = 0
        self.temperature = 24.0
        self.humidity    = 60.0
        # 连接状态
        self.connected   = False
        self.reconnects  = 0
        self.tx_count    = 0   # 上报次数
        self.rx_count    = 0   # 接收次数
        self.start_time  = time.time()
        # 日志队列（最近 200 条）
        self.logs: deque = deque(maxlen=200)

    def log(self, msg: str, level: str = "INFO"):
        ts = datetime.now().strftime("%H:%M:%S")
        entry = (level, f"[{ts}] {msg}")
        with self._lock:
            self.logs.append(entry)

    def drift(self):
        """传感器随机游走"""
        with self._lock:
            self.temperature += random.gauss(0, 0.2)
            self.temperature  = round(max(10.0, min(45.0, self.temperature)), 1)
            self.humidity    += random.gauss(0, 0.3)
            self.humidity     = round(max(10.0, min(98.0, self.humidity)), 1)

    def to_payload(self) -> dict:
        with self._lock:
            return {
                "switch":      self.switch,
                "brightness":  self.brightness,
                "mode":        self.mode,
                "temperature": self.temperature,
                "humidity":    self.humidity,
            }

    @property
    def uptime(self) -> str:
        s = int(time.time() - self.start_time)
        return f"{s // 3600:02d}:{s % 3600 // 60:02d}:{s % 60:02d}"


state = DeviceState()


# ── MQTT 设备客户端 ───────────────────────────────────────────────────────────

class LinkFlowDevice:
    """模拟 IoT 终端，通过 MQTT 接入 LinkFlow 平台"""

    def __init__(self, host: str, port: int, device_id: str, secret: str, name: str):
        self.host      = host
        self.port      = port
        self.device_id = device_id
        self.secret    = secret
        self.name      = name

        # MQTT Topics
        pfx = f"devices/{device_id}"
        self.T_UP     = f"{pfx}/telemetry/up"
        self.T_DOWN   = f"{pfx}/telemetry/down"
        self.T_INVOKE = f"{pfx}/service/invoke"
        self.T_REPLY  = f"{pfx}/service/reply"
        self.T_EVENT  = f"{pfx}/event"
        self.T_V_UP   = f"{pfx}/voice/up"
        self.T_V_DOWN = f"{pfx}/voice/down"

        self._client = mqtt.Client(client_id=device_id, protocol=mqtt.MQTTv311)
        self._client.username_pw_set(device_id, secret)
        self._client.on_connect    = self._on_connect
        self._client.on_disconnect = self._on_disconnect
        self._client.on_message    = self._on_message
        self._client.reconnect_delay_set(min_delay=2, max_delay=30)

    # ── 连接管理 ──────────────────────────────────────────────────────────────

    def start(self):
        state.log(f"正在连接 {self.host}:{self.port} ...")
        try:
            self._client.connect_async(self.host, self.port, keepalive=60)
            self._client.loop_start()
        except Exception as e:
            state.log(f"连接失败: {e}", "ERROR")

    def stop(self):
        self._client.loop_stop()
        self._client.disconnect()

    def simulate_disconnect(self, seconds: float = 3.0):
        """模拟掉线后重连（如重启场景）"""
        def _do():
            state.log(f"模拟断线 {seconds}s ...", "WARN")
            self._client.disconnect()
            time.sleep(seconds)
            state.log("正在重连 ...")
            self._client.reconnect()
        threading.Thread(target=_do, daemon=True).start()

    # ── MQTT 回调 ─────────────────────────────────────────────────────────────

    def _on_connect(self, client, userdata, flags, rc):
        RC_MSG = {
            0: "✅ 连接成功",
            1: "协议版本不支持",
            2: "ClientID 被拒绝",
            3: "服务不可用",
            4: "用户名/密码错误（检查 DEVICE_SECRET）",
            5: "未授权",
        }
        if rc == 0:
            state.connected = True
            state.log(f"WiFi/MQTT 已接入: {RC_MSG[rc]}")
            # 订阅下行 Topic
            subs = [(self.T_DOWN, 1), (self.T_INVOKE, 1), (self.T_V_DOWN, 1)]
            client.subscribe(subs)
            state.log(f"已订阅: telemetry/down | service/invoke | voice/down")
            # 上线立即上报状态
            self._publish_telemetry()
            # 上报启动事件
            self._publish_event("device_boot", {"firmware": "simulator-v1.0"})
        else:
            state.log(f"❌ 连接被拒绝: {RC_MSG.get(rc, rc)}", "ERROR")

    def _on_disconnect(self, client, userdata, rc):
        state.connected = False
        if rc != 0:
            state.reconnects += 1
            state.log(f"⚠️  连接断开 (rc={rc})，等待自动重连... (第{state.reconnects}次)", "WARN")

    def _on_message(self, client, userdata, msg):
        try:
            data = json.loads(msg.payload.decode("utf-8"))
        except Exception:
            state.log(f"消息解析失败: {msg.payload!r}", "ERROR")
            return

        state.rx_count += 1

        if msg.topic == self.T_DOWN:
            self._handle_property_set(data)
        elif msg.topic == self.T_INVOKE:
            self._handle_service_invoke(data)
        elif msg.topic == self.T_V_DOWN:
            self._handle_voice_result(data)

    # ── 属性下发处理 ──────────────────────────────────────────────────────────

    def _handle_property_set(self, data: dict):
        state.log(f"📥 属性下发: {json.dumps(data, ensure_ascii=False)}")
        changed = False

        with state._lock:
            if "switch" in data:
                state.switch = bool(data["switch"])
                state.log(f"   └─ switch → {'🟢 ON' if state.switch else '⚫ OFF'}")
                changed = True
            if "brightness" in data:
                state.brightness = max(0, min(100, int(data["brightness"])))
                bar = "█" * (state.brightness // 10) + "░" * (10 - state.brightness // 10)
                state.log(f"   └─ brightness → {state.brightness}% [{bar}]")
                changed = True
            if "mode" in data:
                state.mode = int(data["mode"])
                labels = {0: "正常", 1: "节能 🌿", 2: "强光 ☀️"}
                state.log(f"   └─ mode → {labels.get(state.mode, state.mode)}")
                changed = True

        if changed:
            # 属性变更后立即回传最新状态（闭合数据链路）
            self._publish_telemetry()

    # ── 服务调用处理 ──────────────────────────────────────────────────────────

    def _handle_service_invoke(self, data: dict):
        req_id  = data.get("id", "unknown")
        service = data.get("service", "")
        params  = data.get("params", {})
        state.log(f"🔧 服务调用: service={service} id={req_id[:16]} params={params}")

        if service == "reboot":
            self._reply_service(req_id, service, True, "设备重启中，3 秒后恢复")
            state.log("♻️  模拟重启：断线 → 等待 3s → 重连")
            self.simulate_disconnect(seconds=3.0)

        elif service == "reset":
            with state._lock:
                state.switch     = False
                state.brightness = 50
                state.mode       = 0
            state.log("🔄 属性已全部重置为默认值")
            self._reply_service(req_id, service, True, "属性已重置")
            self._publish_telemetry()

        else:
            state.log(f"⚠️  未知服务: {service}", "WARN")
            self._reply_service(req_id, service, False, f"未知服务: {service}")

    # ── 语音结果处理 ──────────────────────────────────────────────────────────

    def _handle_voice_result(self, data: dict):
        success = data.get("success", False)
        message = data.get("message", "")
        action  = data.get("action", "")
        if success:
            state.log(f"🎙️  语音成功 → {action}")
        else:
            state.log(f"🎙️  语音失败 → {message}", "WARN")

    # ── 主动发布 ──────────────────────────────────────────────────────────────

    def _publish_telemetry(self):
        if not state.connected:
            return
        state.drift()
        payload = json.dumps(state.to_payload(), ensure_ascii=False)
        self._client.publish(self.T_UP, payload, qos=1)
        state.tx_count += 1
        state.log(f"📤 遥测 #{state.tx_count}: {payload}")

    def publish_telemetry(self):
        self._publish_telemetry()

    def publish_voice(self, text: str):
        if not state.connected:
            state.log("未连接，无法发送语音指令", "WARN")
            return
        payload = json.dumps({"text": text}, ensure_ascii=False)
        self._client.publish(self.T_V_UP, payload, qos=1)
        state.log(f"🎤 发送语音: {text}")

    def _publish_event(self, event_id: str, params: dict):
        if not state.connected:
            return
        payload = json.dumps({
            "id":     event_id,
            "time":   datetime.utcnow().isoformat() + "Z",
            "params": params,
        }, ensure_ascii=False)
        self._client.publish(self.T_EVENT, payload, qos=1)
        state.log(f"⚡ 事件上报: {event_id} {params}")

    def publish_event(self, event_id: str, params: dict):
        self._publish_event(event_id, params)


# ── 后台定时遥测线程 ─────────────────────────────────────────────────────────

def telemetry_loop(device: LinkFlowDevice, interval: float, stop: threading.Event):
    while not stop.wait(interval):
        if state.connected:
            device.publish_telemetry()
        # 温度超过 38℃ 自动上报告警事件
        if state.temperature > 38.0:
            device.publish_event("temperature_alarm", {
                "temperature": state.temperature,
                "threshold":   38.0,
            })


# ── Rich TUI 面板 ────────────────────────────────────────────────────────────

def make_layout(device: LinkFlowDevice) -> Layout:
    layout = Layout()
    layout.split_column(
        Layout(name="header", size=3),
        Layout(name="body"),
        Layout(name="footer", size=3),
    )
    layout["body"].split_row(
        Layout(name="state", ratio=1),
        Layout(name="logs",  ratio=2),
    )
    return layout


def render_header(device: LinkFlowDevice) -> Panel:
    conn_str = "[bold green]● 已连接[/]" if state.connected else "[bold red]○ 未连接[/]"
    t = Text()
    t.append("LinkFlow 设备模拟器  ", style="bold cyan")
    t.append(f"设备: {device.name}  ")
    t.append(conn_str)
    t.append(f"  运行: {state.uptime}  ")
    t.append(f"↑{state.tx_count}  ↓{state.rx_count}  重连:{state.reconnects}")
    return Panel(t, box=box.HORIZONTALS)


def render_state(device: LinkFlowDevice) -> Panel:
    tbl = Table(box=box.SIMPLE, show_header=False, padding=(0, 1))
    tbl.add_column("key",  style="dim", width=12)
    tbl.add_column("val",  style="bold")

    sw = "[green]ON  🟢[/]" if state.switch else "[dim]OFF ⚫[/]"
    tbl.add_row("switch",     sw)

    bar = "█" * (state.brightness // 10) + "░" * (10 - state.brightness // 10)
    tbl.add_row("brightness", f"{state.brightness:3d}% [{bar}]")

    mode_lbl = {0: "正常", 1: "节能 🌿", 2: "强光 ☀️"}.get(state.mode, str(state.mode))
    tbl.add_row("mode",       f"{state.mode} ({mode_lbl})")

    temp_color = "red" if state.temperature > 38 else "yellow" if state.temperature > 30 else "cyan"
    tbl.add_row("temperature", f"[{temp_color}]{state.temperature:.1f} ℃[/]")

    humi_color = "blue" if state.humidity > 80 else "cyan"
    tbl.add_row("humidity",   f"[{humi_color}]{state.humidity:.1f} %[/]")

    tbl.add_row("", "")
    tbl.add_row("broker",     f"{device.host}:{device.port}")
    tbl.add_row("device_id",  f"{device.device_id[:16]}...")

    return Panel(tbl, title="设备状态", border_style="green" if state.connected else "red")


def render_logs() -> Panel:
    lines = list(state.logs)[-20:]  # 最近 20 条
    text = Text()
    level_styles = {
        "INFO": "white",
        "WARN": "yellow",
        "ERROR": "bold red",
    }
    for level, msg in lines:
        text.append(msg + "\n", style=level_styles.get(level, "white"))
    return Panel(text, title="日志", border_style="blue")


def render_footer() -> Panel:
    help_text = (
        "[cyan]1[/]=切换开关  [cyan]2[/]=设置亮度  [cyan]3[/]=切换模式  "
        "[cyan]4[/]=设置温度  [cyan]5[/]=发送语音  [cyan]6[/]=上报事件  "
        "[cyan]7[/]=立即遥测  [cyan]8[/]=模拟断线  [cyan]q[/]=退出"
    )
    return Panel(Text.from_markup(help_text), box=box.HORIZONTALS)


# ── 交互菜单（无 Rich 模式） ──────────────────────────────────────────────────

PLAIN_MENU = """
╔══════════════════════════════════════════╗
║      LinkFlow 设备模拟器 — 控制台        ║
╠══════════════════════════════════════════╣
║  1  切换开关 (switch on/off)             ║
║  2  设置亮度 (0-100)                     ║
║  3  切换模式 (0=正常 1=节能 2=强光)     ║
║  4  设置温度（覆盖传感器值）             ║
║  5  发送语音指令                         ║
║  6  上报事件                             ║
║  7  立即上报一次遥测                     ║
║  8  模拟断线重连（3秒）                 ║
║  s  查看当前状态                         ║
║  q  退出                                 ║
╚══════════════════════════════════════════╝
"""

# 语音指令预设
VOICE_PRESETS = [
    ("打开{name}",         "开灯"),
    ("关闭{name}",         "关灯"),
    ("{name}亮度调到80",   "调亮度到80"),
    ("{name}亮度调高",     "亮度调高"),
    ("{name}亮度调低",     "亮度调低"),
    ("设为节能模式",       "节能模式"),
    ("设为正常模式",       "正常模式"),
    ("设为强光模式",       "强光模式"),
    ("重启{name}",         "重启服务"),
    ("{name}温度是多少",   "查询温度"),
    ("{name}亮度当前是多少", "查询亮度"),
    ("自定义输入...",      ""),
]


def voice_menu(device: LinkFlowDevice):
    print("\n── 语音指令预设 ────────────────────────────────")
    for i, (tpl, desc) in enumerate(VOICE_PRESETS, 1):
        text = tpl.format(name=device.name)
        print(f"  {i:2d}. {text:<30} ({desc})")
    print("────────────────────────────────────────────────")
    raw = input("输入编号 或 直接输入语音文本: ").strip()

    if raw.isdigit():
        idx = int(raw) - 1
        if 0 <= idx < len(VOICE_PRESETS) - 1:
            text = VOICE_PRESETS[idx][0].format(name=device.name)
            device.publish_voice(text)
            return
        elif idx == len(VOICE_PRESETS) - 1:
            raw = input("请输入语音文本: ").strip()

    if raw:
        device.publish_voice(raw)


def show_state_plain(device: LinkFlowDevice):
    print(f"\n── 当前设备状态 ({'已连接' if state.connected else '未连接'}) ──")
    print(f"  switch       : {'ON' if state.switch else 'OFF'}")
    print(f"  brightness   : {state.brightness}%")
    print(f"  mode         : {state.mode} ({['正常','节能','强光'][state.mode]})")
    print(f"  temperature  : {state.temperature:.1f}℃")
    print(f"  humidity     : {state.humidity:.1f}%")
    print(f"  运行时长     : {state.uptime}")
    print(f"  遥测计数     : 上报 {state.tx_count} | 接收 {state.rx_count}")
    print(f"  重连次数     : {state.reconnects}")
    print()


def run_input_loop_plain(device: LinkFlowDevice, stop: threading.Event):
    print(PLAIN_MENU)
    while not stop.is_set():
        try:
            cmd = input("指令> ").strip().lower()
        except (EOFError, KeyboardInterrupt):
            break

        if cmd in ("", "h", "help"):
            print(PLAIN_MENU)

        elif cmd == "1":
            with state._lock:
                state.switch = not state.switch
                s = state.switch
            state.log(f"手动切换 switch → {'ON' if s else 'OFF'}")
            device.publish_telemetry()

        elif cmd == "2":
            try:
                val = int(input("亮度 (0-100): "))
                with state._lock:
                    state.brightness = max(0, min(100, val))
                state.log(f"手动设置 brightness → {state.brightness}%")
                device.publish_telemetry()
            except ValueError:
                print("请输入整数")

        elif cmd == "3":
            print("  0=正常  1=节能  2=强光")
            try:
                val = int(input("模式: "))
                if val in (0, 1, 2):
                    with state._lock:
                        state.mode = val
                    state.log(f"手动切换 mode → {val}")
                    device.publish_telemetry()
                else:
                    print("无效值（0/1/2）")
            except ValueError:
                print("请输入整数")

        elif cmd == "4":
            try:
                val = float(input("温度 (℃): "))
                with state._lock:
                    state.temperature = round(val, 1)
                state.log(f"手动设置 temperature → {state.temperature}℃")
                device.publish_telemetry()
            except ValueError:
                print("请输入数字")

        elif cmd == "5":
            voice_menu(device)

        elif cmd == "6":
            print("  1=温度告警  2=设备启动  3=自定义")
            c = input("选择: ").strip()
            if c == "1":
                device.publish_event("temperature_alarm",
                    {"temperature": state.temperature, "threshold": 38.0})
            elif c == "2":
                device.publish_event("device_boot", {"firmware": "simulator-v1.0"})
            elif c == "3":
                eid = input("事件ID: ").strip()
                if eid:
                    device.publish_event(eid, {})

        elif cmd == "7":
            device.publish_telemetry()

        elif cmd == "8":
            device.simulate_disconnect(seconds=3.0)

        elif cmd == "s":
            show_state_plain(device)

        elif cmd in ("q", "quit", "exit"):
            break

        else:
            print(f"未知指令 '{cmd}'")

    stop.set()


# ── Rich TUI 主循环 ───────────────────────────────────────────────────────────

def run_tui(device: LinkFlowDevice, stop: threading.Event):
    """在后台线程中刷新 Rich 面板，主线程处理键盘输入"""
    console = Console()
    layout  = make_layout(device)

    def _refresh():
        while not stop.is_set():
            layout["header"].update(render_header(device))
            layout["body"]["state"].update(render_state(device))
            layout["body"]["logs"].update(render_logs())
            layout["footer"].update(render_footer())
            time.sleep(0.5)

    refresh_thread = threading.Thread(target=_refresh, daemon=True)

    with Live(layout, console=console, screen=True, refresh_per_second=2):
        refresh_thread.start()
        # 用 getch 风格输入（需要 readline）
        run_input_loop_plain(device, stop)


# ── 主入口 ────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="LinkFlow IoT 设备模拟器",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--host",      default=DEFAULT_HOST,     help="MQTT Broker 地址")
    parser.add_argument("--port",      default=DEFAULT_PORT,     type=int, help="MQTT 端口")
    parser.add_argument("--device-id", default=DEFAULT_ID,       help="设备 ID（UUID）")
    parser.add_argument("--secret",    default=DEFAULT_SECRET,   help="设备密钥（64字符）")
    parser.add_argument("--name",      default=DEFAULT_NAME,     help="设备显示名称（用于语音预设）")
    parser.add_argument("--interval",  default=DEFAULT_INTERVAL, type=float, help="遥测上报间隔（秒）")
    parser.add_argument("--no-tui",    action="store_true",      help="禁用 Rich TUI，使用纯文本模式")
    args = parser.parse_args()

    # 检查是否填入了真实凭证
    if args.device_id == DEFAULT_ID or args.secret == DEFAULT_SECRET:
        print("=" * 60)
        print("⚠️  请填写真实的设备凭证！")
        print()
        print("方式一：命令行参数")
        print("  python device_simulator.py \\")
        print("    --host 192.168.1.100 \\")
        print("    --device-id xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \\")
        print("    --secret xxxx...xxxx \\")
        print("    --name 客厅灯")
        print()
        print("方式二：直接修改本文件顶部的 DEFAULT_* 常量")
        print("=" * 60)
        sys.exit(1)

    use_tui = RICH_OK and not args.no_tui

    print(f"""
╔══════════════════════════════════════════════╗
║       LinkFlow IoT 设备模拟器 v1.0           ║
╠══════════════════════════════════════════════╣
║  Broker    : {args.host}:{args.port:<28}  ║
║  设备名称  : {args.name:<32}  ║
║  设备 ID   : {args.device_id[:32]:<32}  ║
║  遥测间隔  : {args.interval}s{'':<36}  ║
║  界面模式  : {'Rich TUI' if use_tui else '纯文本':<32}  ║
╚══════════════════════════════════════════════╝
""")

    device   = LinkFlowDevice(args.host, args.port, args.device_id, args.secret, args.name)
    stop_evt = threading.Event()

    # 启动 MQTT（模拟 WiFi 接入）
    device.start()
    time.sleep(1.5)  # 等待连接

    # 启动后台定时遥测线程
    t_tele = threading.Thread(
        target=telemetry_loop,
        args=(device, args.interval, stop_evt),
        daemon=True,
    )
    t_tele.start()

    try:
        if use_tui:
            run_tui(device, stop_evt)
        else:
            run_input_loop_plain(device, stop_evt)
    except KeyboardInterrupt:
        pass
    finally:
        print("\n正在断开连接...")
        stop_evt.set()
        device.stop()
        print("模拟器已退出。")


if __name__ == "__main__":
    main()
