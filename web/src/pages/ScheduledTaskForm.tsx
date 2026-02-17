import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Clock, ArrowLeft } from 'lucide-react';
import { scheduledTaskApi, deviceApi, thingModelApi } from '../api';
import type { Device, ThingModel } from '../api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Spinner } from '@/components/ui/spinner';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';

const CRON_PRESETS = [
  { label: '每分钟', value: '* * * * *' },
  { label: '每小时', value: '0 * * * *' },
  { label: '每天 0:00', value: '0 0 * * *' },
  { label: '每天 8:00', value: '0 8 * * *' },
  { label: '每周一 8:00', value: '0 8 * * 1' },
];

export default function ScheduledTaskForm() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isEdit = !!id;

  const [devices, setDevices] = useState<Device[]>([]);
  const [thingModel, setThingModel] = useState<ThingModel | null>(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const [name, setName] = useState('');
  const [deviceId, setDeviceId] = useState('');
  const [cronExpr, setCronExpr] = useState('0 * * * *');
  const [actionType, setActionType] = useState('property_set');
  const [propertyId, setPropertyId] = useState('');
  const [serviceId, setServiceId] = useState('');
  const [value, setValue] = useState('');
  const [enabled, setEnabled] = useState(true);

  useEffect(() => { fetchDevices(); }, []);

  useEffect(() => {
    if (isEdit) fetchTask();
  }, [id]);

  useEffect(() => {
    if (deviceId) loadThingModel(deviceId);
    else setThingModel(null);
  }, [deviceId]);

  const fetchDevices = async () => {
    try {
      const res = await deviceApi.list(1, 100);
      setDevices(res.data.list || []);
    } catch { /* silent */ }
  };

  const loadThingModel = async (devId: string) => {
    const dev = devices.find(d => d.id === devId);
    if (!dev?.model_id) { setThingModel(null); return; }
    try {
      const res = await thingModelApi.get(dev.model_id);
      setThingModel(res.data);
    } catch { setThingModel(null); }
  };

  const fetchTask = async () => {
    setLoading(true);
    try {
      const res = await scheduledTaskApi.get(id!);
      const t = res.data;
      setName(t.name);
      setDeviceId(t.device_id);
      setCronExpr(t.cron_expr);
      setActionType(t.action_type);
      setPropertyId(t.property_id || '');
      setServiceId(t.service_id || '');
      setValue(t.value != null ? JSON.stringify(t.value) : '');
      setEnabled(t.enabled);
    } catch { toast.error('获取任务失败'); }
    finally { setLoading(false); }
  };

  const rwProperties = (thingModel?.properties || []).filter(p => p.accessMode === 'rw');
  const services = thingModel?.services || [];

  const handleSubmit = async () => {
    if (!name.trim() || !deviceId || !cronExpr) {
      toast.error('请填写必填字段');
      return;
    }
    if (actionType === 'property_set' && !propertyId) {
      toast.error('请选择属性');
      return;
    }
    if (actionType === 'service_invoke' && !serviceId) {
      toast.error('请选择服务');
      return;
    }

    let parsedValue: unknown = null;
    if (value.trim()) {
      try { parsedValue = JSON.parse(value); }
      catch { toast.error('值必须是合法的 JSON'); return; }
    }

    setSubmitting(true);
    try {
      const data = {
        device_id: deviceId, name: name.trim(), cron_expr: cronExpr,
        action_type: actionType,
        property_id: actionType === 'property_set' ? propertyId : undefined,
        service_id: actionType === 'service_invoke' ? serviceId : undefined,
        value: parsedValue, enabled,
      };
      if (isEdit) await scheduledTaskApi.update(id!, data);
      else await scheduledTaskApi.create(data);
      toast.success(isEdit ? '更新成功' : '创建成功');
      navigate('/scheduled-tasks');
    } catch { toast.error(isEdit ? '更新失败' : '创建失败'); }
    finally { setSubmitting(false); }
  };

  if (loading) return <div className="flex justify-center py-20"><Spinner size="lg" /></div>;

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate('/scheduled-tasks')}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <Clock className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold">{isEdit ? '编辑定时任务' : '创建定时任务'}</h2>
      </div>

      <Card>
        <CardHeader><CardTitle className="text-base">基本信息</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>任务名称 *</Label>
            <Input value={name} onChange={e => setName(e.target.value)} placeholder="如：每小时上报温度" maxLength={100} />
          </div>
          <div className="space-y-2">
            <Label>设备 *</Label>
            <Select value={deviceId} onValueChange={setDeviceId}>
              <SelectTrigger><SelectValue placeholder="选择设备" /></SelectTrigger>
              <SelectContent>
                {devices.map(d => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Cron 表达式 *</Label>
            <Input value={cronExpr} onChange={e => setCronExpr(e.target.value)} placeholder="分 时 日 月 周" />
            <div className="flex flex-wrap gap-1">
              {CRON_PRESETS.map(p => (
                <Button key={p.value} variant="outline" size="sm" className="text-xs h-7"
                  onClick={() => setCronExpr(p.value)}>{p.label}</Button>
              ))}
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Label>启用</Label>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className="text-base">执行动作</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>动作类型 *</Label>
            <Select value={actionType} onValueChange={v => { setActionType(v); setPropertyId(''); setServiceId(''); }}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="property_set">属性设置</SelectItem>
                <SelectItem value="service_invoke">服务调用</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {!thingModel && deviceId && (
            <p className="text-sm text-muted-foreground">该设备未绑定物模型</p>
          )}

          {actionType === 'property_set' && rwProperties.length > 0 && (
            <div className="space-y-2">
              <Label>属性 *</Label>
              <Select value={propertyId} onValueChange={setPropertyId}>
                <SelectTrigger><SelectValue placeholder="选择属性" /></SelectTrigger>
                <SelectContent>
                  {rwProperties.map(p => <SelectItem key={p.id} value={p.id}>{p.name} ({p.id})</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          )}

          {actionType === 'service_invoke' && services.length > 0 && (
            <div className="space-y-2">
              <Label>服务 *</Label>
              <Select value={serviceId} onValueChange={setServiceId}>
                <SelectTrigger><SelectValue placeholder="选择服务" /></SelectTrigger>
                <SelectContent>
                  {services.map(s => <SelectItem key={s.id} value={s.id}>{s.name} ({s.id})</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label>值 (JSON)</Label>
            <Input value={value} onChange={e => setValue(e.target.value)}
              placeholder={actionType === 'property_set' ? '如: 25 或 true' : '如: {"param1": "value1"}'} />
          </div>
        </CardContent>
      </Card>

      <div className="flex gap-3">
        <Button variant="outline" className="flex-1" onClick={() => navigate('/scheduled-tasks')}>取消</Button>
        <Button className="flex-1" onClick={handleSubmit} disabled={submitting}>
          {submitting ? '提交中...' : isEdit ? '更新' : '创建'}
        </Button>
      </div>
    </div>
  );
}
