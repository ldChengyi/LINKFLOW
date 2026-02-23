import { useEffect, useRef, useState } from 'react';
import {
  Copy, Check,
  Wifi, Lock, Activity, Shield,
} from 'lucide-react';
import SiteNavBar from '../components/SiteNavBar';

// ── Sidebar config ────────────────────────────────────────────────────────────

const SECTIONS = [
  {
    group: '快速开始',
    items: [
      { id: 'intro', label: '三步接入设备' },
      { id: 'mqtt-connect', label: 'MQTT 连接参数' },
      { id: 'telemetry-up', label: '遥测数据上报' },
    ],
  },
  {
    group: 'MQTT 协议',
    items: [
      { id: 'topic-list', label: 'Topic 结构总览' },
      { id: 'event-up', label: '事件上报' },
      { id: 'ota-progress', label: 'OTA 进度上报' },
      { id: 'acl', label: 'ACL 访问控制' },
    ],
  },
  {
    group: 'REST API',
    items: [
      { id: 'api-auth', label: '认证接口' },
      { id: 'api-device', label: '设备管理' },
      { id: 'api-alert', label: '告警与任务' },
      { id: 'api-ota', label: 'OTA 固件升级' },
      { id: 'api-example', label: '请求示例' },
    ],
  },
];

// ── Small reusable components ─────────────────────────────────────────────────

function CodeBlock({ lang, code }: { lang: string; code: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };
  return (
    <div style={{ borderRadius: 10, border: '1px solid var(--color-border)', overflow: 'hidden', background: 'var(--color-card)' }}>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '8px 16px', borderBottom: '1px solid var(--color-border)',
        background: 'var(--color-secondary)',
      }}>
        <span style={{ fontSize: 12, color: 'var(--color-muted-foreground)', fontFamily: 'monospace' }}>{lang}</span>
        <button onClick={copy} style={{
          background: 'none', border: 'none', cursor: 'pointer',
          color: copied ? 'var(--color-primary)' : 'var(--color-muted-foreground)',
          fontSize: 12, display: 'flex', alignItems: 'center', gap: 4,
        }}>
          {copied ? <><Check size={12} /> 已复制</> : <><Copy size={12} /> 复制</>}
        </button>
      </div>
      <pre style={{
        margin: 0, padding: '16px 20px', fontSize: 13, lineHeight: 1.7,
        color: 'var(--color-foreground)',
        fontFamily: "'Cascadia Code','Fira Code','JetBrains Mono',Consolas,monospace",
        overflowX: 'auto', whiteSpace: 'pre',
      }}>
        <code>{code}</code>
      </pre>
    </div>
  );
}

