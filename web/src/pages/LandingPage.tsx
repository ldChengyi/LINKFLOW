import { useState, type ReactNode, type CSSProperties } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTheme } from '../hooks/useTheme';
import {
  Wifi, Bell, Clock, Shield,
  ChevronRight, Radio, Zap, Globe, Server,
  ArrowRight, Terminal, Database, Lock,
  Activity, Package, Layers, Palette,
  Check,
} from 'lucide-react';

// ── Helpers ─────────────────────────────────────────────────────────────────

function NavBar() {
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();

  const goToDashboard = () => {
    const token = localStorage.getItem('token');
    navigate(token ? '/devices' : '/login');
  };

  return (
    <header
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 50,
        borderBottom: '1px solid var(--color-border)',
        background: 'rgba(10,15,13,0.85)',
        backdropFilter: 'blur(12px)',
      }}
    >
      <div style={{ maxWidth: 1200, margin: '0 auto', padding: '0 24px' }}>
        <div style={{ display: 'flex', alignItems: 'center', height: 64, gap: 32 }}>
          {/* Logo */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
            <div style={{
              width: 34, height: 34, borderRadius: 8,
              background: 'var(--color-primary)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
            }}>
              <Radio size={18} color="var(--color-primary-foreground)" />
            </div>
            <span style={{ fontWeight: 700, fontSize: 18, color: 'var(--color-foreground)', letterSpacing: '-0.3px' }}>
              LinkFlow
            </span>
          </div>

          {/* Nav links */}
          <nav style={{ display: 'flex', gap: 4, flex: 1 }}>
            {[
              { label: '快速开始', href: '#quickstart' },
              { label: 'MQTT 协议', href: '#mqtt' },
              { label: 'REST API', href: '#rest-api' },
              { label: '平台功能', href: '#features' },
            ].map(item => (
              <a
                key={item.href}
                href={item.href}
                style={{
                  padding: '6px 12px',
                  borderRadius: 6,
                  fontSize: 14,
                  color: 'var(--color-muted-foreground)',
                  textDecoration: 'none',
                  transition: 'color 0.15s, background 0.15s',
                }}
                onMouseEnter={e => {
                  (e.target as HTMLElement).style.color = 'var(--color-foreground)';
                  (e.target as HTMLElement).style.background = 'var(--color-accent)';
                }}
                onMouseLeave={e => {
                  (e.target as HTMLElement).style.color = 'var(--color-muted-foreground)';
                  (e.target as HTMLElement).style.background = 'transparent';
                }}
              >
                {item.label}
              </a>
            ))}
          </nav>

          {/* Right side */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            {/* Theme toggle */}
            <button
              onClick={() => setTheme(theme === 'green' ? 'blue' : 'green')}
              style={{
                width: 34, height: 34, borderRadius: 8,
                border: '1px solid var(--color-border)',
                background: 'var(--color-card)',
                cursor: 'pointer',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                color: 'var(--color-muted-foreground)',
                transition: 'border-color 0.15s, color 0.15s',
              }}
              title={theme === 'green' ? '切换天蓝主题' : '切换翠绿主题'}
            >
              <Palette size={15} />
            </button>

            {/* Enter Dashboard */}
            <button
              onClick={goToDashboard}
              style={{
                display: 'flex', alignItems: 'center', gap: 6,
                padding: '8px 16px',
                borderRadius: 8,
                border: 'none',
                background: 'var(--color-primary)',
                color: 'var(--color-primary-foreground)',
                fontSize: 14,
                fontWeight: 600,
                cursor: 'pointer',
                transition: 'opacity 0.15s',
              }}
              onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
              onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
            >
              进入后台
              <ArrowRight size={14} />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}

// ── Section Wrapper ──────────────────────────────────────────────────────────

function Section({ id, children, style }: { id?: string; children: ReactNode; style?: CSSProperties }) {
  return (
    <section id={id} style={{ padding: '80px 24px', ...style }}>
      <div style={{ maxWidth: 1200, margin: '0 auto' }}>
        {children}
      </div>
    </section>
  );
}

function SectionTitle({ badge, title, desc }: { badge?: string; title: string; desc?: string }) {
  return (
    <div style={{ textAlign: 'center', marginBottom: 56 }}>
      {badge && (
        <span style={{
          display: 'inline-block',
          padding: '4px 12px',
          borderRadius: 20,
          border: '1px solid var(--color-primary)',
          color: 'var(--color-primary)',
          fontSize: 12,
          fontWeight: 600,
          letterSpacing: '0.5px',
          marginBottom: 16,
          textTransform: 'uppercase',
        }}>
          {badge}
        </span>
      )}
      <h2 style={{
        fontSize: 32,
        fontWeight: 700,
        color: 'var(--color-foreground)',
        margin: 0,
        marginBottom: desc ? 12 : 0,
        letterSpacing: '-0.5px',
      }}>
        {title}
      </h2>
      {desc && (
        <p style={{ fontSize: 16, color: 'var(--color-muted-foreground)', margin: 0 }}>
          {desc}
        </p>
      )}
    </div>
  );
}

// ── Code Block ───────────────────────────────────────────────────────────────

function CodeBlock({ lang, code }: { lang: string; code: string }) {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div style={{
      borderRadius: 10,
      border: '1px solid var(--color-border)',
      overflow: 'hidden',
      background: 'var(--color-card)',
    }}>
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '8px 16px',
        borderBottom: '1px solid var(--color-border)',
        background: 'var(--color-secondary)',
      }}>
        <span style={{ fontSize: 12, color: 'var(--color-muted-foreground)', fontFamily: 'monospace' }}>
          {lang}
        </span>
        <button
          onClick={copy}
          style={{
            background: 'none', border: 'none', cursor: 'pointer',
            color: copied ? 'var(--color-primary)' : 'var(--color-muted-foreground)',
            fontSize: 12, display: 'flex', alignItems: 'center', gap: 4,
          }}
        >
          {copied ? <><Check size={12} /> 已复制</> : '复制'}
        </button>
      </div>
      <pre style={{
        margin: 0,
        padding: '16px 20px',
        fontSize: 13,
        lineHeight: 1.7,
        color: 'var(--color-foreground)',
        fontFamily: "'Cascadia Code', 'Fira Code', 'JetBrains Mono', Consolas, monospace",
        overflowX: 'auto',
        whiteSpace: 'pre',
      }}>
        <code>{code}</code>
      </pre>
    </div>
  );
}

