import { useEffect, useState, useCallback } from 'react';
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { toast } from 'sonner';
import {
  LayoutDashboard, Box, Cpu, Activity, Blocks, ScrollText, LogOut, Menu, X, ChevronDown, ChevronRight, Leaf, User, Bell, AlertTriangle, Clock, Palette, Terminal, KeyRound, Upload, Radio, Shield, ClipboardList,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { authApi, alertLogApi, statsApi } from '../api';
import type { StatsOverview } from '../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
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
import DeviceDebug from './DeviceDebug';
import FirmwareList from './FirmwareList';
import OTATaskList from './OTATaskList';
import ScheduledTaskLogList from './ScheduledTaskLogList';

interface UserInfo {
  user_id: string;
  role: string;
}

interface NavItem {
  key: string; label: string; icon: any; path: string; badge?: boolean;
}
interface NavGroup {
  group: string; icon: any; items: NavItem[];
}

const navGroups: (NavItem | NavGroup)[] = [
  { key: 'dashboard', label: '仪表盘', icon: LayoutDashboard, path: '/' },
  {
    group: '设备', icon: Cpu, items: [
      { key: 'thing-models', label: '物模型', icon: Box, path: '/thing-models' },
      { key: 'devices', label: '设备列表', icon: Cpu, path: '/devices' },
      { key: 'device-data', label: '设备数据', icon: Activity, path: '/device-data' },
      { key: 'device-debug', label: '在线调试', icon: Terminal, path: '/device-debug' },
    ],
  },
  {
    group: '运维', icon: Shield, items: [
      { key: 'alert-rules', label: '告警规则', icon: Bell, path: '/alert-rules' },
      { key: 'alert-logs', label: '告警历史', icon: AlertTriangle, path: '/alert-logs', badge: true },
      { key: 'scheduled-tasks', label: '定时任务', icon: Clock, path: '/scheduled-tasks' },
      { key: 'scheduled-task-logs', label: '执行历史', icon: ClipboardList, path: '/scheduled-task-logs' },
    ],
  },
  {
    group: '升级', icon: Radio, items: [
      { key: 'firmwares', label: '固件管理', icon: Upload, path: '/firmwares' },
      { key: 'ota-tasks', label: 'OTA升级', icon: Radio, path: '/ota-tasks' },
    ],
  },
  {
    group: '系统', icon: Blocks, items: [
      { key: 'modules', label: '功能模块', icon: Blocks, path: '/modules' },
      { key: 'audit-logs', label: '审计日志', icon: ScrollText, path: '/audit-logs' },
    ],
  },
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
    { label: '今日告警', value: stats?.today_alerts ?? '--', color: 'text-yellow-400' },
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

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-12 w-full max-w-3xl">
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
  const [alertUnreadCount, setAlertUnreadCount] = useState(0);
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({ '设备': true, '运维': true, '升级': true, '系统': true });

  const toggleGroup = (group: string) => setExpandedGroups((prev) => ({ ...prev, [group]: !prev[group] }));

  // 修改密码 Dialog 状态
  const [changePwOpen, setChangePwOpen] = useState(false);
  const [oldPw, setOldPw] = useState('');
  const [newPw, setNewPw] = useState('');
  const [confirmPw, setConfirmPw] = useState('');
  const [changePwLoading, setChangePwLoading] = useState(false);

  useEffect(() => {
    fetchUser();
    fetchStats();
    fetchUnreadCount();
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

  const fetchUnreadCount = async () => {
    try {
      const { data } = await alertLogApi.unreadCount();
      setAlertUnreadCount(data.count ?? 0);
    } catch { /* silent */ }
  };

  // 子页面确认告警后调用，减少未读数
  const decrementAlertCount = useCallback((n: number) => {
    setAlertUnreadCount((prev) => Math.max(0, prev - n));
  }, []);

  // WebSocket 消息处理
  const handleWSMessage = useCallback((msg: WSMessage) => {
    // 统计更新
    if (msg.type === 'stats') {
      setStats((prev) => prev ? { ...prev, online_devices: msg.data.online_devices } : prev);
    }

    // 告警 toast + 未读数+1
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
      setAlertUnreadCount((prev) => prev + 1);
      setStats((prev) => prev ? { ...prev, today_alerts: (prev.today_alerts ?? 0) + 1 } : prev);
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

  const handleChangePassword = async () => {
    if (!oldPw) { toast.error('请输入当前密码'); return; }
    if (newPw.length < 6) { toast.error('新密码至少6位'); return; }
    if (newPw !== confirmPw) { toast.error('两次新密码不一致'); return; }

    setChangePwLoading(true);
    try {
      await authApi.changePassword({ old_password: oldPw, new_password: newPw });
      toast.success('密码修改成功');
      setChangePwOpen(false);
      setOldPw(''); setNewPw(''); setConfirmPw('');
    } catch (err: any) {
      const msg = err?.response?.data?.msg || '修改失败';
      toast.error(msg);
    } finally {
      setChangePwLoading(false);
    }
  };

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path.startsWith('/thing-models')) return 'thing-models';
    if (path.startsWith('/device-debug')) return 'device-debug';
    if (path.startsWith('/device-data')) return 'device-data';
    if (path.startsWith('/devices')) return 'devices';
    if (path.startsWith('/modules')) return 'modules';
    if (path.startsWith('/audit-logs')) return 'audit-logs';
    if (path.startsWith('/alert-rules')) return 'alert-rules';
    if (path.startsWith('/alert-logs')) return 'alert-logs';
    if (path.startsWith('/scheduled-task-logs')) return 'scheduled-task-logs';
    if (path.startsWith('/scheduled-tasks')) return 'scheduled-tasks';
    if (path.startsWith('/firmwares')) return 'firmwares';
    if (path.startsWith('/ota-tasks')) return 'ota-tasks';
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

        <nav className="flex-1 py-4 px-2 space-y-1 overflow-y-auto">
          {navGroups.map((entry, idx) => {
            if ('key' in entry) {
              const isActive = selectedKey === entry.key;
              return (
                <button key={entry.key} onClick={() => navigate(entry.path)}
                  className={cn('w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors cursor-pointer',
                    isActive ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-muted-foreground hover:bg-sidebar-accent/50 hover:text-foreground'
                  )}>
                  <entry.icon className="h-5 w-5 shrink-0" />
                  {!collapsed && <span className="flex-1 text-left">{entry.label}</span>}
                </button>
              );
            }
            const group = entry as NavGroup;
            const expanded = expandedGroups[group.group];
            const hasActiveSub = group.items.some((i) => selectedKey === i.key);
            return (
              <div key={group.group} className={idx > 0 ? 'pt-2' : ''}>
                {!collapsed && (
                  <button onClick={() => toggleGroup(group.group)}
                    className={cn('w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer',
                      hasActiveSub ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
                    )}>
                    {expanded ? <ChevronDown className="h-4 w-4 shrink-0" /> : <ChevronRight className="h-4 w-4 shrink-0" />}
                    <span className="flex-1 text-left">{group.group}</span>
                  </button>
                )}
                {(collapsed || expanded) && group.items.map((item) => {
                  const isActive = selectedKey === item.key;
                  const showBadge = item.badge && alertUnreadCount > 0;
                  return (
                    <button key={item.key} onClick={() => navigate(item.path)}
                      className={cn('w-full flex items-center gap-3 rounded-lg text-sm font-medium transition-colors cursor-pointer',
                        collapsed ? 'px-3 py-2.5' : 'pl-7 pr-3 py-2',
                        isActive ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-muted-foreground hover:bg-sidebar-accent/50 hover:text-foreground'
                      )}>
                      <item.icon className={cn('shrink-0', collapsed ? 'h-5 w-5' : 'h-4 w-4')} />
                      {!collapsed && (
                        <>
                          <span className="flex-1 text-left">{item.label}</span>
                          {showBadge && (
                            <span className="text-xs bg-destructive text-destructive-foreground rounded-full px-1.5 py-0.5 min-w-[18px] text-center leading-none">
                              {alertUnreadCount > 99 ? '99+' : alertUnreadCount}
                            </span>
                          )}
                        </>
                      )}
                    </button>
                  );
                })}
              </div>
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
              <DropdownMenuItem onClick={() => setChangePwOpen(true)}>
                <KeyRound className="h-4 w-4 mr-2" />
                修改密码
              </DropdownMenuItem>
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
            <Route path="/device-debug" element={<DeviceDebug />} />
            <Route path="/modules" element={<ModuleList />} />
            <Route path="/audit-logs" element={<AuditLogList />} />
            <Route path="/alert-rules" element={<AlertRuleList />} />
            <Route path="/alert-rules/new" element={<AlertRuleForm />} />
            <Route path="/alert-rules/:id/edit" element={<AlertRuleForm />} />
            <Route path="/alert-logs" element={<AlertLogList onAcknowledge={decrementAlertCount} />} />
            <Route path="/scheduled-tasks" element={<ScheduledTaskList />} />
            <Route path="/scheduled-tasks/new" element={<ScheduledTaskForm />} />
            <Route path="/scheduled-tasks/:id/edit" element={<ScheduledTaskForm />} />
            <Route path="/scheduled-task-logs" element={<ScheduledTaskLogList />} />
            <Route path="/firmwares" element={<FirmwareList />} />
            <Route path="/ota-tasks" element={<OTATaskList />} />
          </Routes>
        </main>
      </div>

      {/* 修改密码 Dialog */}
      <Dialog open={changePwOpen} onOpenChange={(open) => {
        setChangePwOpen(open);
        if (!open) { setOldPw(''); setNewPw(''); setConfirmPw(''); }
      }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>修改密码</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="space-y-2">
              <Label>当前密码</Label>
              <Input
                type="password"
                value={oldPw}
                onChange={(e) => setOldPw(e.target.value)}
                placeholder="请输入当前密码"
              />
            </div>
            <div className="space-y-2">
              <Label>新密码</Label>
              <Input
                type="password"
                value={newPw}
                onChange={(e) => setNewPw(e.target.value)}
                placeholder="至少6位"
              />
            </div>
            <div className="space-y-2">
              <Label>确认新密码</Label>
              <Input
                type="password"
                value={confirmPw}
                onChange={(e) => setConfirmPw(e.target.value)}
                placeholder="再次输入新密码"
                onKeyDown={(e) => e.key === 'Enter' && handleChangePassword()}
              />
            </div>
            <div className="flex gap-3 pt-2">
              <Button onClick={handleChangePassword} disabled={changePwLoading} className="flex-1">
                {changePwLoading ? '修改中...' : '确认修改'}
              </Button>
              <Button variant="outline" onClick={() => setChangePwOpen(false)} className="flex-1">
                取消
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