function ApiTable({ rows }: { rows: { method: string; path: string; desc: string; auth?: boolean }[] }) {
  const methodColor: Record<string, string> = {
    GET: '#22c55e', POST: '#3b82f6', PUT: '#f59e0b', DELETE: '#ef4444',
  };
  return (
    <div style={{ border: '1px solid var(--color-border)', borderRadius: 10, overflow: 'hidden' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr style={{ background: 'var(--color-secondary)' }}>
            {['方法', '路径', '说明', '认证'].map(h => (
              <th key={h} style={{
                padding: '10px 16px', textAlign: 'left',
                color: 'var(--color-muted-foreground)', fontWeight: 600, fontSize: 12,
                borderBottom: '1px solid var(--color-border)',
              }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} style={{
              borderBottom: i < rows.length - 1 ? '1px solid var(--color-border)' : 'none',
              background: i % 2 === 0 ? 'var(--color-card)' : 'transparent',
            }}>
              <td style={{ padding: '10px 16px' }}>
                <span style={{
                  display: 'inline-block', padding: '2px 8px', borderRadius: 4,
                  fontSize: 11, fontWeight: 700, fontFamily: 'monospace',
                  color: methodColor[row.method] || 'var(--color-foreground)',
                  background: (methodColor[row.method] || '#888') + '1a',
                }}>{row.method}</span>
              </td>
              <td style={{ padding: '10px 16px' }}>
                <code style={{ fontSize: 12, color: 'var(--color-foreground)', fontFamily: 'monospace' }}>{row.path}</code>
              </td>
              <td style={{ padding: '10px 16px', color: 'var(--color-muted-foreground)' }}>{row.desc}</td>
              <td style={{ padding: '10px 16px' }}>
                {row.auth !== false
                  ? <span style={{ fontSize: 11, color: 'var(--color-primary)' }}>● JWT</span>
                  : <span style={{ fontSize: 11, color: 'var(--color-muted-foreground)' }}>公开</span>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TopicTable({ rows }: { rows: { topic: string; dir: string; desc: string }[] }) {
  return (
    <div style={{ border: '1px solid var(--color-border)', borderRadius: 10, overflow: 'hidden' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr style={{ background: 'var(--color-secondary)' }}>
            {['Topic', '方向', '说明'].map(h => (
              <th key={h} style={{
                padding: '10px 16px', textAlign: 'left',
                color: 'var(--color-muted-foreground)', fontWeight: 600, fontSize: 12,
                borderBottom: '1px solid var(--color-border)',
              }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} style={{
              borderBottom: i < rows.length - 1 ? '1px solid var(--color-border)' : 'none',
              background: i % 2 === 0 ? 'var(--color-card)' : 'transparent',
            }}>
              <td style={{ padding: '10px 16px' }}>
                <code style={{ fontSize: 12, color: 'var(--color-primary)', fontFamily: 'monospace' }}>{row.topic}</code>
              </td>
              <td style={{ padding: '10px 16px' }}>
                <span style={{
                  display: 'inline-block', padding: '2px 8px', borderRadius: 4,
                  fontSize: 11, fontWeight: 600,
                  ...(row.dir === '设备→云端'
                    ? { color: '#22c55e', background: '#22c55e1a' }
                    : { color: '#3b82f6', background: '#3b82f61a' }),
                }}>{row.dir}</span>
              </td>
              <td style={{ padding: '10px 16px', color: 'var(--color-muted-foreground)' }}>{row.desc}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// Section anchor heading
function H2({ id, children }: { id: string; children: React.ReactNode }) {
  return (
    <h2 id={id} style={{
      fontSize: 22, fontWeight: 700, margin: '0 0 8px',
      color: 'var(--color-foreground)', letterSpacing: '-0.3px',
      scrollMarginTop: 80,
    }}>
      {children}
    </h2>
  );
}

function H3({ children }: { children: React.ReactNode }) {
  return (
    <h3 style={{
      fontSize: 15, fontWeight: 600, margin: '24px 0 10px',
      color: 'var(--color-foreground)',
    }}>
      {children}
    </h3>
  );
}

function P({ children }: { children: React.ReactNode }) {
  return (
    <p style={{ margin: '0 0 16px', fontSize: 14, color: 'var(--color-muted-foreground)', lineHeight: 1.75 }}>
      {children}
    </p>
  );
}

function Divider() {
  return <div style={{ borderTop: '1px solid var(--color-border)', margin: '48px 0' }} />;
}

// ── Main ──────────────────────────────────────────────────────────────────────

export default function DocsPage() {
  const [activeId, setActiveId] = useState('intro');
  const contentRef = useRef<HTMLDivElement>(null);

  // Track active section via IntersectionObserver
  useEffect(() => {
    const allIds = SECTIONS.flatMap(g => g.items.map(i => i.id));
    const observers: IntersectionObserver[] = [];

    allIds.forEach(id => {
      const el = document.getElementById(id);
      if (!el) return;
      const obs = new IntersectionObserver(
        ([entry]) => { if (entry.isIntersecting) setActiveId(id); },
        { rootMargin: '-20% 0px -70% 0px', threshold: 0 }
      );
      obs.observe(el);
      observers.push(obs);
    });

    return () => observers.forEach(o => o.disconnect());
  }, []);

  const scrollTo = (id: string) => {
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--color-background)', color: 'var(--color-foreground)' }}>

      <SiteNavBar page="docs" />

      {/* ── Body: sidebar + content ── */}
      <div style={{ maxWidth: 1400, margin: '0 auto', display: 'flex', minHeight: 'calc(100vh - 60px)' }}>

        {/* Sidebar */}
        <aside style={{
          width: 220, flexShrink: 0,
          position: 'sticky', top: 60, alignSelf: 'flex-start',
          height: 'calc(100vh - 60px)', overflowY: 'auto',
          padding: '28px 0 28px 24px',
          borderRight: '1px solid var(--color-border)',
        }}>
          {SECTIONS.map(group => (
            <div key={group.group} style={{ marginBottom: 24 }}>
              <p style={{
                margin: '0 0 6px', fontSize: 11, fontWeight: 700,
                color: 'var(--color-muted-foreground)',
                textTransform: 'uppercase', letterSpacing: '0.8px',
              }}>
                {group.group}
              </p>
              {group.items.map(item => {
                const isActive = activeId === item.id;
                return (
                  <button
                    key={item.id}
                    onClick={() => scrollTo(item.id)}
                    style={{
                      display: 'block', width: '100%', textAlign: 'left',
                      padding: '5px 10px', marginBottom: 1,
                      borderRadius: 6, border: 'none', cursor: 'pointer',
                      fontSize: 13,
                      background: isActive ? 'var(--color-primary)18' : 'transparent',
                      color: isActive ? 'var(--color-primary)' : 'var(--color-muted-foreground)',
                      fontWeight: isActive ? 600 : 400,
                      borderLeft: isActive ? '2px solid var(--color-primary)' : '2px solid transparent',
                      transition: 'all 0.15s',
                    }}
                    onMouseEnter={e => {
                      if (!isActive) (e.currentTarget as HTMLElement).style.color = 'var(--color-foreground)';
                    }}
                    onMouseLeave={e => {
                      if (!isActive) (e.currentTarget as HTMLElement).style.color = 'var(--color-muted-foreground)';
                    }}
                  >
                    {item.label}
                  </button>
                );
              })}
            </div>
          ))}
        </aside>

        {/* Content */}
        <main ref={contentRef} style={{ flex: 1, padding: '48px 64px 96px 64px', minWidth: 0 }}>
          <div style={{ maxWidth: 760 }}>

            {/* ════════════════════ 快速开始 ════════════════════ */}

            <H2 id="intro">三步接入设备</H2>
            <P>从注册账号到设备数据上报，整个流程不超过 5 分钟。</P>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16, marginBottom: 32 }}>
              {[
                { step: '01', icon: Lock, title: '创建设备', desc: '注册账号 → 创建物模型 → 创建设备，获取 device_id 和 device_secret。' },
                { step: '02', icon: Wifi, title: 'MQTT 连接', desc: '以 device_id 作为 ClientID 和 Username，device_secret 作为 Password，连接到 1883 端口。' },
                { step: '03', icon: Activity, title: '上报数据', desc: '向 devices/{id}/telemetry/up 发布 JSON，平台自动校验、存储并实时推送。' },
              ].map(({ step, icon: Icon, title, desc }) => (
                <div key={step} style={{
                  padding: 20, border: '1px solid var(--color-border)',
                  borderRadius: 10, background: 'var(--color-card)',
                  position: 'relative', overflow: 'hidden',
                }}>
                  <div style={{
                    position: 'absolute', top: 10, right: 14,
                    fontSize: 36, fontWeight: 800, color: 'var(--color-primary)',
                    opacity: 0.08, fontFamily: 'monospace',
                  }}>{step}</div>
                  <div style={{
                    width: 34, height: 34, borderRadius: 8,
                    background: 'var(--color-primary)1a', border: '1px solid var(--color-primary)30',
                    display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: 12,
                  }}>
                    <Icon size={16} color="var(--color-primary)" />
                  </div>
                  <p style={{ margin: '0 0 6px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>{title}</p>
                  <p style={{ margin: 0, fontSize: 12, color: 'var(--color-muted-foreground)', lineHeight: 1.6 }}>{desc}</p>
                </div>
              ))}
            </div>

            <Divider />

            {/* ─── MQTT 连接参数 ─── */}
            <H2 id="mqtt-connect">MQTT 连接参数</H2>
            <P>平台内嵌 Mochi MQTT v2 Broker，使用标准 MQTT 3.1.1 协议，无需额外客户端库。</P>

            <CodeBlock lang="config  ·  连接参数" code={`Host:       your-server-ip
Port:       1883
ClientID:   {device_id}       # 设备 UUID（创建设备时生成）
Username:   {device_id}       # 与 ClientID 相同
Password:   {device_secret}   # 64 位随机密钥（创建设备时生成）
Protocol:   MQTT 3.1.1
KeepAlive:  60s`} />

            <H3>Python 示例</H3>
            <CodeBlock lang="python" code={`import paho.mqtt.client as mqtt
import json

DEVICE_ID     = "your-device-uuid"
DEVICE_SECRET = "your-64-char-secret"
SERVER        = "your-server-ip"

client = mqtt.Client(client_id=DEVICE_ID, protocol=mqtt.MQTTv311)
client.username_pw_set(DEVICE_ID, DEVICE_SECRET)
client.connect(SERVER, 1883, keepalive=60)

# 上报遥测数据
payload = json.dumps({"temperature": 25.6, "humidity": 68})
client.publish(f"devices/{DEVICE_ID}/telemetry/up", payload)`} />

            <Divider />

            {/* ─── 遥测数据上报 ─── */}
            <H2 id="telemetry-up">遥测数据上报</H2>
            <P>
              向 <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)', fontSize: 13 }}>devices/&#123;device_id&#125;/telemetry/up</code> 发布 JSON，
              平台根据绑定的物模型自动校验数据类型与范围。未绑定物模型的设备跳过校验直接存储。
            </P>

            <CodeBlock lang="json  ·  Topic: devices/{device_id}/telemetry/up" code={`{
  "temperature": 25.6,
  "humidity":    68,
  "switch":      true,
  "mode":        2
}`} />

            <H3>校验规则</H3>
            <div style={{
              padding: 16, borderRadius: 8, border: '1px solid var(--color-border)',
              background: 'var(--color-secondary)', marginBottom: 16,
            }}>
              {[
                ['int / float', '校验数值范围（min ~ max）'],
                ['bool', '只接受 true / false'],
                ['string', '校验字符串长度'],
                ['enum', '校验 value 是否在枚举列表中'],
              ].map(([type, rule]) => (
                <div key={type} style={{ display: 'flex', gap: 16, padding: '6px 0', borderBottom: '1px solid var(--color-border)', fontSize: 13 }}>
                  <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)', width: 100, flexShrink: 0 }}>{type}</code>
                  <span style={{ color: 'var(--color-muted-foreground)' }}>{rule}</span>
                </div>
              ))}
              <p style={{ margin: '10px 0 0', fontSize: 12, color: 'var(--color-muted-foreground)' }}>
                不合法字段标记 <code style={{ fontFamily: 'monospace' }}>valid=false</code> 后仍然存储（便于故障排查）。
              </p>
            </div>

            {/* ════════════════════ MQTT 协议 ════════════════════ */}
            <Divider />

            <H2 id="topic-list">Topic 结构总览</H2>
            <P>所有 Topic 以 <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)', fontSize: 13 }}>devices/&#123;device_id&#125;/</code> 为前缀，设备只能访问属于自己的 Topic。</P>

            <TopicTable rows={[
              { topic: 'devices/{id}/telemetry/up', dir: '设备→云端', desc: '上报遥测属性数据（温度、湿度、开关等）' },
              { topic: 'devices/{id}/telemetry/down', dir: '云端→设备', desc: '服务端下发属性设置指令' },
              { topic: 'devices/{id}/event', dir: '设备→云端', desc: '设备事件上报（报警、状态变更等）' },
              { topic: 'devices/{id}/service/invoke', dir: '云端→设备', desc: '服务调用下发（重启、校准等）' },
              { topic: 'devices/{id}/service/reply', dir: '设备→云端', desc: '服务执行结果回传' },
              { topic: 'devices/{id}/voice/up', dir: '设备→云端', desc: '语音文本上报，触发 NLP 解析执行' },
              { topic: 'devices/{id}/voice/down', dir: '云端→设备', desc: '语音指令执行结果回传' },
              { topic: 'devices/{id}/ota/down', dir: '云端→设备', desc: 'OTA 升级命令下发（含固件下载地址）' },
              { topic: 'devices/{id}/ota/up', dir: '设备→云端', desc: 'OTA 升级进度上报（0–100）' },
            ]} />

            <Divider />

            {/* ─── 事件上报 ─── */}
            <H2 id="event-up">事件上报</H2>
            <P>设备主动上报离散事件（如传感器告警、状态跳变），与遥测数据分开存储。</P>

            <CodeBlock lang="json  ·  Topic: devices/{device_id}/event" code={`{
  "event_id":  "high_temp_alarm",
  "severity":  "warning",
  "params": {
    "temperature": 85.2,
    "threshold":   80
  },
  "timestamp": 1708300000
}`} />

            <Divider />

            {/* ─── OTA 进度上报 ─── */}
            <H2 id="ota-progress">OTA 进度上报</H2>
            <P>
              设备收到 OTA 下发命令后，通过 HTTP 下载固件（Basic Auth：device_id:device_secret），
              下载与安装过程中持续上报进度。
            </P>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <CodeBlock lang="json  ·  OTA 下发命令（ota/down）" code={`{
  "task_id":      "550e8400-...",
  "firmware_id":  "abc123",
  "version":      "v1.2.0",
  "download_url": "http://server/api/firmwares/{id}/download",
  "sha256":       "e3b0c44298fc..."
}`} />
              <CodeBlock lang="json  ·  进度上报（ota/up）" code={`{
  "task_id":  "550e8400-...",
  "status":   "downloading",
  "progress": 45,
  "message":  "Downloading firmware..."
}

// status 可选值:
// downloading → installing → completed / failed`} />
            </div>

            <Divider />

            {/* ─── ACL ─── */}
            <H2 id="acl">ACL 访问控制</H2>
            <P>平台对每台设备的 MQTT 操作权限进行严格限制，防止设备间数据互访。</P>

            <div style={{
              padding: 20, border: '1px solid var(--color-border)', borderRadius: 10,
              background: 'var(--color-secondary)', display: 'flex', gap: 14, alignItems: 'flex-start',
            }}>
              <Shield size={18} color="var(--color-primary)" style={{ flexShrink: 0, marginTop: 2 }} />
              <div style={{ fontSize: 13, color: 'var(--color-muted-foreground)', lineHeight: 1.7 }}>
                <p style={{ margin: '0 0 8px', fontWeight: 600, color: 'var(--color-foreground)' }}>规则说明</p>
                <ul style={{ margin: 0, paddingLeft: 18 }}>
                  <li>设备只能 Subscribe / Publish 自己的 Topic（以 <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)' }}>devices/&#123;device_id&#125;/</code> 开头）</li>
                  <li>禁止设备 Publish 到 <code style={{ fontFamily: 'monospace' }}>/down</code>、<code style={{ fontFamily: 'monospace' }}>/invoke</code> 结尾的 Topic（仅服务端可写）</li>
                  <li>认证失败（ClientID / 密钥错误）直接拒绝连接，不返回错误详情</li>
                </ul>
              </div>
            </div>

            {/* ════════════════════ REST API ════════════════════ */}
            <Divider />

            <H2 id="api-auth">认证接口</H2>
            <P>
              登录成功后返回 JWT Token，有效期 24 小时。后续所有需认证的请求在
              <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)', fontSize: 13 }}> Authorization</code> 请求头携带
              <code style={{ fontFamily: 'monospace', fontSize: 13 }}> Bearer &lt;token&gt;</code>。
            </P>
            <ApiTable rows={[
              { method: 'POST', path: '/api/auth/register', desc: '用户注册（邮箱 + 密码）', auth: false },
              { method: 'POST', path: '/api/auth/login', desc: '用户登录，返回 JWT Token', auth: false },
              { method: 'POST', path: '/api/auth/logout', desc: '登出并将 Token 加入黑名单' },
              { method: 'PUT', path: '/api/auth/password', desc: '修改密码（需验证旧密码）' },
              { method: 'GET', path: '/api/me', desc: '获取当前登录用户信息' },
            ]} />

            <Divider />

            <H2 id="api-device">设备管理</H2>
            <P>设备数据受 RLS 行级安全保护，每个用户只能访问自己的设备。</P>
            <ApiTable rows={[
              { method: 'GET', path: '/api/devices', desc: '设备列表（含实时在线状态，来自 Redis）' },
              { method: 'POST', path: '/api/devices', desc: '创建设备，返回自动生成的 64 位密钥' },
              { method: 'GET', path: '/api/devices/:id', desc: '设备详情' },
              { method: 'PUT', path: '/api/devices/:id', desc: '更新设备名称、物模型绑定、元数据' },
              { method: 'DELETE', path: '/api/devices/:id', desc: '删除设备' },
              { method: 'GET', path: '/api/devices/:id/data/latest', desc: '设备最新一条遥测数据' },
              { method: 'GET', path: '/api/devices/:id/data/history', desc: '历史遥测（?start=&end=&limit=）' },
              { method: 'GET', path: '/api/devices/:id/data/export', desc: '历史数据导出为 CSV 文件' },
            ]} />

            <Divider />

            <H2 id="api-alert">告警与任务</H2>
            <ApiTable rows={[
              { method: 'GET', path: '/api/alert-rules', desc: '告警规则列表（?device_id= 按设备筛选）' },
              { method: 'POST', path: '/api/alert-rules', desc: '创建告警阈值规则（>/>=/</<=/==/!=）' },
              { method: 'PUT', path: '/api/alert-rules/:id', desc: '更新规则（含冷却时间 cooldown_minutes）' },
              { method: 'DELETE', path: '/api/alert-rules/:id', desc: '删除告警规则' },
              { method: 'GET', path: '/api/alert-logs', desc: '告警历史记录（?device_id= 筛选）' },
              { method: 'PUT', path: '/api/alert-logs/:id/acknowledge', desc: '标记告警为已确认' },
              { method: 'GET', path: '/api/scheduled-tasks', desc: '定时任务列表（?device_id= 筛选）' },
              { method: 'POST', path: '/api/scheduled-tasks', desc: '创建 Cron 定时任务（属性设置/服务调用）' },
              { method: 'PUT', path: '/api/scheduled-tasks/:id', desc: '更新定时任务' },
              { method: 'DELETE', path: '/api/scheduled-tasks/:id', desc: '删除定时任务' },
            ]} />

            <Divider />

            <H2 id="api-ota">OTA 固件升级</H2>
            <P>固件文件存储在服务端本地，设备通过 HTTP Basic Auth 下载（ClientID:Secret）。</P>
            <ApiTable rows={[
              { method: 'POST', path: '/api/firmwares', desc: '上传固件文件（multipart/form-data，含版本号）' },
              { method: 'GET', path: '/api/firmwares', desc: '固件列表（含文件大小、SHA256 校验值）' },
              { method: 'DELETE', path: '/api/firmwares/:id', desc: '删除固件（同时删除本地文件）' },
              { method: 'GET', path: '/api/firmwares/:id/download', desc: '设备下载固件（HTTP Basic Auth）', auth: false },
              { method: 'POST', path: '/api/ota-tasks', desc: '创建 OTA 升级任务，自动通过 MQTT 推送命令' },
              { method: 'GET', path: '/api/ota-tasks', desc: 'OTA 任务列表（?device_id=，含实时进度）' },
              { method: 'PUT', path: '/api/ota-tasks/:id/cancel', desc: '取消 OTA 升级任务' },
            ]} />

            <Divider />

            {/* ─── 请求示例 ─── */}
            <H2 id="api-example">请求示例</H2>

            <H3>登录并获取 Token</H3>
            <CodeBlock lang="bash" code={`curl -X POST http://your-server/api/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{"email": "user@example.com", "password": "your-password"}'

# 返回
{
  "code": 200,
  "data": { "token": "eyJhbGci..." }
}`} />

            <H3>携带 Token 查询设备列表</H3>
            <CodeBlock lang="bash" code={`curl http://your-server/api/devices \\
  -H "Authorization: Bearer eyJhbGci..."`} />

            <H3>查询历史遥测数据</H3>
            <CodeBlock lang="bash" code={`curl "http://your-server/api/devices/{device_id}/data/history?\\
  start=2024-01-01T00:00:00Z&\\
  end=2024-01-02T00:00:00Z&\\
  limit=500" \\
  -H "Authorization: Bearer eyJhbGci..."`} />

            <H3>创建设备</H3>
            <CodeBlock lang="bash" code={`curl -X POST http://your-server/api/devices \\
  -H "Authorization: Bearer eyJhbGci..." \\
  -H "Content-Type: application/json" \\
  -d '{
    "name":     "客厅温湿度传感器",
    "model_id": "your-thing-model-uuid"
  }'

# 返回（device_secret 仅在创建时返回，请立即保存）
{
  "code": 200,
  "data": {
    "id":            "550e8400-e29b-...",
    "device_secret": "a3f9b2c1..."
  }
}`} />

          </div>
        </main>
      </div>
    </div>
  );
}
