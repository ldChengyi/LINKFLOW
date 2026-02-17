import { useEffect, useState } from 'react';
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { toast } from 'sonner';
import {
  LayoutDashboard, Box, Cpu, Activity, Blocks, LogOut, Menu, X, ChevronDown, Leaf, User,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { authApi, statsApi } from '../api';
import type { StatsOverview } from '../api';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import ThingModelList from './ThingModelList';
import ThingModelForm from './ThingModelForm';
import DeviceList from './DeviceList';
import DeviceForm from './DeviceForm';
import DeviceData from './DeviceData';
import ModuleList from './ModuleList';

interface UserInfo {
  user_id: string;
  role: string;
}

const navItems = [
  { key: 'dashboard', label: '仪表盘', icon: LayoutDashboard, path: '/' },
  { key: 'thing-models', label: '物模型', icon: Box, path: '/thing-models' },
  { key: 'devices', label: '设备管理', icon: Cpu, path: '/devices' },
  { key: 'device-data', label: '设备数据', icon: Activity, path: '/device-data' },
  { key: 'modules', label: '功能模块', icon: Blocks, path: '/modules' },
];

function WelcomePage({ user }: { user: UserInfo | null }) {
  const [stats, setStats] = useState<StatsOverview | null>(null);

  useEffect(() => {
    statsApi.overview().then(({ data }) => setStats(data)).catch(() => {});
  }, []);

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

  useEffect(() => {
    fetchUser();
  }, []);

  const fetchUser = async () => {
    try {
      const { data } = await authApi.me();
      setUser(data);
    } catch {
      // handled by interceptor
    }
  };

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
        </header>

        <main className="flex-1 overflow-auto p-6">
          <Routes>
            <Route path="/" element={<WelcomePage user={user} />} />
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
          </Routes>
        </main>
      </div>
    </div>
  );
}
