import { useNavigate, Link } from 'react-router-dom';
import { useTheme } from '../hooks/useTheme';
import { Radio, ArrowLeft, ArrowRight, Palette } from 'lucide-react';

interface SiteNavBarProps {
  /** 'home' 显示页内导航链接；'docs' 显示"← 返回首页"和"接入文档"徽章 */
  page: 'home' | 'docs';
}

const HEADER_BG = 'rgba(10,15,13,0.88)';
const WHITE_PRIMARY = '#ffffff';
const WHITE_DIM = 'rgba(255,255,255,0.5)';
const WHITE_DIMMER = 'rgba(255,255,255,0.18)';
const WHITE_FAINT = 'rgba(255,255,255,0.06)';

export default function SiteNavBar({ page }: SiteNavBarProps) {
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();

  const goToDashboard = () => {
    const token = localStorage.getItem('token');
    navigate(token ? '/devices' : '/login');
  };

  return (
    <header style={{
      position: 'sticky', top: 0, zIndex: 50,
      borderBottom: '1px solid rgba(255,255,255,0.08)',
      background: HEADER_BG,
      backdropFilter: 'blur(14px)',
      height: 60,
      display: 'flex', alignItems: 'center',
    }}>
      <div style={{
        maxWidth: 1400, margin: '0 auto', padding: '0 28px',
        width: '100%', display: 'flex', alignItems: 'center', gap: 12,
      }}>

        {/* Docs only: 返回首页 */}
        {page === 'docs' && (
          <>
            <button
              onClick={() => navigate('/')}
              style={{
                display: 'flex', alignItems: 'center', gap: 5,
                background: 'none', border: 'none', cursor: 'pointer',
                color: WHITE_DIM, fontSize: 13,
                padding: '4px 8px', borderRadius: 6,
                transition: 'color 0.15s', flexShrink: 0,
              }}
              onMouseEnter={e => (e.currentTarget.style.color = WHITE_PRIMARY)}
              onMouseLeave={e => (e.currentTarget.style.color = WHITE_DIM)}
            >
              <ArrowLeft size={14} />
              返回首页
            </button>
            <div style={{ width: 1, height: 18, background: WHITE_DIMMER, flexShrink: 0 }} />
          </>
        )}

        {/* Logo */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 9, flexShrink: 0 }}>
          <div style={{
            width: 32, height: 32, borderRadius: 8,
            background: 'var(--color-primary)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <Radio size={16} color="#fff" />
          </div>
          <span style={{ fontWeight: 700, fontSize: 16, color: WHITE_PRIMARY, letterSpacing: '-0.2px' }}>
            LinkFlow
          </span>
          {page === 'docs' && (
            <span style={{
              fontSize: 11, padding: '2px 8px', borderRadius: 10,
              border: `1px solid ${WHITE_DIMMER}`,
              color: 'rgba(255,255,255,0.4)',
              letterSpacing: '0.2px',
            }}>
              接入文档
            </span>
          )}
        </div>

        {/* Home only: nav links */}
        {page === 'home' && (
          <nav style={{ display: 'flex', gap: 2, marginLeft: 16 }}>
            {[
              { label: '快速开始', to: '/docs', isLink: true },
              { label: '平台功能', to: '#features', isLink: false },
            ].map(item => (
              item.isLink ? (
                <Link
                  key={item.label}
                  to={item.to}
                  style={{
                    padding: '5px 12px', borderRadius: 6, fontSize: 14,
                    color: WHITE_DIM, textDecoration: 'none',
                    transition: 'color 0.15s, background 0.15s',
                  }}
                  onMouseEnter={e => {
                    (e.currentTarget as HTMLElement).style.color = WHITE_PRIMARY;
                    (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.07)';
                  }}
                  onMouseLeave={e => {
                    (e.currentTarget as HTMLElement).style.color = WHITE_DIM;
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  {item.label}
                </Link>
              ) : (
                <a
                  key={item.label}
                  href={item.to}
                  style={{
                    padding: '5px 12px', borderRadius: 6, fontSize: 14,
                    color: WHITE_DIM, textDecoration: 'none',
                    transition: 'color 0.15s, background 0.15s',
                  }}
                  onMouseEnter={e => {
                    (e.currentTarget as HTMLElement).style.color = WHITE_PRIMARY;
                    (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.07)';
                  }}
                  onMouseLeave={e => {
                    (e.currentTarget as HTMLElement).style.color = WHITE_DIM;
                    (e.currentTarget as HTMLElement).style.background = 'transparent';
                  }}
                >
                  {item.label}
                </a>
              )
            ))}
          </nav>
        )}

        <div style={{ flex: 1 }} />

        {/* Theme toggle */}
        <button
          onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
          title={theme === 'dark' ? '切换浅色主题' : '切换深色主题'}
          style={{
            width: 34, height: 34, borderRadius: 8,
            border: `1px solid ${WHITE_DIMMER}`,
            background: WHITE_FAINT,
            cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: WHITE_DIM,
            transition: 'border-color 0.15s, color 0.15s',
          }}
          onMouseEnter={e => {
            e.currentTarget.style.borderColor = 'rgba(255,255,255,0.35)';
            e.currentTarget.style.color = WHITE_PRIMARY;
          }}
          onMouseLeave={e => {
            e.currentTarget.style.borderColor = WHITE_DIMMER;
            e.currentTarget.style.color = WHITE_DIM;
          }}
        >
          <Palette size={15} />
        </button>

        {/* Enter dashboard */}
        <button
          onClick={goToDashboard}
          style={{
            display: 'flex', alignItems: 'center', gap: 6,
            padding: '7px 16px', borderRadius: 8, border: 'none',
            background: 'var(--color-primary)', color: '#fff',
            fontSize: 13, fontWeight: 600, cursor: 'pointer',
            transition: 'opacity 0.15s',
          }}
          onMouseEnter={e => (e.currentTarget.style.opacity = '0.85')}
          onMouseLeave={e => (e.currentTarget.style.opacity = '1')}
        >
          进入后台
          {page === 'home' && <ArrowRight size={14} />}
        </button>

      </div>
    </header>
  );
}
