import { type ReactNode, type CSSProperties } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  Wifi, Bell, Clock, Shield,
  ChevronRight, Radio, Zap, Globe, Server,
  ArrowRight, Terminal, Database,
  Activity, Package, Layers,
} from 'lucide-react';
import SiteNavBar from '../components/SiteNavBar';

// ── Helpers ───────────────────────────────────────────────────────────────────

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
          display: 'inline-block', padding: '4px 12px', borderRadius: 20,
          border: '1px solid var(--color-primary)', color: 'var(--color-primary)',
          fontSize: 12, fontWeight: 600, letterSpacing: '0.5px',
          marginBottom: 16, textTransform: 'uppercase',
        }}>
          {badge}
        </span>
      )}
      <h2 style={{
        fontSize: 32, fontWeight: 700, color: 'var(--color-foreground)',
        margin: 0, marginBottom: desc ? 12 : 0, letterSpacing: '-0.5px',
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

// ── Main ──────────────────────────────────────────────────────────────────────

export default function LandingPage() {
  const navigate = useNavigate();

  const goToDashboard = () => {
    const token = localStorage.getItem('token');
    navigate(token ? '/devices' : '/login');
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--color-background)', color: 'var(--color-foreground)' }}>
      <SiteNavBar page="home" />

      {/* ── Hero ── */}
      <Section style={{ paddingTop: 100, paddingBottom: 100, position: 'relative', overflow: 'hidden' }}>
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
            border: '1px solid var(--color-border)', background: 'var(--color-card)',
            fontSize: 13, color: 'var(--color-muted-foreground)', marginBottom: 28,
          }}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--color-primary)', display: 'inline-block' }} />
            轻量级物联网云平台
          </div>

          <h1 style={{
            fontSize: 'clamp(40px, 6vw, 72px)', fontWeight: 800,
            margin: '0 0 20px', letterSpacing: '-2px', lineHeight: 1.1,
            background: 'linear-gradient(135deg, var(--color-foreground) 0%, var(--color-primary) 100%)',
            WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
          }}>
            LinkFlow
          </h1>

          <p style={{
            fontSize: 18, color: 'var(--color-muted-foreground)',
            maxWidth: 560, margin: '0 auto 40px', lineHeight: 1.7,
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
                padding: '12px 24px', borderRadius: 10, border: 'none',
                background: 'var(--color-primary)', color: 'var(--color-primary-foreground)',
                fontSize: 15, fontWeight: 600, cursor: 'pointer', transition: 'opacity 0.15s',
              }}
              onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
              onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
            >
              进入后台管理 <ArrowRight size={16} />
            </button>
            <Link
              to="/docs"
              style={{
                display: 'flex', alignItems: 'center', gap: 8,
                padding: '12px 24px', borderRadius: 10,
                border: '1px solid var(--color-border)', background: 'var(--color-card)',
                color: 'var(--color-foreground)', fontSize: 15, fontWeight: 500,
                textDecoration: 'none', transition: 'border-color 0.15s',
              }}
              onMouseEnter={e => ((e.currentTarget as HTMLElement).style.borderColor = 'var(--color-primary)')}
              onMouseLeave={e => ((e.currentTarget as HTMLElement).style.borderColor = 'var(--color-border)')}
            >
              <Terminal size={16} />
              查看接入文档
            </Link>
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
              <div key={label} style={{
                padding: '20px 16px', background: 'var(--color-card)',
                textAlign: 'center', borderRight: '1px solid var(--color-border)',
              }}>
                <Icon size={20} color="var(--color-primary)" style={{ marginBottom: 8 }} />
                <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--color-foreground)' }}>{value}</div>
                <div style={{ fontSize: 11, color: 'var(--color-muted-foreground)', marginTop: 2 }}>{label}</div>
              </div>
            ))}
          </div>
        </div>
      </Section>

      {/* ── Features ── */}
      <Section id="features" style={{ borderTop: '1px solid var(--color-border)', background: 'var(--color-card)' }}>
        <SectionTitle badge="平台功能" title="开箱即用的 IoT 能力" />

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 20 }}>
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
                padding: 24, border: '1px solid var(--color-border)', borderRadius: 12,
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
                background: 'var(--color-primary)18', border: '1px solid var(--color-primary)28',
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
                    color: 'var(--color-muted-foreground)', background: 'var(--color-secondary)',
                  }}>{tag}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </Section>

      {/* ── Tech Stack ── */}
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
            <span key={tech} style={{
              padding: '8px 16px', borderRadius: 8,
              border: '1px solid var(--color-border)', background: 'var(--color-card)',
              fontSize: 13, color: 'var(--color-foreground)', fontWeight: 500,
            }}>{tech}</span>
          ))}
        </div>
      </Section>

      {/* ── CTA Banner ── */}
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
              padding: '14px 32px', borderRadius: 10, border: 'none',
              background: 'var(--color-primary)', color: 'var(--color-primary-foreground)',
              fontSize: 16, fontWeight: 600, cursor: 'pointer', transition: 'opacity 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
            onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
          >
            进入后台管理 <ChevronRight size={18} />
          </button>
        </div>
      </Section>

      {/* ── Footer ── */}
      <footer style={{ borderTop: '1px solid var(--color-border)', padding: '24px', textAlign: 'center' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <Radio size={14} color="var(--color-primary)" />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--color-foreground)' }}>LinkFlow</span>
            <span style={{ fontSize: 13, color: 'var(--color-muted-foreground)' }}>· 物联网云平台</span>
          </div>
          <div style={{ display: 'flex', gap: 24 }}>
            {[
              { label: '接入文档', href: '/docs', isRoute: true },
              { label: '平台功能', href: '#features', isRoute: false },
            ].map(link => (
              link.isRoute ? (
                <Link key={link.href} to={link.href} style={{
                  fontSize: 13, color: 'var(--color-muted-foreground)', textDecoration: 'none',
                }}>
                  {link.label}
                </Link>
              ) : (
                <a key={link.href} href={link.href} style={{
                  fontSize: 13, color: 'var(--color-muted-foreground)', textDecoration: 'none',
                }}
                  onMouseEnter={e => ((e.target as HTMLElement).style.color = 'var(--color-foreground)')}
                  onMouseLeave={e => ((e.target as HTMLElement).style.color = 'var(--color-muted-foreground)')}
                >
                  {link.label}
                </a>
              )
            ))}
          </div>
        </div>
      </footer>
    </div>
  );
}
