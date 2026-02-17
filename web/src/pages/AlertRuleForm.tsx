import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { alertRuleApi, deviceApi, thingModelApi } from '../api';
import type { Device, Property, AlertRule } from '../api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';

const operators = [
  { value: '>', label: '> 大于' },
  { value: '>=', label: '>= 大于等于' },
  { value: '<', label: '< 小于' },
  { value: '<=', label: '<= 小于等于' },
  { value: '==', label: '== 等于' },
  { value: '!=', label: '!= 不等于' },
];

const severities = [
  { value: 'info', label: '信息' },
  { value: 'warning', label: '警告' },
  { value: 'critical', label: '严重' },
];

export default function AlertRuleForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = !!id;

  const [devices, setDevices] = useState<Device[]>([]);
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(false);

  const [name, setName] = useState('');
  const [deviceId, setDeviceId] = useState('');
  const [propertyId, setPropertyId] = useState('');
  const [operator, setOperator] = useState('>');
  const [threshold, setThreshold] = useState('0');
  const [severity, setSeverity] = useState('warning');
  const [enabled, setEnabled] = useState(true);

  useEffect(() => {
    fetchDevices();
    if (isEdit) fetchRule();
  }, []);

  useEffect(() => {
    if (deviceId) fetchProperties(deviceId);
  }, [deviceId]);

  const fetchDevices = async () => {
    try {
      const res = await deviceApi.list(1, 100);
      setDevices(res.data.list || []);
    } catch { /* silent */ }
  };

  const fetchProperties = async (devId: string) => {
    const device = devices.find((d) => d.id === devId);
    if (!device?.model_id) {
      setProperties([]);
      return;
    }
    try {
      const res = await thingModelApi.get(device.model_id);
      setProperties(res.data.properties || []);
    } catch {
      setProperties([]);
    }
  };

  const fetchRule = async () => {
    try {
      const res = await alertRuleApi.get(id!);
      const rule: AlertRule = res.data;
      setName(rule.name);
      setDeviceId(rule.device_id);
      setPropertyId(rule.property_id);
      setOperator(rule.operator);
      setThreshold(String(rule.threshold));
      setSeverity(rule.severity);
      setEnabled(rule.enabled);
    } catch {
      toast.error('获取规则失败');
      navigate('/alert-rules');
    }
  };

  const handleSubmit = async () => {
    if (!name.trim()) { toast.error('请输入规则名称'); return; }
    if (!deviceId) { toast.error('请选择设备'); return; }
    if (!propertyId) { toast.error('请选择属性'); return; }

    setLoading(true);
    try {
      const data = {
        name: name.trim(),
        device_id: deviceId,
        property_id: propertyId,
        operator,
        threshold: parseFloat(threshold) || 0,
        severity,
        enabled,
      };
      if (isEdit) {
        await alertRuleApi.update(id!, data);
        toast.success('更新成功');
      } else {
        await alertRuleApi.create(data);
        toast.success('创建成功');
      }
      navigate('/alert-rules');
    } catch {
      toast.error(isEdit ? '更新失败' : '创建失败');
    } finally {
      setLoading(false);
    }
  };

  // 当 devices 加载完成后，如果是编辑模式且 deviceId 已设置，重新加载属性
  useEffect(() => {
    if (devices.length > 0 && deviceId) {
      fetchProperties(deviceId);
    }
  }, [devices]);

  return (
    <Card className="max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle>{isEdit ? '编辑告警规则' : '新建告警规则'}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-2">
          <Label>规则名称</Label>
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="如：温度过高告警" />
        </div>

        <div className="space-y-2">
          <Label>设备</Label>
          <Select value={deviceId} onValueChange={setDeviceId}>
            <SelectTrigger>
              <SelectValue placeholder="选择设备" />
            </SelectTrigger>
            <SelectContent>
              {devices.map((d) => (
                <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label>属性</Label>
          {properties.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {deviceId ? '该设备未绑定物模型或无属性定义' : '请先选择设备'}
            </p>
          ) : (
            <Select value={propertyId} onValueChange={setPropertyId}>
              <SelectTrigger>
                <SelectValue placeholder="选择属性" />
              </SelectTrigger>
              <SelectContent>
                {properties.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name} ({p.id}) · {p.dataType}{p.unit ? ` (${p.unit})` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>运算符</Label>
            <Select value={operator} onValueChange={setOperator}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {operators.map((op) => (
                  <SelectItem key={op.value} value={op.value}>{op.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>阈值</Label>
            <Input type="number" value={threshold} onChange={(e) => setThreshold(e.target.value)} />
          </div>
        </div>

        <div className="space-y-2">
          <Label>告警级别</Label>
          <Select value={severity} onValueChange={setSeverity}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {severities.map((s) => (
                <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-3">
          <Switch checked={enabled} onCheckedChange={setEnabled} />
          <Label>启用规则</Label>
        </div>

        <div className="flex items-center gap-3 pt-4">
          <Button onClick={handleSubmit} disabled={loading}>
            {loading ? '保存中...' : (isEdit ? '更新' : '创建')}
          </Button>
          <Button variant="outline" onClick={() => navigate('/alert-rules')}>取消</Button>
        </div>
      </CardContent>
    </Card>
  );
}
