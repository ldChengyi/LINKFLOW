import { useEffect, useState, useCallback } from 'react';
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { toast } from 'sonner';
import {
  LayoutDashboard, Box, Cpu, Activity, Blocks, ScrollText, LogOut, Menu, X, ChevronDown, Leaf, User, Bell, AlertTriangle, Clock, Palette,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { authApi, statsApi } from '../api';
import type { StatsOverview } from '../api';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useWebSocket } from '../hooks/useWebSocket';
import { useTheme } from '../hooks/useTheme';
import type { WSMessage } from '../hooks/useWebSocket';
import ThingModelList from './ThingModelList';
import ThingModelForm from './ThingModelForm';
import DeviceList from './DeviceList';
import DeviceForm from './DeviceForm';
import DeviceData from './DeviceData';
import ModuleList from './ModuleList';
import AuditLogList from './AuditLogList';
import AlertRuleList from './AlertRuleList';
import AlertRuleForm from './AlertRuleForm';
import AlertLogList from './AlertLogList';
import ScheduledTaskList from './ScheduledTaskList';
import ScheduledTaskForm from './ScheduledTaskForm';

interface UserInfo {
  user_id: string;
  role: string;
}

const navItems = [
  { key: 'dashboard', label: '仪表盘', icon: LayoutDashboard, path: '/' },
  { key: 'thing-models', label: '物模型', icon: Box, path: '/thing-models' },
  { key: 'devices', label: '设备管理', icon: Cpu, path: '/devices' },
  { key: 'device-data', label: '设备数据', icon: Activity, path: '/device-data' },
  { key: 'alert-rules', label: '告警规则', icon: Bell, path: '/alert-rules' },
  { key: 'alert-logs', label: '告警历史', icon: AlertTriangle, path: '/alert-logs' },
  { key: 'scheduled-tasks', label: '定时任务', icon: Clock, path: '/scheduled-tasks' },
  { key: 'modules', label: '功能模块', icon: Blocks, path: '/modules' },
  { key: 'audit-logs', label: '审计日志', icon: ScrollText, path: '/audit-logs' },
];

// 全局 WS 事件总线，子页面可以订阅
type WSListener = (msg: WSMessage) => void;
const wsListeners = new Set<WSListener>();
export function onWSMessage(listener: WSListener) {
  wsListeners.add(listener);
  return () => { wsListeners.delete(listener); };
}