// ── API Table ─────────────────────────────────────────────────────────────────

function ApiTable({ rows }: { rows: { method: string; path: string; desc: string; auth?: boolean }[] }) {
  const methodColor: Record<string, string> = {
    GET: '#22c55e', POST: '#3b82f6', PUT: '#f59e0b', DELETE: '#ef4444',
  };

  return (
    <div style={{
      border: '1px solid var(--color-border)',
      borderRadius: 10,
      overflow: 'hidden',
    }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr style={{ background: 'var(--color-secondary)' }}>
            {['方法', '路径', '说明', '认证'].map(h => (
              <th key={h} style={{
                padding: '10px 16px', textAlign: 'left',
                color: 'var(--color-muted-foreground)',
                fontWeight: 600, fontSize: 12,
                borderBottom: '1px solid var(--color-border)',
              }}>
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={i}
              style={{
                borderBottom: i < rows.length - 1 ? '1px solid var(--color-border)' : 'none',
                background: i % 2 === 0 ? 'var(--color-card)' : 'transparent',
              }}
            >
              <td style={{ padding: '10px 16px' }}>
                <span style={{
                  display: 'inline-block',
                  padding: '2px 8px',
                  borderRadius: 4,
                  fontSize: 11,
                  fontWeight: 700,
                  fontFamily: 'monospace',
                  color: methodColor[row.method] || 'var(--color-foreground)',
                  background: (methodColor[row.method] || '#888') + '1a',
                }}>
                  {row.method}
                </span>
              </td>
              <td style={{ padding: '10px 16px' }}>
                <code style={{ fontSize: 12, color: 'var(--color-foreground)', fontFamily: 'monospace' }}>
                  {row.path}
                </code>
              </td>
              <td style={{ padding: '10px 16px', color: 'var(--color-muted-foreground)' }}>
                {row.desc}
              </td>
              <td style={{ padding: '10px 16px' }}>
                {row.auth !== false ? (
                  <span style={{ fontSize: 11, color: 'var(--color-primary)' }}>● JWT</span>
                ) : (
                  <span style={{ fontSize: 11, color: 'var(--color-muted-foreground)' }}>公开</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── MQTT Topic Table ──────────────────────────────────────────────────────────

function TopicTable({ rows }: { rows: { topic: string; dir: string; desc: string }[] }) {
  return (
    <div style={{
      border: '1px solid var(--color-border)',
      borderRadius: 10,
      overflow: 'hidden',
    }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr style={{ background: 'var(--color-secondary)' }}>
            {['Topic', '方向', '说明'].map(h => (
              <th key={h} style={{
                padding: '10px 16px', textAlign: 'left',
                color: 'var(--color-muted-foreground)',
                fontWeight: 600, fontSize: 12,
                borderBottom: '1px solid var(--color-border)',
              }}>
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr
              key={i}
              style={{
                borderBottom: i < rows.length - 1 ? '1px solid var(--color-border)' : 'none',
                background: i % 2 === 0 ? 'var(--color-card)' : 'transparent',
              }}
            >
              <td style={{ padding: '10px 16px' }}>
                <code style={{ fontSize: 12, color: 'var(--color-primary)', fontFamily: 'monospace' }}>
                  {row.topic}
                </code>
              </td>
              <td style={{ padding: '10px 16px' }}>
                <span style={{
                  display: 'inline-block',
                  padding: '2px 8px', borderRadius: 4, fontSize: 11, fontWeight: 600,
                  ...(row.dir === '设备→云端'
                    ? { color: '#22c55e', background: '#22c55e1a' }
                    : { color: '#3b82f6', background: '#3b82f61a' }),
                }}>
                  {row.dir}
                </span>
              </td>
              <td style={{ padding: '10px 16px', color: 'var(--color-muted-foreground)' }}>
                {row.desc}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── Main LandingPage ──────────────────────────────────────────────────────────

export default function LandingPage() {
  const navigate = useNavigate();

  const goToDashboard = () => {
    const token = localStorage.getItem('token');
    navigate(token ? '/devices' : '/login');
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--color-background)', color: 'var(--color-foreground)' }}>
      <NavBar />

      {/* ── Hero ──────────────────────────────────────────────────────────── */}
      <Section style={{ paddingTop: 100, paddingBottom: 100, position: 'relative', overflow: 'hidden' }}>
        {/* Background glow */}
        <div style={{
          position: 'absolute', top: -200, left: '50%', transform: 'translateX(-50%)',
          width: 600, height: 600, borderRadius: '50%',
          background: 'radial-gradient(circle, var(--color-primary)12, transparent 70%)',
          pointerEvents: 'none',
        }} />

        <div style={{ textAlign: 'center', position: 'relative' }}>
          <div style={{
            display: 'inline-flex', alignItems: 'center', gap: 8,
            padding: '6px 14px', borderRadius: 20,
            border: '1px solid var(--color-border)',
            background: 'var(--color-card)',
            fontSize: 13, color: 'var(--color-muted-foreground)',
            marginBottom: 28,
          }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--color-primary)', display: 'inline-block', animation: 'pulse 2s infinite' }} />
            轻量级物联网云平台
          </div>

          <h1 style={{
            fontSize: 'clamp(40px, 6vw, 72px)',
            fontWeight: 800,
            margin: '0 0 20px',
            letterSpacing: '-2px',
            lineHeight: 1.1,
            background: 'linear-gradient(135deg, var(--color-foreground) 0%, var(--color-primary) 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
          }}>
            LinkFlow
          </h1>

          <p style={{
            fontSize: 18,
            color: 'var(--color-muted-foreground)',
            maxWidth: 560,
            margin: '0 auto 40px',
            lineHeight: 1.7,
          }}>
            基于 Go 构建的物联网平台，支持 MQTT 设备接入、实时遥测、
            告警系统、OTA 固件升级与数据可视化。
          </p>

          {/* CTA */}
          <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap', marginBottom: 64 }}>
            <button
              onClick={goToDashboard}
              style={{
                display: 'flex', alignItems: 'center', gap: 8,
                padding: '12px 24px', borderRadius: 10,
                border: 'none',
                background: 'var(--color-primary)',
                color: 'var(--color-primary-foreground)',
                fontSize: 15, fontWeight: 600, cursor: 'pointer',
                transition: 'opacity 0.15s',
              }}
              onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
              onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
            >
              进入后台管理
              <ArrowRight size={16} />
            </button>
            <a
              href="#quickstart"
              style={{
                display: 'flex', alignItems: 'center', gap: 8,
                padding: '12px 24px', borderRadius: 10,
                border: '1px solid var(--color-border)',
                background: 'var(--color-card)',
                color: 'var(--color-foreground)',
                fontSize: 15, fontWeight: 500, cursor: 'pointer',
                textDecoration: 'none',
                transition: 'border-color 0.15s',
              }}
              onMouseEnter={e => ((e.currentTarget as HTMLElement).style.borderColor = 'var(--color-primary)')}
              onMouseLeave={e => ((e.currentTarget as HTMLElement).style.borderColor = 'var(--color-border)')}
            >
              <Terminal size={16} />
              查看接入文档
            </a>
          </div>

          {/* Stats */}
          <div style={{
            display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 1, maxWidth: 640, margin: '0 auto',
            border: '1px solid var(--color-border)', borderRadius: 12, overflow: 'hidden',
          }}>
            {[
              { icon: Wifi, label: 'MQTT Broker', value: 'TCP 1883' },
              { icon: Globe, label: 'REST API', value: '30+ 端点' },
              { icon: Activity, label: '实时推送', value: 'WebSocket' },
              { icon: Database, label: '时序存储', value: 'TimescaleDB' },
            ].map(({ icon: Icon, label, value }) => (
              <div
                key={label}
                style={{
                  padding: '20px 16px',
                  background: 'var(--color-card)',
                  textAlign: 'center',
                  borderRight: '1px solid var(--color-border)',
                }}
              >
                <Icon size={20} color="var(--color-primary)" style={{ marginBottom: 8 }} />
                <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-foreground)' }}>{value}</div>
                <div style={{ fontSize: 11, color: 'var(--color-muted-foreground)', marginTop: 2 }}>{label}</div>
              </div>
            ))}
          </div>
        </div>
      </Section>

      {/* ── Quick Start ─────────────────────────────────────────────────────── */}
      <Section id="quickstart" style={{ borderTop: '1px solid var(--color-border)' }}>
        <SectionTitle badge="快速开始" title="三步接入设备" desc="从注册到数据上报，整个流程不超过5分钟" />

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 24, marginBottom: 48 }}>
          {[
            {
              step: '01', icon: Lock, title: '注册并创建设备',
              desc: '在后台注册账号，创建物模型定义数据结构，然后创建设备获取 device_id 和 device_secret。',
            },
            {
              step: '02', icon: Wifi, title: '通过 MQTT 连接',
              desc: '使用设备凭证建立 MQTT 连接，ClientID 为 device_id，Password 为 device_secret，端口 1883。',
            },
            {
              step: '03', icon: Activity, title: '上报遥测数据',
              desc: '向 devices/{id}/telemetry/up 发布 JSON 数据，平台自动校验、存储，并实时推送到前端。',
            },
          ].map(({ step, icon: Icon, title, desc }) => (
            <div
              key={step}
              style={{
                padding: 28,
                border: '1px solid var(--color-border)',
                borderRadius: 12,
                background: 'var(--color-card)',
                position: 'relative',
                overflow: 'hidden',
              }}
            >
              <div style={{
                position: 'absolute', top: 16, right: 20,
                fontSize: 48, fontWeight: 800, color: 'var(--color-primary)',
                opacity: 0.08, lineHeight: 1, fontFamily: 'monospace',
              }}>
                {step}
              </div>
              <div style={{
                width: 40, height: 40, borderRadius: 10,
                background: 'var(--color-primary)1a',
                border: '1px solid var(--color-primary)33',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                marginBottom: 16,
              }}>
                <Icon size={18} color="var(--color-primary)" />
              </div>
              <h3 style={{ margin: '0 0 8px', fontSize: 16, fontWeight: 600, color: 'var(--color-foreground)' }}>
                {title}
              </h3>
              <p style={{ margin: 0, fontSize: 13, color: 'var(--color-muted-foreground)', lineHeight: 1.6 }}>
                {desc}
              </p>
            </div>
          ))}
        </div>

        {/* Connection example */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          <div>
            <p style={{ margin: '0 0 10px', fontSize: 13, fontWeight: 600, color: 'var(--color-muted-foreground)' }}>
              MQTT 连接参数
            </p>
            <CodeBlock
              lang="config"
              code={`Host:       your-server-ip
Port:       1883
ClientID:   {device_id}       # 设备 UUID
Username:   {device_id}       # 同 ClientID
Password:   {device_secret}   # 64位密钥（自动生成）
Protocol:   MQTT 3.1.1
KeepAlive:  60s`}
            />
          </div>
          <div>
            <p style={{ margin: '0 0 10px', fontSize: 13, fontWeight: 600, color: 'var(--color-muted-foreground)' }}>
              遥测数据上报（Publish）
            </p>
            <CodeBlock
              lang="json  ·  Topic: devices/{device_id}/telemetry/up"
              code={`{
  "temperature": 25.6,
  "humidity": 68,
  "switch": true,
  "mode": "auto"
}`}
            />
          </div>
        </div>
      </Section>

      {/* ── MQTT Documentation ───────────────────────────────────────────────── */}
      <Section id="mqtt" style={{ borderTop: '1px solid var(--color-border)', background: 'var(--color-card)' }}>
        <SectionTitle
          badge="MQTT 协议"
          title="Topic 结构说明"
          desc="平台使用标准 MQTT 3.1.1 协议，所有 Topic 以 devices/{device_id} 为前缀"
        />

        <TopicTable rows={[
          { topic: 'devices/{id}/telemetry/up', dir: '设备→云端', desc: '设备上报遥测属性数据（温度、湿度等）' },
          { topic: 'devices/{id}/telemetry/down', dir: '云端→设备', desc: '服务端下发属性设置指令' },
          { topic: 'devices/{id}/event', dir: '设备→云端', desc: '设备事件上报（报警、状态变更等）' },
          { topic: 'devices/{id}/service/invoke', dir: '云端→设备', desc: '服务调用下发（重启、校准等）' },
          { topic: 'devices/{id}/service/reply', dir: '设备→云端', desc: '服务执行结果回传' },
          { topic: 'devices/{id}/voice/up', dir: '设备→云端', desc: '语音文本上报，触发本地 NLP 解析' },
          { topic: 'devices/{id}/voice/down', dir: '云端→设备', desc: '语音指令执行结果回传' },
          { topic: 'devices/{id}/ota/down', dir: '云端→设备', desc: 'OTA 升级命令下发（含固件下载地址）' },
          { topic: 'devices/{id}/ota/up', dir: '设备→云端', desc: 'OTA 升级进度上报（0-100%）' },
        ]} />

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20, marginTop: 32 }}>
          <div>
            <p style={{ margin: '0 0 10px', fontSize: 13, fontWeight: 600, color: 'var(--color-muted-foreground)' }}>
              事件上报示例
            </p>
            <CodeBlock
              lang="json  ·  Topic: devices/{device_id}/event"
              code={`{
  "event_id": "high_temp_alarm",
  "severity": "warning",
  "params": {
    "temperature": 85.2,
    "threshold": 80
  },
  "timestamp": 1708300000
}`}
            />
          </div>
          <div>
            <p style={{ margin: '0 0 10px', fontSize: 13, fontWeight: 600, color: 'var(--color-muted-foreground)' }}>
              OTA 进度上报示例
            </p>
            <CodeBlock
              lang="json  ·  Topic: devices/{device_id}/ota/up"
              code={`{
  "task_id": "550e8400-e29b-41d4-a716",
  "status": "downloading",
  "progress": 45,
  "message": "Downloading firmware..."
}`}
            />
          </div>
        </div>

        {/* ACL Rules */}
        <div style={{
          marginTop: 32,
          padding: 20,
          border: '1px solid var(--color-border)',
          borderRadius: 10,
          background: 'var(--color-secondary)',
          display: 'flex', gap: 16, alignItems: 'flex-start',
        }}>
          <Shield size={18} color="var(--color-warning)" style={{ flexShrink: 0, marginTop: 2 }} />
          <div>
            <p style={{ margin: '0 0 6px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>
              ACL 访问控制
            </p>
            <p style={{ margin: 0, fontSize: 13, color: 'var(--color-muted-foreground)', lineHeight: 1.6 }}>
              每台设备只能发布/订阅自己的 Topic（以 <code style={{ fontFamily: 'monospace', color: 'var(--color-primary)' }}>devices/&#123;device_id&#125;/</code> 为前缀）。
              禁止设备发布到 <code style={{ fontFamily: 'monospace' }}>/down</code> 和 <code style={{ fontFamily: 'monospace' }}>/invoke</code> 结尾的 Topic（仅服务端可写）。
            </p>
          </div>
        </div>
      </Section>

      {/* ── REST API ─────────────────────────────────────────────────────────── */}
      <Section id="rest-api" style={{ borderTop: '1px solid var(--color-border)' }}>
        <SectionTitle
          badge="REST API"
          title="HTTP 接口文档"
          desc="所有 API 均以 /api 为前缀，需认证接口在 Authorization 请求头携带 Bearer Token"
        />

        <div style={{ marginBottom: 32 }}>
          <p style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>
            认证接口
          </p>
          <ApiTable rows={[
            { method: 'POST', path: '/api/auth/register', desc: '用户注册（邮箱 + 密码）', auth: false },
            { method: 'POST', path: '/api/auth/login', desc: '用户登录，返回 JWT Token', auth: false },
            { method: 'POST', path: '/api/auth/logout', desc: '登出并将 Token 加入黑名单' },
            { method: 'PUT', path: '/api/auth/password', desc: '修改密码（需验证旧密码）' },
            { method: 'GET', path: '/api/me', desc: '获取当前登录用户信息' },
          ]} />
        </div>

        <div style={{ marginBottom: 32 }}>
          <p style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>
            设备管理
          </p>
          <ApiTable rows={[
            { method: 'GET', path: '/api/devices', desc: '设备列表（含实时在线状态）' },
            { method: 'POST', path: '/api/devices', desc: '创建设备，自动生成 64 位密钥' },
            { method: 'GET', path: '/api/devices/:id', desc: '设备详情' },
            { method: 'PUT', path: '/api/devices/:id', desc: '更新设备信息' },
            { method: 'DELETE', path: '/api/devices/:id', desc: '删除设备' },
            { method: 'GET', path: '/api/devices/:id/data/latest', desc: '设备最新一条遥测数据' },
            { method: 'GET', path: '/api/devices/:id/data/history', desc: '历史遥测数据（时间范围查询）' },
            { method: 'GET', path: '/api/devices/:id/data/export', desc: '历史数据导出为 CSV 文件' },
          ]} />
        </div>

        <div style={{ marginBottom: 32 }}>
          <p style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>
            告警与任务
          </p>
          <ApiTable rows={[
            { method: 'GET', path: '/api/alert-rules', desc: '告警规则列表（支持设备筛选）' },
            { method: 'POST', path: '/api/alert-rules', desc: '创建告警阈值规则' },
            { method: 'GET', path: '/api/alert-logs', desc: '告警历史记录' },
            { method: 'PUT', path: '/api/alert-logs/:id/acknowledge', desc: '标记告警为已确认' },
            { method: 'GET', path: '/api/scheduled-tasks', desc: '定时任务列表' },
            { method: 'POST', path: '/api/scheduled-tasks', desc: '创建 Cron 定时任务' },
          ]} />
        </div>

        <div>
          <p style={{ margin: '0 0 12px', fontSize: 14, fontWeight: 600, color: 'var(--color-foreground)' }}>
            OTA 固件升级
          </p>
          <ApiTable rows={[
            { method: 'POST', path: '/api/firmwares', desc: '上传固件文件（multipart/form-data）' },
            { method: 'GET', path: '/api/firmwares', desc: '固件列表（含 SHA256 校验值）' },
            { method: 'GET', path: '/api/firmwares/:id/download', desc: '设备下载固件（HTTP Basic Auth）', auth: false },
            { method: 'POST', path: '/api/ota-tasks', desc: '创建 OTA 升级任务，自动推送 MQTT 指令' },
            { method: 'GET', path: '/api/ota-tasks', desc: 'OTA 任务列表（含实时进度）' },
            { method: 'PUT', path: '/api/ota-tasks/:id/cancel', desc: '取消 OTA 升级任务' },
          ]} />
        </div>

        {/* Auth example */}
        <div style={{ marginTop: 32 }}>
          <p style={{ margin: '0 0 10px', fontSize: 13, fontWeight: 600, color: 'var(--color-muted-foreground)' }}>
            登录 & Token 使用示例
          </p>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
            <CodeBlock
              lang="bash  ·  登录获取 Token"
              code={`curl -X POST /api/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "password": "your-password"
  }'

# 返回
{
  "code": 200,
  "data": {
    "token": "eyJhbGci..."
  }
}`}
            />
            <CodeBlock
              lang="bash  ·  携带 Token 请求"
              code={`# 获取设备列表
curl /api/devices \\
  -H "Authorization: Bearer eyJhbGci..."

# 查询历史遥测数据
curl "/api/devices/{id}/data/history?\\
  start=2024-01-01T00:00:00Z&\\
  end=2024-01-02T00:00:00Z&\\
  limit=500" \\
  -H "Authorization: Bearer eyJhbGci..."`}
            />
          </div>
        </div>
      </Section>

      {/* ── Features ─────────────────────────────────────────────────────────── */}
      <Section id="features" style={{ borderTop: '1px solid var(--color-border)', background: 'var(--color-card)' }}>
        <SectionTitle badge="平台功能" title="开箱即用的 IoT 能力" />

        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 20,
        }}>
          {[
            {
              icon: Wifi, title: 'MQTT 设备接入',
              desc: '内嵌 Mochi MQTT v2 Broker，支持 TCP 1883 接入，基于物模型的数据校验（类型/范围/枚举），不合法数据标注后仍存储。',
              tags: ['Mochi MQTT v2', '物模型校验', 'ACL'],
            },
            {
              icon: Activity, title: '实时遥测监控',
              desc: 'TimescaleDB 时序数据库存储遥测数据，WebSocket 实时推送，支持历史趋势图表（1h/6h/24h/7d 范围切换）。',
              tags: ['TimescaleDB', 'WebSocket', 'recharts'],
            },
            {
              icon: Bell, title: '智能告警系统',
              desc: '阈值告警规则（>/>=/</<=/==/!=），告警冷却防抖，WebSocket 实时通知，告警确认流程与历史审计。',
              tags: ['阈值规则', '告警冷却', '未读角标'],
            },
            {
              icon: Package, title: 'OTA 固件升级',
              desc: '固件文件管理（SHA256校验），MQTT 指令推送，设备 HTTP 下载（Basic Auth），实时进度上报与 WS 推送。',
              tags: ['SHA256', 'HTTP 下载', '进度实时'],
            },
            {
              icon: Clock, title: '定时任务调度',
              desc: 'Cron 表达式驱动，支持属性设置与服务调用，goroutine 每分钟评估，任务执行历史完整记录。',
              tags: ['Cron', 'MQTT下发', '执行历史'],
            },
            {
              icon: Zap, title: '语音控制（NLP）',
              desc: '本地关键词匹配引擎，支持开关控制、数值设置、相对调节（调高/调低）、枚举匹配和服务调用。',
              tags: ['本地NLP', '无云依赖', 'MQTT'],
            },
            {
              icon: Layers, title: '物模型系统',
              desc: '类阿里云物模型结构（属性/事件/服务），支持多种数据类型，功能模块绑定配置，RLS 行级安全隔离。',
              tags: ['Properties', 'Events', 'Services'],
            },
            {
              icon: Shield, title: '安全与审计',
              desc: 'JWT + Redis Token 黑名单，PostgreSQL RLS 行级数据隔离，操作审计日志，设备 ACL 访问控制。',
              tags: ['JWT', 'RLS', '审计日志'],
            },
            {
              icon: Server, title: '设备在线调试',
              desc: '区分真实 MQTT 连接与模拟上线，模拟设备自动回传数据，支持属性设置与服务调用，心跳续期机制。',
              tags: ['模拟上线', '心跳续期', '属性下发'],
            },
          ].map(({ icon: Icon, title, desc, tags }) => (
            <div
              key={title}
              style={{
                padding: 24,
                border: '1px solid var(--color-border)',
                borderRadius: 12,
                background: 'var(--color-background)',
                display: 'flex', flexDirection: 'column', gap: 12,
                transition: 'border-color 0.2s, transform 0.2s',
              }}
              onMouseEnter={e => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--color-primary)';
                (e.currentTarget as HTMLElement).style.transform = 'translateY(-2px)';
              }}
              onMouseLeave={e => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--color-border)';
                (e.currentTarget as HTMLElement).style.transform = 'translateY(0)';
              }}
            >
              <div style={{
                width: 38, height: 38, borderRadius: 9,
                background: 'var(--color-primary)18',
                border: '1px solid var(--color-primary)28',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <Icon size={17} color="var(--color-primary)" />
              </div>
              <div>
                <h3 style={{ margin: '0 0 6px', fontSize: 15, fontWeight: 600, color: 'var(--color-foreground)' }}>
                  {title}
                </h3>
                <p style={{ margin: 0, fontSize: 13, color: 'var(--color-muted-foreground)', lineHeight: 1.6 }}>
                  {desc}
                </p>
              </div>
              <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginTop: 'auto' }}>
                {tags.map(tag => (
                  <span key={tag} style={{
                    fontSize: 11, padding: '2px 8px', borderRadius: 4,
                    border: '1px solid var(--color-border)',
                    color: 'var(--color-muted-foreground)',
                    background: 'var(--color-secondary)',
                  }}>
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </Section>

      {/* ── Tech Stack ───────────────────────────────────────────────────────── */}
      <Section style={{ borderTop: '1px solid var(--color-border)', paddingTop: 56, paddingBottom: 56 }}>
        <div style={{ textAlign: 'center', marginBottom: 40 }}>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--color-muted-foreground)', textTransform: 'uppercase', letterSpacing: '1px', fontWeight: 600 }}>
            技术栈
          </p>
        </div>
        <div style={{ display: 'flex', gap: 12, justifyContent: 'center', flexWrap: 'wrap' }}>
          {[
            'Go 1.24', 'Gin', 'Mochi MQTT v2', 'PostgreSQL 15', 'TimescaleDB',
            'Redis 7', 'React 18', 'TypeScript', 'Tailwind CSS', 'shadcn/ui',
            'gorilla/websocket', 'pgx v5', 'Docker',
          ].map(tech => (
            <span
              key={tech}
              style={{
                padding: '8px 16px', borderRadius: 8,
                border: '1px solid var(--color-border)',
                background: 'var(--color-card)',
                fontSize: 13, color: 'var(--color-foreground)',
                fontWeight: 500,
              }}
            >
              {tech}
            </span>
          ))}
        </div>
      </Section>

      {/* ── CTA Banner ───────────────────────────────────────────────────────── */}
      <Section style={{
        borderTop: '1px solid var(--color-border)',
        background: 'linear-gradient(135deg, var(--color-card) 0%, var(--color-secondary) 100%)',
        paddingTop: 64, paddingBottom: 64,
      }}>
        <div style={{ textAlign: 'center' }}>
          <h2 style={{ fontSize: 32, fontWeight: 700, margin: '0 0 12px', letterSpacing: '-0.5px' }}>
            开始使用 LinkFlow
          </h2>
          <p style={{ fontSize: 16, color: 'var(--color-muted-foreground)', margin: '0 0 32px' }}>
            登录后台，创建物模型，接入您的第一台设备
          </p>
          <button
            onClick={goToDashboard}
            style={{
              display: 'inline-flex', alignItems: 'center', gap: 8,
              padding: '14px 32px', borderRadius: 10,
              border: 'none',
              background: 'var(--color-primary)',
              color: 'var(--color-primary-foreground)',
              fontSize: 16, fontWeight: 600, cursor: 'pointer',
              transition: 'opacity 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
            onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
          >
            进入后台管理
            <ChevronRight size={18} />
          </button>
        </div>
      </Section>

      {/* ── Footer ───────────────────────────────────────────────────────────── */}
      <footer style={{
        borderTop: '1px solid var(--color-border)',
        padding: '24px',
        textAlign: 'center',
      }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Radio size={14} color="var(--color-primary)" />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-foreground)' }}>LinkFlow</span>
            <span style={{ fontSize: 13, color: 'var(--color-muted-foreground)' }}>· 物联网云平台</span>
          </div>
          <div style={{ display: 'flex', gap: 24 }}>
            {[
              { label: '快速开始', href: '#quickstart' },
              { label: 'MQTT 协议', href: '#mqtt' },
              { label: 'REST API', href: '#rest-api' },
            ].map(link => (
              <a key={link.href} href={link.href} style={{
                fontSize: 13, color: 'var(--color-muted-foreground)', textDecoration: 'none',
                transition: 'color 0.15s',
              }}
                onMouseEnter={e => ((e.target as HTMLElement).style.color = 'var(--color-foreground)')}
                onMouseLeave={e => ((e.target as HTMLElement).style.color = 'var(--color-muted-foreground)')}
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
      </footer>
    </div>
  );
}
