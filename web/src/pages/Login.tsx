import { useState } from 'react';
import { authApi } from '../api';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Loader2, Mail, Lock, Leaf, CheckCircle2, XCircle } from 'lucide-react';
import { useTheme } from '../hooks/useTheme';

type FormMsg = { type: 'success' | 'error'; text: string } | null;

export default function Login() {
  useTheme();
  const [loading, setLoading] = useState(false);

  const [loginEmail, setLoginEmail] = useState('');
  const [loginPassword, setLoginPassword] = useState('');

  const [regEmail, setRegEmail] = useState('');
  const [regPassword, setRegPassword] = useState('');
  const [regConfirm, setRegConfirm] = useState('');

  const [formMsg, setFormMsg] = useState<FormMsg>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, boolean>>({});

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormMsg(null);
    const errors: Record<string, boolean> = {};
    if (!loginEmail) errors['login-email'] = true;
    if (!loginPassword) errors['login-password'] = true;
    if (Object.keys(errors).length) {
      setFieldErrors(errors);
      setFormMsg({ type: 'error', text: '请填写邮箱和密码' });
      return;
    }
    setFieldErrors({});
    setLoading(true);
    try {
      const { data } = await authApi.login({ email: loginEmail, password: loginPassword });
      localStorage.setItem('token', data.token);
      setFormMsg({ type: 'success', text: '登录成功，正在跳转...' });
      setTimeout(() => { window.location.href = '/'; }, 800);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { msg?: string } } };
      const msg = err.response?.data?.msg || '登录失败，请检查邮箱和密码';
      setFormMsg({ type: 'error', text: msg });
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormMsg(null);
    const errors: Record<string, boolean> = {};
    if (!regEmail) errors['reg-email'] = true;
    if (!regPassword) errors['reg-password'] = true;
    if (!regConfirm) errors['reg-confirm'] = true;
    if (Object.keys(errors).length) {
      setFieldErrors(errors);
      setFormMsg({ type: 'error', text: '请填写所有字段' });
      return;
    }
    if (regPassword.length < 6) {
      setFieldErrors({ 'reg-password': true });
      setFormMsg({ type: 'error', text: '密码至少6位' });
      return;
    }
    if (regPassword !== regConfirm) {
      setFieldErrors({ 'reg-confirm': true });
      setFormMsg({ type: 'error', text: '两次密码输入不一致' });
      return;
    }
    setFieldErrors({});
    setLoading(true);
    try {
      const { data } = await authApi.register({ email: regEmail, password: regPassword });
      localStorage.setItem('token', data.token);
      setFormMsg({ type: 'success', text: '注册成功，正在跳转...' });
      setTimeout(() => { window.location.href = '/'; }, 800);
    } catch (error: unknown) {
      const err = error as { response?: { data?: { msg?: string } } };
      const msg = err.response?.data?.msg || '注册失败，请稍后重试';
      setFormMsg({ type: 'error', text: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden" style={{ background: `linear-gradient(to bottom right, var(--color-login-from), var(--color-login-via), var(--color-login-to))` }}>
      {/* Background glow */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-primary/5 rounded-full blur-3xl" />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-primary/3 rounded-full blur-3xl" />
      </div>

      <Card className="w-[420px] relative z-10 bg-card/80 backdrop-blur-xl border-border/50 shadow-2xl shadow-black/40">
        <CardHeader className="text-center pb-2">
          <div className="flex items-center justify-center gap-2 mb-2">
            <div className="p-2 rounded-lg bg-primary/10">
              <Leaf className="h-6 w-6 text-primary" />
            </div>
          </div>
          <CardTitle className="text-2xl font-bold tracking-tight">LinkFlow</CardTitle>
          <CardDescription>物联网云平台</CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="login" className="w-full" onValueChange={() => { setFormMsg(null); setFieldErrors({}); }}>
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="login">登录</TabsTrigger>
              <TabsTrigger value="register">注册</TabsTrigger>
            </TabsList>

            <TabsContent value="login">
              <form onSubmit={handleLogin} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="login-email">邮箱</Label>
                  <div className="relative">
                    <Mail className={cn("absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4", fieldErrors['login-email'] ? 'text-destructive' : 'text-muted-foreground')} />
                    <Input
                      id="login-email"
                      type="email"
                      placeholder="请输入邮箱"
                      className={cn("pl-10", fieldErrors['login-email'] && 'border-destructive')}
                      value={loginEmail}
                      onChange={(e) => { setLoginEmail(e.target.value); setFieldErrors(p => ({ ...p, 'login-email': false })); }}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="login-password">密码</Label>
                  <div className="relative">
                    <Lock className={cn("absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4", fieldErrors['login-password'] ? 'text-destructive' : 'text-muted-foreground')} />
                    <Input
                      id="login-password"
                      type="password"
                      placeholder="请输入密码"
                      className={cn("pl-10", fieldErrors['login-password'] && 'border-destructive')}
                      value={loginPassword}
                      onChange={(e) => { setLoginPassword(e.target.value); setFieldErrors(p => ({ ...p, 'login-password': false })); }}
                    />
                  </div>
                </div>
                {formMsg && (
                  <div className={cn("flex items-center gap-2 text-sm px-3 py-2 rounded-lg", formMsg.type === 'error' ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary')}>
                    {formMsg.type === 'error' ? <XCircle className="h-4 w-4 shrink-0" /> : <CheckCircle2 className="h-4 w-4 shrink-0" />}
                    {formMsg.text}
                  </div>
                )}
                <Button type="submit" className="w-full" disabled={loading}>
                  {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                  登录
                </Button>
              </form>
            </TabsContent>

            <TabsContent value="register">
              <form onSubmit={handleRegister} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="reg-email">邮箱</Label>
                  <div className="relative">
                    <Mail className={cn("absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4", fieldErrors['reg-email'] ? 'text-destructive' : 'text-muted-foreground')} />
                    <Input
                      id="reg-email"
                      type="email"
                      placeholder="请输入邮箱"
                      className={cn("pl-10", fieldErrors['reg-email'] && 'border-destructive')}
                      value={regEmail}
                      onChange={(e) => { setRegEmail(e.target.value); setFieldErrors(p => ({ ...p, 'reg-email': false })); }}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="reg-password">密码</Label>
                  <div className="relative">
                    <Lock className={cn("absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4", fieldErrors['reg-password'] ? 'text-destructive' : 'text-muted-foreground')} />
                    <Input
                      id="reg-password"
                      type="password"
                      placeholder="密码至少6位"
                      className={cn("pl-10", fieldErrors['reg-password'] && 'border-destructive')}
                      value={regPassword}
                      onChange={(e) => { setRegPassword(e.target.value); setFieldErrors(p => ({ ...p, 'reg-password': false })); }}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="reg-confirm">确认密码</Label>
                  <div className="relative">
                    <Lock className={cn("absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4", fieldErrors['reg-confirm'] ? 'text-destructive' : 'text-muted-foreground')} />
                    <Input
                      id="reg-confirm"
                      type="password"
                      placeholder="再次输入密码"
                      className={cn("pl-10", fieldErrors['reg-confirm'] && 'border-destructive')}
                      value={regConfirm}
                      onChange={(e) => { setRegConfirm(e.target.value); setFieldErrors(p => ({ ...p, 'reg-confirm': false })); }}
                    />
                  </div>
                </div>
                {formMsg && (
                  <div className={cn("flex items-center gap-2 text-sm px-3 py-2 rounded-lg", formMsg.type === 'error' ? 'bg-destructive/10 text-destructive' : 'bg-primary/10 text-primary')}>
                    {formMsg.type === 'error' ? <XCircle className="h-4 w-4 shrink-0" /> : <CheckCircle2 className="h-4 w-4 shrink-0" />}
                    {formMsg.text}
                  </div>
                )}
                <Button type="submit" className="w-full" disabled={loading}>
                  {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                  注册
                </Button>
              </form>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