function WelcomePage({ user, stats }: { user: UserInfo | null; stats: StatsOverview | null }) {
  const statItems = [
    { label: '设备总数', value: stats?.total_devices ?? '--', color: 'text-primary' },
    { label: '在线设备', value: stats?.online_devices ?? '--', color: 'text-emerald-400' },
    { label: '物模型', value: stats?.total_thing_models ?? '--', color: 'text-blue-400' },
  ];

  return (
    <div className="flex flex-col items-center justify-center py-20">
      <div className="p-4 rounded-2xl bg-primary/10 mb-6">
        <Leaf className="h-12 w-12 text-primary" />
      </div>
      <h2 className="text-3xl font-bold mb-2">欢迎使用 LinkFlow</h2>
      <p className="text-muted-foreground text-lg mb-6">物联网云平台管理系统</p>
      {user && (
        <p className="text-sm text-muted-foreground font-mono bg-muted px-4 py-2 rounded-lg">
          用户ID: {user.user_id}
        </p>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-12 w-full max-w-2xl">
        {statItems.map((stat) => (
          <div key={stat.label} className="bg-card border border-border rounded-lg p-6 text-center">
            <p className={cn('text-3xl font-bold', stat.color)}>{stat.value}</p>
            <p className="text-sm text-muted-foreground mt-1">{stat.label}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function Dashboard() {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [collapsed, setCollapsed] = useState(false);
  const [stats, setStats] = useState<StatsOverview | null>(null);
  const { theme, setTheme } = useTheme();

  useEffect(() => {
    fetchUser();
    fetchStats();
  }, []);

  const fetchUser = async () => {
    try {
      const { data } = await authApi.me();
      setUser(data);
    } catch {
      // handled by interceptor
    }
  };

  const fetchStats = async () => {
    try {
      const { data } = await statsApi.overview();
      setStats(data);
    } catch { /* silent */ }
  };

  // WebSocket 消息处理
  const handleWSMessage = useCallback((msg: WSMessage) => {
    // 统计更新
    if (msg.type === 'stats') {
      setStats((prev) => prev ? { ...prev, online_devices: msg.data.online_devices } : prev);
    }

    // 告警 toast
    if (msg.type === 'alert') {
      const d = msg.data;
      const severityMap: Record<string, 'warning' | 'error' | 'info'> = {
        critical: 'error', warning: 'warning', info: 'info',
      };
      const method = severityMap[d.severity] || 'warning';
      toast[method](`告警: ${d.rule_name}`, {
        description: `${d.device_name} · ${d.property_name} = ${d.actual_value} ${d.operator} ${d.threshold}`,
        duration: 8000,
      });
    }

    // 广播给子页面
    wsListeners.forEach((fn) => fn(msg));
  }, []);

  useWebSocket(handleWSMessage);

  const handleLogout = async () => {
    try {
      await authApi.logout();
      localStorage.removeItem('token');
      toast.success('已退出登录');
      window.location.href = '/login';
    } catch {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
  };

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/thing-models')) return 'thing-models';
    if (path.startsWith('/device-data')) return 'device-data';
    if (path.startsWith('/devices')) return 'devices';
    if (path.startsWith('/modules')) return 'modules';
    if (path.startsWith('/audit-logs')) return 'audit-logs';
    if (path.startsWith('/alert-rules')) return 'alert-rules';
    if (path.startsWith('/alert-logs')) return 'alert-logs';
    if (path.startsWith('/scheduled-tasks')) return 'scheduled-tasks';
    return 'dashboard';
  };

  const selectedKey = getSelectedKey();

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      {/* Sidebar */}
      <aside
        className={cn(
          'flex flex-col bg-sidebar border-r border-sidebar-border transition-all duration-300 shrink-0',
          collapsed ? 'w-16' : 'w-60'
        )}
      >
        <div className="h-16 flex items-center justify-center border-b border-sidebar-border shrink-0">
          {collapsed ? (
            <span className="text-lg font-bold text-primary">LF</span>
          ) : (
            <div className="flex items-center gap-2">
              <Leaf className="h-5 w-5 text-primary" />
              <span className="text-lg font-bold text-foreground">LinkFlow</span>
            </div>
          )}
        </div>

        <nav className="flex-1 py-4 px-2 space-y-1">
          {navItems.map((item) => {
            const isActive = selectedKey === item.key;
            return (
              <button
                key={item.key}
                onClick={() => navigate(item.path)}
                className={cn(
                  'w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors cursor-pointer',
                  isActive
                    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                    : 'text-muted-foreground hover:bg-sidebar-accent/50 hover:text-foreground'
                )}
              >
                <item.icon className="h-5 w-5 shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </button>
            );
          })}
        </nav>
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="h-16 flex items-center justify-between px-6 border-b border-border bg-card shrink-0">
          <Button variant="ghost" size="icon" onClick={() => setCollapsed(!collapsed)}>
            {collapsed ? <Menu className="h-5 w-5" /> : <X className="h-5 w-5" />}
          </Button>

          <div className="flex items-center gap-2">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" title="切换主题">
                  <Palette className="h-5 w-5" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => setTheme('green')} className={theme === 'green' ? 'text-primary' : ''}>
                  <span className="inline-block w-3 h-3 rounded-full bg-[#22c55e] mr-2" />
                  翠绿主题
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => setTheme('blue')} className={theme === 'blue' ? 'text-primary' : ''}>
                  <span className="inline-block w-3 h-3 rounded-full bg-[#38bdf8] mr-2" />
                  天蓝主题
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-accent transition-colors cursor-pointer">
                <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center">
                  <User className="h-4 w-4 text-primary" />
                </div>
                <span className="text-sm text-muted-foreground">{user?.role || '用户'}</span>
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleLogout} className="text-destructive focus:text-destructive">
                <LogOut className="h-4 w-4 mr-2" />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>

        <main className="flex-1 overflow-auto p-6">
          <Routes>
            <Route path="/" element={<WelcomePage user={user} stats={stats} />} />
            <Route path="/thing-models" element={<ThingModelList />} />
            <Route path="/thing-models/new" element={<ThingModelForm />} />
            <Route path="/thing-models/:id" element={<ThingModelForm />} />
            <Route path="/thing-models/:id/edit" element={<ThingModelForm />} />
            <Route path="/devices" element={<DeviceList />} />
            <Route path="/devices/new" element={<DeviceForm />} />
            <Route path="/devices/:id" element={<DeviceForm />} />
            <Route path="/devices/:id/edit" element={<DeviceForm />} />
            <Route path="/device-data" element={<DeviceData />} />
            <Route path="/modules" element={<ModuleList />} />
            <Route path="/audit-logs" element={<AuditLogList />} />
            <Route path="/alert-rules" element={<AlertRuleList />} />
            <Route path="/alert-rules/new" element={<AlertRuleForm />} />
            <Route path="/alert-rules/:id/edit" element={<AlertRuleForm />} />
            <Route path="/alert-logs" element={<AlertLogList />} />
            <Route path="/scheduled-tasks" element={<ScheduledTaskList />} />
            <Route path="/scheduled-tasks/new" element={<ScheduledTaskForm />} />
            <Route path="/scheduled-tasks/:id/edit" element={<ScheduledTaskForm />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}
