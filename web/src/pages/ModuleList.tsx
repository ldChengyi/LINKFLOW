import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Blocks, Mic, Eye, Clock, Cpu, Upload, Radio, Download, CheckCircle2, ArrowRight } from 'lucide-react';
import { moduleApi } from '../api';
import type { Module } from '../api';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose,
} from '@/components/ui/dialog';

const moduleIcons: Record<string, typeof Mic> = {
  voice: Mic,
  scheduler: Clock,
};

export default function ModuleList() {
  const [modules, setModules] = useState<Module[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Module | null>(null);
  const [otaDetailOpen, setOtaDetailOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    fetchModules();
  }, []);

  const fetchModules = async () => {
    try {
      const res = await moduleApi.list();
      setModules(res.data || []);
    } catch {
      toast.error('获取模块列表失败');
    } finally {
      setLoading(false);
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
    <>
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Blocks className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold">功能模块</h2>
        <span className="text-sm text-muted-foreground">平台内置的可扩展模块，可在物模型中启用</span>
      </div>

      {modules.length === 0 ? (
        <Card>
          <CardContent className="py-16 text-center text-muted-foreground">
            暂无可用模块
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {modules.map((mod) => {
            const Icon = moduleIcons[mod.id] || Blocks;
            return (
              <Card key={mod.id} className="flex flex-col">
                <CardContent className="flex flex-col flex-1 pt-6">
                  <div className="flex items-start gap-3 mb-3">
                    <div className="p-2.5 rounded-lg bg-primary/10 shrink-0">
                      <Icon className="h-5 w-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm">{mod.name}</span>
                        <Badge variant="outline" className="font-mono text-xs">{mod.id}</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1 line-clamp-2">{mod.description}</p>
                    </div>
                  </div>
                  <div className="flex items-center justify-between mt-auto pt-3 border-t">
                    <span className="text-xs text-muted-foreground">创建于 {new Date(mod.created_at).toLocaleDateString()}</span>
                    <Button variant="outline" size="sm" onClick={() => setSelected(mod)}>
                      <Eye className="h-4 w-4 mr-1" />
                      查看详情
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* 平台内置功能 */}
      <div className="space-y-4 pt-2">
        <div className="flex items-center gap-3 border-t pt-6">
          <Cpu className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">平台内置功能</h2>
          <span className="text-sm text-muted-foreground">无需绑定物模型，平台直接提供</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <Card className="flex flex-col">
            <CardContent className="flex flex-col flex-1 pt-6">
              <div className="flex items-start gap-3 mb-3">
                <div className="p-2.5 rounded-lg bg-primary/10 shrink-0">
                  <Cpu className="h-5 w-5 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm">OTA 固件升级</span>
                    <Badge variant="outline" className="font-mono text-xs">ota</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                    通过 MQTT 向设备远程推送固件升级指令，设备经 HTTP 下载固件后自行烧录并上报进度。
                  </p>
                </div>
              </div>
              <div className="flex items-center justify-between mt-auto pt-3 border-t">
                <Badge variant="secondary" className="text-xs">平台内置</Badge>
                <Button variant="outline" size="sm" onClick={() => setOtaDetailOpen(true)}>
                  <Eye className="h-4 w-4 mr-1" />
                  查看详情
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>

    <Dialog open={!!selected} onOpenChange={(open) => !open && setSelected(null)}>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] overflow-y-auto">
        {selected && <ModuleDetail mod={selected} />}
        <DialogClose asChild>
          <Button variant="outline" className="w-full">关闭</Button>
        </DialogClose>
      </DialogContent>
    </Dialog>

    <Dialog open={otaDetailOpen} onOpenChange={setOtaDetailOpen}>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] overflow-y-auto">
        <OTADetail navigate={navigate} />
        <DialogClose asChild>
          <Button variant="outline" className="w-full">关闭</Button>
        </DialogClose>
      </DialogContent>
    </Dialog>
    </>
  );
}

function ModuleDetail({ mod }: { mod: Module }) {
  const Icon = moduleIcons[mod.id] || Blocks;
  const schema = mod.config_schema || {};
  const configKeys = Object.keys(schema);

  return (
    <>
      <DialogHeader>
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-lg bg-primary/10">
            <Icon className="h-5 w-5 text-primary" />
          </div>
          <div>
            <DialogTitle className="flex items-center gap-2">
              {mod.name}
              <Badge variant="outline" className="font-mono text-xs">{mod.id}</Badge>
            </DialogTitle>
            <p className="text-sm text-muted-foreground mt-0.5">{mod.description}</p>
          </div>
        </div>
      </DialogHeader>

      <div className="space-y-5 mt-2">
        {/* 使用方式 */}
        <div className="space-y-2">
          <p className="text-sm font-medium">使用方式</p>
          <div className="text-sm text-muted-foreground space-y-1 bg-muted/50 rounded-lg p-3">
            <p>1. 在物模型编辑页的「模块」标签页中启用此模块</p>
            <p>2. 选择需要暴露给模块的属性和服务</p>
            <p>3. 绑定该物模型的设备将自动获得模块能力</p>
          </div>
        </div>

        {/* 配置项 */}
        {configKeys.length > 0 && (
          <div className="space-y-2">
            <p className="text-sm font-medium">配置项</p>
            <div className="rounded-lg border overflow-hidden">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-muted/50 text-muted-foreground">
                    <th className="text-left px-3 py-2 font-medium">字段</th>
                    <th className="text-left px-3 py-2 font-medium">类型</th>
                    <th className="text-left px-3 py-2 font-medium">说明</th>
                  </tr>
                </thead>
                <tbody>
                  {configKeys.map((key) => {
                    const field = schema[key] as Record<string, unknown>;
                    return (
                      <tr key={key} className="border-t">
                        <td className="px-3 py-2 font-mono text-xs">{key}</td>
                        <td className="px-3 py-2">
                          <Badge variant="secondary" className="text-xs">
                            {String(field.type || 'any')}
                          </Badge>
                        </td>
                        <td className="px-3 py-2 text-muted-foreground text-xs">
                          {String(field.description || '-')}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* 语音模块 MQTT Topic */}
        {mod.id === 'voice' && (
          <div className="space-y-4">
            <div className="space-y-2">
              <p className="text-sm font-medium">MQTT Topic</p>
              <div className="space-y-1.5">
                <TopicRow direction="up" topic="devices/{device_id}/voice/up" desc="设备上报语音文本" />
                <TopicRow direction="down" topic="devices/{device_id}/voice/down" desc="平台返回执行结果" />
              </div>
              <div className="space-y-1.5 mt-3">
                <p className="text-xs text-muted-foreground">上报 Payload</p>
                <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">
{JSON.stringify({ text: "打开灯" }, null, 2)}
                </pre>
                <p className="text-xs text-muted-foreground mt-2">返回 Payload（成功）</p>
                <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">
{JSON.stringify({ success: true, message: "指令已执行", action: "设置 台灯.switch = true" }, null, 2)}
                </pre>
                <p className="text-xs text-muted-foreground mt-2">返回 Payload（失败）</p>
                <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">
{JSON.stringify({ success: false, message: "未匹配到目标设备" }, null, 2)}
                </pre>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm font-medium">支持的指令类型</p>
              <div className="rounded-lg border overflow-hidden">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="bg-muted/50 text-muted-foreground">
                      <th className="text-left px-3 py-2 font-medium">意图</th>
                      <th className="text-left px-3 py-2 font-medium">关键词示例</th>
                      <th className="text-left px-3 py-2 font-medium">适用类型</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">开关控制</td>
                      <td className="px-3 py-2 font-mono">打开 / 关闭 / 开启 / 关掉</td>
                      <td className="px-3 py-2 text-muted-foreground">bool</td>
                    </tr>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">设定值</td>
                      <td className="px-3 py-2 font-mono">调到X / 设为X / 设置为X</td>
                      <td className="px-3 py-2 text-muted-foreground">int / float</td>
                    </tr>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">相对调节</td>
                      <td className="px-3 py-2 font-mono">调高 / 调低 / 增大 / 减小</td>
                      <td className="px-3 py-2 text-muted-foreground">int / float（按 step）</td>
                    </tr>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">枚举选择</td>
                      <td className="px-3 py-2 font-mono">说出枚举项的 label</td>
                      <td className="px-3 py-2 text-muted-foreground">enum</td>
                    </tr>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">服务调用</td>
                      <td className="px-3 py-2 font-mono">执行 / 重启 / 调用</td>
                      <td className="px-3 py-2 text-muted-foreground">service</td>
                    </tr>
                    <tr className="border-t">
                      <td className="px-3 py-2 font-medium">状态查询</td>
                      <td className="px-3 py-2 font-mono">X是多少 / 当前X / 查询X</td>
                      <td className="px-3 py-2 text-muted-foreground">所有类型（只读）</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p className="text-xs text-muted-foreground">平台按设备名或属性名在用户所有设备中匹配目标，结果有歧义时不执行。</p>
            </div>
          </div>
        )}

        {/* 定时任务模块 MQTT Topic */}
        {mod.id === 'scheduler' && (
          <div className="space-y-2">
            <p className="text-sm font-medium">MQTT Topic（设备需订阅）</p>
            <div className="space-y-1.5">
              <TopicRow direction="down" topic="devices/{device_id}/telemetry/down" desc="属性设置下发" />
              <TopicRow direction="down" topic="devices/{device_id}/service/invoke" desc="服务调用下发" />
            </div>
            <div className="space-y-3 mt-3">
              <div>
                <p className="text-xs text-muted-foreground mb-1.5">属性设置 Payload（property_set）</p>
                <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">
{JSON.stringify({ switch: true, brightness: 80 }, null, 2)}
                </pre>
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-1.5">服务调用 Payload（service_invoke）</p>
                <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">
{JSON.stringify({ id: "sched_a1b2c3d4_1700000000", service: "reboot", params: {} }, null, 2)}
                </pre>
              </div>
            </div>
            <div className="bg-muted/50 rounded-lg p-3 text-xs text-muted-foreground space-y-1 mt-1">
              <p>调试方法：在 MQTTX 中使用设备凭据连接后，订阅以下两个 topic：</p>
              <code className="block font-mono mt-1">{'devices/{device_id}/telemetry/down'}</code>
              <code className="block font-mono">{'devices/{device_id}/service/invoke'}</code>
              <p className="mt-1">定时任务触发时即可在 MQTTX 中看到下发的消息。</p>
            </div>
          </div>
        )}
      </div>
    </>
  );
}

function OTADetail({ navigate }: { navigate: (path: string) => void }) {
  return (
    <>
      <DialogHeader>
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-lg bg-primary/10">
            <Cpu className="h-5 w-5 text-primary" />
          </div>
          <div>
            <DialogTitle className="flex items-center gap-2">
              OTA 固件升级
              <Badge variant="outline" className="font-mono text-xs">ota</Badge>
            </DialogTitle>
            <p className="text-sm text-muted-foreground mt-0.5">
              通过 MQTT 向设备远程推送固件升级指令，设备经 HTTP 下载固件后自行烧录并上报进度，平台实时跟踪升级状态。
            </p>
          </div>
        </div>
      </DialogHeader>

      <div className="space-y-5 mt-2">
        {/* 升级流程 */}
        <div className="space-y-2">
          <p className="text-sm font-medium">升级流程</p>
          <div className="flex items-center gap-1 flex-wrap">
            {[
              { icon: Upload, label: '上传固件' },
              { icon: CheckCircle2, label: '创建任务' },
              { icon: Radio, label: 'MQTT 推送' },
              { icon: Download, label: '设备下载' },
              { icon: Cpu, label: '烧录固件' },
              { icon: CheckCircle2, label: '进度上报' },
            ].map((step, i, arr) => (
              <div key={i} className="flex items-center gap-1">
                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-muted/60 text-xs">
                  <step.icon className="h-3.5 w-3.5 text-primary" />
                  <span>{step.label}</span>
                </div>
                {i < arr.length - 1 && <ArrowRight className="h-3.5 w-3.5 text-muted-foreground shrink-0" />}
              </div>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">
            设备上线时若存在 pending 任务，平台将自动推送升级指令；升级完成后设备固件版本字段自动更新。
          </p>
        </div>

        {/* MQTT Topic */}
        <div className="space-y-2">
          <p className="text-sm font-medium">MQTT Topic</p>
          <div className="space-y-1.5">
            <TopicRow direction="down" topic="devices/{device_id}/ota/down" desc="平台下发升级命令" />
            <TopicRow direction="up" topic="devices/{device_id}/ota/up" desc="设备上报升级进度" />
          </div>
          <div className="space-y-3 mt-3">
            <div>
              <p className="text-xs text-muted-foreground mb-1.5">下发 Payload（ota/down）</p>
              <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">{JSON.stringify({
                task_id: "uuid",
                firmware_id: "uuid",
                version: "1.2.0",
                download_url: "http://host/api/firmwares/{id}/download",
                checksum: "sha256hex",
                size: 102400
              }, null, 2)}</pre>
            </div>
            <div>
              <p className="text-xs text-muted-foreground mb-1.5">上报 Payload（ota/up）</p>
              <pre className="bg-muted rounded-lg p-3 text-xs font-mono overflow-x-auto">{JSON.stringify({
                task_id: "uuid",
                status: "downloading",
                progress: 45,
                message: "Downloading firmware..."
              }, null, 2)}</pre>
              <p className="text-xs text-muted-foreground mt-1.5">
                status 可选值：<code className="font-mono bg-muted px-1 rounded">downloading</code>{' '}
                <code className="font-mono bg-muted px-1 rounded">installing</code>{' '}
                <code className="font-mono bg-muted px-1 rounded">completed</code>{' '}
                <code className="font-mono bg-muted px-1 rounded">failed</code>
              </p>
            </div>
          </div>
        </div>

        {/* 固件下载认证 */}
        <div className="space-y-2">
          <p className="text-sm font-medium">固件下载认证</p>
          <div className="text-sm text-muted-foreground space-y-1 bg-muted/50 rounded-lg p-3">
            <p>固件文件通过 HTTP Basic Auth 保护，设备需使用以下凭据下载固件：</p>
            <div className="flex items-center gap-6 mt-2 flex-wrap">
              <div className="flex items-center gap-2">
                <span className="text-xs font-medium text-foreground">用户名</span>
                <code className="bg-muted px-2 py-0.5 rounded text-xs font-mono">device_id</code>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-medium text-foreground">密码</span>
                <code className="bg-muted px-2 py-0.5 rounded text-xs font-mono">device_secret</code>
              </div>
            </div>
          </div>
        </div>

        {/* 快速导航 */}
        <div className="flex items-center gap-3 pt-1 border-t">
          <span className="text-xs text-muted-foreground">快速跳转</span>
          <Button variant="outline" size="sm" onClick={() => navigate('/firmwares')}>
            <Upload className="h-3.5 w-3.5 mr-1.5" />
            固件管理
          </Button>
          <Button variant="outline" size="sm" onClick={() => navigate('/ota-tasks')}>
            <Cpu className="h-3.5 w-3.5 mr-1.5" />
            OTA 任务
          </Button>
        </div>
      </div>
    </>
  );
}

function TopicRow({ direction, topic, desc }: { direction: 'up' | 'down'; topic: string; desc: string }) {
  return (
    <div className="flex items-center gap-3 py-1.5 px-3 rounded-lg bg-muted/50">
      <span className={`text-xs font-bold px-2 py-0.5 rounded ${direction === 'up' ? 'bg-violet-500/15 text-violet-400' : 'bg-cyan-500/15 text-cyan-400'}`}>
        {direction === 'up' ? 'PUB' : 'SUB'}
      </span>
      <code className="text-xs font-mono flex-1">{topic}</code>
      <span className="text-xs text-muted-foreground shrink-0">{desc}</span>
    </div>
  );
}
