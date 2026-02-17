import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ArrowLeft, Copy, Check } from 'lucide-react';
import { deviceApi, thingModelApi } from '../api';
import type { ThingModel } from '../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Spinner } from '@/components/ui/spinner';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog';

export default function DeviceForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = !!id && id !== 'new';

  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [thingModels, setThingModels] = useState<ThingModel[]>([]);
  const [copied, setCopied] = useState(false);

  // Form fields
  const [name, setName] = useState('');
  const [modelId, setModelId] = useState('');
  const [metadata, setMetadata] = useState('');

  // Device info (edit mode)
  const [deviceSecret, setDeviceSecret] = useState('');
  const [deviceStatus, setDeviceStatus] = useState('');

  // Secret modal (create mode)
  const [secretModalOpen, setSecretModalOpen] = useState(false);
  const [newSecret, setNewSecret] = useState('');

  useEffect(() => {
    fetchThingModels();
    if (isEdit) fetchDevice();
  }, [id]);

  const fetchThingModels = async () => {
    try {
      const res = await thingModelApi.list(1, 100);
      setThingModels(res.data.list || []);
    } catch {
      // silent
    }
  };

  const fetchDevice = async () => {
    setLoading(true);
    try {
      const res = await deviceApi.get(id!);
      setName(res.data.name);
      setModelId(res.data.model_id || '');
      setMetadata(
        res.data.metadata && Object.keys(res.data.metadata).length > 0
          ? JSON.stringify(res.data.metadata, null, 2)
          : ''
      );
      setDeviceSecret(res.data.device_secret);
      setDeviceStatus(res.data.status);
    } catch {
      toast.error('获取设备信息失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error('请输入设备名称');
      return;
    }

    let parsedMetadata: Record<string, unknown> | undefined;
    if (metadata.trim()) {
      try {
        parsedMetadata = JSON.parse(metadata);
      } catch {
        toast.error('元数据格式错误，请输入有效的 JSON');
        return;
      }
    }

    setSaving(true);
    try {
      const data = { name, model_id: modelId || '', metadata: parsedMetadata };
      if (isEdit) {
        await deviceApi.update(id!, data);
        toast.success('更新成功');
        navigate('/devices');
      } else {
        const res = await deviceApi.create(data);
        setNewSecret(res.data.device_secret);
        setSecretModalOpen(true);
      }
    } catch (error: unknown) {
      const err = error as { response?: { data?: { msg?: string } } };
      toast.error(err.response?.data?.msg || '操作失败');
    } finally {
      setSaving(false);
    }
  };

  const handleCopySecret = async (secret: string) => {
    await navigator.clipboard.writeText(secret);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-3">
            <Button variant="ghost" size="icon" onClick={() => navigate('/devices')}>
              <ArrowLeft className="h-5 w-5" />
            </Button>
            <CardTitle>{isEdit ? '编辑设备' : '新建设备'}</CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={() => navigate('/devices')}>取消</Button>
            <Button onClick={handleSubmit} disabled={saving}>
              {saving ? '保存中...' : '保存'}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="max-w-xl space-y-6">
            <div className="space-y-2">
              <Label htmlFor="device-name">设备名称 *</Label>
              <Input
                id="device-name"
                placeholder="如：客厅温湿度传感器"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label>绑定物模型</Label>
              <Select value={modelId} onValueChange={setModelId}>
                <SelectTrigger>
                  <SelectValue placeholder="选择物模型（可选）" />
                </SelectTrigger>
                <SelectContent>
                  {thingModels.map((m) => (
                    <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="device-metadata">元数据 (JSON)</Label>
              <Textarea
                id="device-metadata"
                rows={4}
                placeholder='{"location": "客厅", "floor": 1}'
                value={metadata}
                onChange={(e) => setMetadata(e.target.value)}
              />
            </div>
          </div>

          {isEdit && (
            <div className="mt-8 p-4 bg-muted rounded-lg max-w-xl space-y-3">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">设备状态：</span>
                <Badge variant={deviceStatus === 'online' ? 'success' : 'secondary'}>
                  {deviceStatus === 'online' ? '在线' : '离线'}
                </Badge>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">设备密钥：</span>
                <code className="text-sm bg-background px-3 py-1 rounded font-mono flex-1 break-all">
                  {deviceSecret}
                </code>
                <Button variant="ghost" size="icon" onClick={() => handleCopySecret(deviceSecret)}>
                  {copied ? <Check className="h-4 w-4 text-primary" /> : <Copy className="h-4 w-4" />}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Secret modal after creation */}
      <Dialog open={secretModalOpen} onOpenChange={() => {}}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>设备创建成功</DialogTitle>
            <DialogDescription>
              请妥善保存设备密钥，此密钥用于设备接入认证：
            </DialogDescription>
          </DialogHeader>
          <div className="bg-muted p-4 rounded-lg font-mono text-sm break-all">
            {newSecret}
          </div>
          <p className="text-sm text-destructive">
            请立即复制保存，关闭后可在设备详情中查看。
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => handleCopySecret(newSecret)}>
              {copied ? <Check className="h-4 w-4 mr-2" /> : <Copy className="h-4 w-4 mr-2" />}
              复制密钥
            </Button>
            <Button onClick={() => { setSecretModalOpen(false); navigate('/devices'); }}>
              我已保存，关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
