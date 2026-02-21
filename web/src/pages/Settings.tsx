import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { Settings2, Mic, ChevronDown, ChevronRight, Save, Info } from 'lucide-react';
import { settingsApi } from '../api';
import type { PlatformSettings } from '../api';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import { cn } from '@/lib/utils';

export default function Settings() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<PlatformSettings>({
    voice_mode: 'local',
    dify_api_url: '',
    dify_api_key: '',
  });
  const [guideOpen, setGuideOpen] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const res = await settingsApi.get();
      setSettings(res.data);
    } catch {
      toast.error('获取设置失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const res = await settingsApi.update({
        voice_mode: settings.voice_mode,
        dify_api_url: settings.dify_api_url,
        dify_api_key: settings.dify_api_key,
      });
      setSettings(res.data);
      toast.success('设置已保存');
    } catch (err: any) {
      const msg = err?.response?.data?.msg || '保存失败';
      toast.error(msg);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner size="lg" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center gap-3">
        <Settings2 className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold">系统设置</h2>
        <span className="text-sm text-muted-foreground">平台全局配置，修改后立即生效，无需重启</span>
      </div>

      {/* 语音控制设置卡片 */}
      <Card>
        <CardContent className="p-6 space-y-5">
          <div className="flex items-center gap-2 mb-1">
            <Mic className="h-4 w-4 text-primary" />
            <h3 className="font-medium">语音控制模式</h3>
          </div>

          {/* 模式选择 */}
          <div className="space-y-3">
            {/* 本地 NLP */}
            <label
              className={cn(
                'flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors',
                settings.voice_mode === 'local'
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-muted-foreground/50'
              )}
              onClick={() => setSettings((s) => ({ ...s, voice_mode: 'local' }))}
            >
              <div className={cn(
                'mt-0.5 h-4 w-4 rounded-full border-2 flex items-center justify-center shrink-0',
                settings.voice_mode === 'local' ? 'border-primary' : 'border-muted-foreground'
              )}>
                {settings.voice_mode === 'local' && (
                  <div className="h-2 w-2 rounded-full bg-primary" />
                )}
              </div>
              <div>
                <p className="font-medium text-sm">本地 NLP（无网络依赖）</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  使用内置关键词匹配引擎，支持开关、数值调节、枚举选择、服务调用，无需外部服务
                </p>
              </div>
            </label>

            {/* Dify AI */}
            <label
              className={cn(
                'flex items-start gap-3 p-4 rounded-lg border cursor-pointer transition-colors',
                settings.voice_mode === 'dify'
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-muted-foreground/50'
              )}
              onClick={() => setSettings((s) => ({ ...s, voice_mode: 'dify' }))}
            >
              <div className={cn(
                'mt-0.5 h-4 w-4 rounded-full border-2 flex items-center justify-center shrink-0',
                settings.voice_mode === 'dify' ? 'border-primary' : 'border-muted-foreground'
              )}>
                {settings.voice_mode === 'dify' && (
                  <div className="h-2 w-2 rounded-full bg-primary" />
                )}
              </div>
              <div>
                <p className="font-medium text-sm">Dify AI（自然语言理解）</p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  调用 Dify 工作流处理语音指令，支持更复杂的自然语言表达，需配置 Dify 连接参数
                </p>
              </div>
            </label>
          </div>

          {/* Dify 连接参数（仅 dify 模式下展示） */}
          {settings.voice_mode === 'dify' && (
            <div className="space-y-4 pt-2 border-t border-border">
              <div className="space-y-1.5">
                <Label htmlFor="dify-url">Dify 工作流地址</Label>
                <Input
                  id="dify-url"
                  placeholder="http://dify.example.com/v1"
                  value={settings.dify_api_url}
                  onChange={(e) => setSettings((s) => ({ ...s, dify_api_url: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">
                  Dify 服务的 API 基础地址，格式为 <code className="bg-muted px-1 rounded">http://&lt;host&gt;/v1</code>
                </p>
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="dify-key">Dify API Key</Label>
                <Input
                  id="dify-key"
                  type="password"
                  placeholder={settings.dify_api_key.startsWith('****') ? settings.dify_api_key : '请输入 API Key'}
                  value={settings.dify_api_key.startsWith('****') ? '' : settings.dify_api_key}
                  onChange={(e) => setSettings((s) => ({ ...s, dify_api_key: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">
                  Dify 工作流应用的 API Key，以 <code className="bg-muted px-1 rounded">app-</code> 开头。留空则保留已保存的值。
                </p>
              </div>
            </div>
          )}

          <Button onClick={handleSave} disabled={saving} className="w-full sm:w-auto">
            {saving ? <Spinner size="sm" className="mr-2" /> : <Save className="h-4 w-4 mr-2" />}
            保存设置
          </Button>
        </CardContent>
      </Card>

      {/* Dify 工作流配置说明（折叠卡片） */}
      <Card>
        <CardContent className="p-0">
          <button
            className="w-full flex items-center gap-3 px-6 py-4 text-left hover:bg-muted/40 transition-colors rounded-lg"
            onClick={() => setGuideOpen((v) => !v)}
          >
            <Info className="h-4 w-4 text-primary shrink-0" />
            <span className="font-medium text-sm flex-1">Dify 工作流配置指引</span>
            {guideOpen
              ? <ChevronDown className="h-4 w-4 text-muted-foreground" />
              : <ChevronRight className="h-4 w-4 text-muted-foreground" />}
          </button>

          {guideOpen && (
            <div className="px-6 pb-6 space-y-4 border-t border-border">
              <p className="text-sm text-muted-foreground pt-4">
                在 Dify 中创建一个 <strong>Workflow（工作流）</strong>应用，配置以下输入/输出变量，并在 System Prompt 中描述返回格式。
              </p>

              <div className="space-y-2">
                <p className="text-sm font-medium">输入变量</p>
                <div className="rounded-lg bg-muted p-3 font-mono text-xs space-y-1">
                  <div><span className="text-primary">device_context</span>  <span className="text-muted-foreground">— String，设备属性/服务上下文（平台自动生成）</span></div>
                  <div><span className="text-primary">user_input</span>  <span className="text-muted-foreground">— String，用户语音原文，如"把温度调到26度"</span></div>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium">输出变量</p>
                <div className="rounded-lg bg-muted p-3 font-mono text-xs">
                  <div><span className="text-primary">result</span>  <span className="text-muted-foreground">— String，JSON 格式的指令对象</span></div>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium">result 输出格式（JSON 字符串）</p>
                <pre className="rounded-lg bg-muted p-3 text-xs overflow-x-auto">{`{
  "action": "set_property" | "invoke_service" | "query_status" | "unknown",
  "property_id": "属性ID",    // action=set_property/query_status 时必填
  "value": <值>,               // action=set_property 时必填
  "service_id": "服务ID",     // action=invoke_service 时必填
  "params": {},                // 服务调用参数（可为空对象）
  "message": "人类可读描述"   // 始终填写
}`}</pre>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium">推荐 System Prompt</p>
                <pre className="rounded-lg bg-muted p-3 text-xs overflow-x-auto whitespace-pre-wrap">{`你是 IoT 设备控制助手。根据 device_context 和 user_input，输出且仅输出一个 JSON 对象。

规则：
- 只能操作 device_context.properties 和 device_context.services 中列出的项
- 数值需在 min/max 范围内；枚举需匹配 enum_values.value 之一
- bool：打开/开启/启动=true，关闭/关掉/停止=false
- 无法理解时 action="unknown"，message 说明原因`}</pre>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
