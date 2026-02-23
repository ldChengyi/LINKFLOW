import { useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ArrowLeft, Plus, Pencil, Trash2, Mic, Upload, FileJson, Download, Copy, Check } from 'lucide-react';
import { thingModelApi, moduleApi } from '../api';
import type { Property, Event, Service, Param, Module, ThingModelModule } from '../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose,
} from '@/components/ui/dialog';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';

const dataTypes = [
  { value: 'int', label: '整数 (int)' },
  { value: 'float', label: '浮点数 (float)' },
  { value: 'bool', label: '布尔 (bool)' },
  { value: 'string', label: '字符串 (string)' },
  { value: 'enum', label: '枚举 (enum)' },
];

const accessModes = [
  { value: 'r', label: '只读 (r)' },
  { value: 'rw', label: '读写 (rw)' },
];

// ──────────────────────────────────────────────
// ParamEditor
// ──────────────────────────────────────────────
function ParamEditor({ params, onChange }: { params: Param[]; onChange: (p: Param[]) => void }) {
  const addParam = () => onChange([...params, { id: '', dataType: 'string' }]);
  const removeParam = (index: number) => onChange(params.filter((_, i) => i !== index));
  const updateParam = (index: number, field: keyof Param, value: string) => {
    const updated = [...params];
    updated[index] = { ...updated[index], [field]: value };
    onChange(updated);
  };

  return (
    <div className="space-y-2">
      {params.map((p, i) => (
        <div key={i} className="flex items-center gap-2">
          <Input
            placeholder="标识符"
            className="w-28"
            value={p.id}
            onChange={(e) => updateParam(i, 'id', e.target.value)}
          />
          <Input
            placeholder="名称"
            className="w-28"
            value={p.name || ''}
            onChange={(e) => updateParam(i, 'name', e.target.value)}
          />
          <Select value={p.dataType} onValueChange={(v) => updateParam(i, 'dataType', v)}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {dataTypes.map((dt) => (
                <SelectItem key={dt.value} value={dt.value}>{dt.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="ghost" size="icon" className="text-destructive shrink-0" onClick={() => removeParam(i)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <Button variant="outline" size="sm" className="w-full border-dashed" onClick={addParam}>
        <Plus className="h-4 w-4 mr-1" /> 添加参数
      </Button>
    </div>
  );
}

// ──────────────────────────────────────────────
// EnumEditor
// ──────────────────────────────────────────────
type EnumValue = { value: number; label: string };

function EnumEditor({
  enumValues,
  onChange,
}: {
  enumValues: EnumValue[];
  onChange: (ev: EnumValue[]) => void;
}) {
  const add = () => onChange([...enumValues, { value: enumValues.length, label: '' }]);
  const remove = (i: number) => onChange(enumValues.filter((_, idx) => idx !== i));
  const update = (i: number, field: keyof EnumValue, raw: string) => {
    const next = [...enumValues];
    if (field === 'value') next[i] = { ...next[i], value: Number(raw) };
    else next[i] = { ...next[i], label: raw };
    onChange(next);
  };

  return (
    <div className="space-y-2">
      {enumValues.length === 0 && (
        <p className="text-xs text-muted-foreground">暂无枚举项，点击下方添加</p>
      )}
      {enumValues.map((ev, i) => (
        <div key={i} className="flex items-center gap-2">
          <Input
            type="number"
            placeholder="值"
            className="w-20"
            value={ev.value}
            onChange={(e) => update(i, 'value', e.target.value)}
          />
          <Input
            placeholder="标签（如：开启）"
            className="flex-1"
            value={ev.label}
            onChange={(e) => update(i, 'label', e.target.value)}
          />
          <Button variant="ghost" size="icon" className="text-destructive shrink-0" onClick={() => remove(i)}>
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ))}
      <Button variant="outline" size="sm" className="w-full border-dashed" onClick={add}>
        <Plus className="h-4 w-4 mr-1" /> 添加枚举项
      </Button>
    </div>
  );
}

// ──────────────────────────────────────────────
// JSON 模板（供 AI 生成参考）
// ──────────────────────────────────────────────
const JSON_TEMPLATE = JSON.stringify(
  {
    name: "智能灯",
    description: "支持亮度、色温调节的智能 LED 灯",
    properties: [
      { id: "switch", name: "开关", dataType: "bool", accessMode: "rw" },
      { id: "brightness", name: "亮度", dataType: "int", unit: "%", min: 0, max: 100, step: 1, accessMode: "rw" },
      { id: "color_temp", name: "色温", dataType: "int", unit: "K", min: 2700, max: 6500, step: 100, accessMode: "rw" },
      {
        id: "mode", name: "模式", dataType: "enum",
        enumValues: [
          { value: 0, label: "普通" },
          { value: 1, label: "阅读" },
          { value: 2, label: "影院" },
        ],
        accessMode: "rw",
      },
      { id: "power", name: "功率", dataType: "float", unit: "W", min: 0, max: 20, accessMode: "r" },
    ],
    events: [
      {
        id: "overheating",
        name: "过热告警",
        params: [
          { id: "temp", name: "温度", dataType: "float" },
          { id: "level", name: "级别", dataType: "string" },
        ],
      },
    ],
    services: [
      {
        id: "reboot",
        name: "重启",
        inputParams: [],
        outputParams: [{ id: "result", name: "结果", dataType: "bool" }],
      },
      {
        id: "set_scene",
        name: "设置场景",
        inputParams: [{ id: "scene_id", name: "场景ID", dataType: "int" }],
        outputParams: [{ id: "success", name: "是否成功", dataType: "bool" }],
      },
    ],
  },
  null,
  2
);

// ──────────────────────────────────────────────
// 固定示例 JSON（智能温湿度传感器）
// ──────────────────────────────────────────────
const EXAMPLE_JSON = JSON.stringify(
  {
    name: "智能温湿度传感器",
    description: "监测环境温度与湿度，支持超限告警事件上报",
    properties: [
      { id: "temperature", name: "温度", dataType: "float", unit: "°C", min: -40, max: 85, accessMode: "r" },
      { id: "humidity", name: "湿度", dataType: "float", unit: "%RH", min: 0, max: 100, accessMode: "r" },
      { id: "battery", name: "电量", dataType: "int", unit: "%", min: 0, max: 100, accessMode: "r" },
      { id: "report_interval", name: "上报间隔", dataType: "int", unit: "s", min: 10, max: 3600, step: 10, accessMode: "rw" },
      {
        id: "status", name: "工作状态", dataType: "enum",
        enumValues: [
          { value: 0, label: "正常" },
          { value: 1, label: "告警" },
          { value: 2, label: "故障" },
        ],
        accessMode: "r",
      },
    ],
    events: [
      {
        id: "temp_alarm",
        name: "温度超限告警",
        params: [
          { id: "temperature", name: "当前温度", dataType: "float" },
          { id: "threshold", name: "阈值", dataType: "float" },
          { id: "alarm_type", name: "告警类型", dataType: "string" },
        ],
      },
      {
        id: "low_battery",
        name: "低电量告警",
        params: [{ id: "battery", name: "当前电量", dataType: "int" }],
      },
    ],
    services: [
      {
        id: "calibrate",
        name: "校准",
        inputParams: [
          { id: "temp_offset", name: "温度偏移", dataType: "float" },
          { id: "humi_offset", name: "湿度偏移", dataType: "float" },
        ],
        outputParams: [{ id: "success", name: "是否成功", dataType: "bool" }],
      },
      {
        id: "reboot",
        name: "重启",
        inputParams: [],
        outputParams: [{ id: "result", name: "结果", dataType: "bool" }],
      },
    ],
  },
  null,
  2
);

// ──────────────────────────────────────────────
// JSON 导入解析工具
// ──────────────────────────────────────────────
interface ImportedModel {
  name?: string;
  description?: string;
  properties?: Property[];
  events?: Event[];
  services?: Service[];
}

function parseImportJSON(raw: string): ImportedModel {
  const obj = JSON.parse(raw);
  // 宽松处理：接受顶层结构或 { data: ... } 包裹
  const src = obj?.data ?? obj;
  return {
    name: typeof src.name === 'string' ? src.name : undefined,
    description: typeof src.description === 'string' ? src.description : undefined,
    properties: Array.isArray(src.properties) ? src.properties : undefined,
    events: Array.isArray(src.events) ? src.events : undefined,
    services: Array.isArray(src.services) ? src.services : undefined,
  };
}

// ──────────────────────────────────────────────
// Main Component
// ──────────────────────────────────────────────
export default function ThingModelForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = !!id && id !== 'new';

  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  // Basic info
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  // Sub-items
  const [properties, setProperties] = useState<Property[]>([]);
  const [events, setEvents] = useState<Event[]>([]);
  const [services, setServices] = useState<Service[]>([]);

  // Modules
  const [availableModules, setAvailableModules] = useState<Module[]>([]);
  const [enabledModules, setEnabledModules] = useState<ThingModelModule[]>([]);

  // Property modal
  const [propModal, setPropModal] = useState(false);
  const [editingProp, setEditingProp] = useState<Property | null>(null);
  const [propForm, setPropForm] = useState<Partial<Property>>({});

  // Event modal
  const [eventModal, setEventModal] = useState(false);
  const [editingEvent, setEditingEvent] = useState<Event | null>(null);
  const [eventFormData, setEventFormData] = useState<{ id: string; name: string; params: Param[] }>({ id: '', name: '', params: [] });

  // Service modal
  const [serviceModal, setServiceModal] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const [serviceFormData, setServiceFormData] = useState<{ id: string; name: string; inputParams: Param[]; outputParams: Param[] }>({ id: '', name: '', inputParams: [], outputParams: [] });

  // Import modal
  const [importModal, setImportModal] = useState(false);
  const [importTab, setImportTab] = useState<'text' | 'file'>('text');
  const [importText, setImportText] = useState('');
  const [importPreview, setImportPreview] = useState<ImportedModel | null>(null);
  const [importError, setImportError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Export modal
  const [exportModal, setExportModal] = useState(false);
  const [exportCopied, setExportCopied] = useState(false);

  // Example modal
  const [exampleModal, setExampleModal] = useState(false);
  const [exampleCopied, setExampleCopied] = useState(false);

  useEffect(() => {
    fetchModules();
    if (isEdit) fetchData();
  }, [id]);

  const fetchModules = async () => {
    try {
      const res = await moduleApi.list();
      setAvailableModules(res.data || []);
    } catch { /* silent */ }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await thingModelApi.get(id!);
      setName(res.data.name);
      setDescription(res.data.description);
      setProperties(res.data.properties || []);
      setEvents(res.data.events || []);
      setServices(res.data.services || []);
      setEnabledModules(res.data.modules || []);
    } catch {
      toast.error('获取物模型失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    if (!name.trim()) {
      toast.error('请输入名称');
      return;
    }
    setSaving(true);
    try {
      const data = { name, description, properties, events, services, modules: enabledModules };
      if (isEdit) {
        await thingModelApi.update(id!, data);
        toast.success('更新成功');
      } else {
        await thingModelApi.create(data);
        toast.success('创建成功');
      }
      navigate('/thing-models');
    } catch (error: unknown) {
      const err = error as { response?: { data?: { msg?: string } } };
      toast.error(err.response?.data?.msg || '操作失败');
    } finally {
      setSaving(false);
    }
  };

  // ── Property handlers ──
  const openAddProp = () => {
    setEditingProp(null);
    setPropForm({ accessMode: 'r', dataType: 'float' });
    setPropModal(true);
  };
  const openEditProp = (prop: Property) => {
    setEditingProp(prop);
    setPropForm({ ...prop });
    setPropModal(true);
  };
  const saveProp = () => {
    if (!propForm.id || !propForm.name || !propForm.dataType || !propForm.accessMode) {
      toast.error('请填写必填项');
      return;
    }
    if (propForm.dataType === 'enum' && (!propForm.enumValues || propForm.enumValues.length === 0)) {
      toast.error('枚举类型至少需要一个枚举项');
      return;
    }
    const prop = propForm as Property;
    if (editingProp) {
      if (prop.id !== editingProp.id && properties.some((p) => p.id === prop.id)) {
        toast.error('属性标识符已存在');
        return;
      }
      setProperties(properties.map((p) => (p.id === editingProp.id ? prop : p)));
    } else {
      if (properties.some((p) => p.id === prop.id)) {
        toast.error('属性标识符已存在');
        return;
      }
      setProperties([...properties, prop]);
    }
    setPropModal(false);
  };

  // ── Event handlers ──
  const openAddEvent = () => {
    setEditingEvent(null);
    setEventFormData({ id: '', name: '', params: [] });
    setEventModal(true);
  };
  const openEditEvent = (event: Event) => {
    setEditingEvent(event);
    setEventFormData({ id: event.id, name: event.name, params: event.params ? [...event.params] : [] });
    setEventModal(true);
  };
  const saveEvent = () => {
    if (!eventFormData.id || !eventFormData.name) {
      toast.error('请填写必填项');
      return;
    }
    const ev: Event = { id: eventFormData.id, name: eventFormData.name, params: eventFormData.params };
    if (editingEvent) {
      if (ev.id !== editingEvent.id && events.some((e) => e.id === ev.id)) {
        toast.error('事件标识符已存在');
        return;
      }
      setEvents(events.map((e) => (e.id === editingEvent.id ? ev : e)));
    } else {
      if (events.some((e) => e.id === ev.id)) {
        toast.error('事件标识符已存在');
        return;
      }
      setEvents([...events, ev]);
    }
    setEventModal(false);
  };

  // ── Service handlers ──
  const openAddService = () => {
    setEditingService(null);
    setServiceFormData({ id: '', name: '', inputParams: [], outputParams: [] });
    setServiceModal(true);
  };
  const openEditService = (service: Service) => {
    setEditingService(service);
    setServiceFormData({
      id: service.id,
      name: service.name,
      inputParams: service.inputParams ? [...service.inputParams] : [],
      outputParams: service.outputParams ? [...service.outputParams] : [],
    });
    setServiceModal(true);
  };
  const saveService = () => {
    if (!serviceFormData.id || !serviceFormData.name) {
      toast.error('请填写必填项');
      return;
    }
    const svc: Service = {
      id: serviceFormData.id,
      name: serviceFormData.name,
      inputParams: serviceFormData.inputParams,
      outputParams: serviceFormData.outputParams,
    };
    if (editingService) {
      if (svc.id !== editingService.id && services.some((s) => s.id === svc.id)) {
        toast.error('服务标识符已存在');
        return;
      }
      setServices(services.map((s) => (s.id === editingService.id ? svc : s)));
    } else {
      if (services.some((s) => s.id === svc.id)) {
        toast.error('服务标识符已存在');
        return;
      }
      setServices([...services, svc]);
    }
    setServiceModal(false);
  };

  // ── Import handlers ──
  const tryParseImport = (raw: string) => {
    setImportError('');
    setImportPreview(null);
    if (!raw.trim()) return;
    try {
      const result = parseImportJSON(raw);
      setImportPreview(result);
    } catch (e) {
      setImportError('JSON 格式错误，请检查内容');
    }
  };

  const handleImportTextChange = (v: string) => {
    setImportText(v);
    tryParseImport(v);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const text = ev.target?.result as string;
      setImportText(text);
      tryParseImport(text);
    };
    reader.readAsText(file);
  };

  const confirmImport = () => {
    if (!importPreview) return;
    if (importPreview.name) setName(importPreview.name);
    if (importPreview.description !== undefined) setDescription(importPreview.description);
    if (importPreview.properties) setProperties(importPreview.properties);
    if (importPreview.events) setEvents(importPreview.events);
    if (importPreview.services) setServices(importPreview.services);
    setImportModal(false);
    setImportText('');
    setImportPreview(null);
    toast.success('导入成功，请检查并保存');
  };

  const openImportModal = () => {
    setImportText('');
    setImportPreview(null);
    setImportError('');
    setImportTab('text');
    setImportModal(true);
  };

  // ── Export helpers ──
  const buildExportJSON = () =>
    JSON.stringify({ name, description, properties, events, services }, null, 2);

  const handleExportCopy = async () => {
    await navigator.clipboard.writeText(buildExportJSON());
    setExportCopied(true);
    setTimeout(() => setExportCopied(false), 2000);
  };

  const handleExportDownload = () => {
    const blob = new Blob([buildExportJSON()], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name || 'thing-model'}.json`;
    a.click();
    URL.revokeObjectURL(url);
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
            <Button variant="ghost" size="icon" onClick={() => navigate('/thing-models')}>
              <ArrowLeft className="h-5 w-5" />
            </Button>
            <CardTitle>{isEdit ? '编辑物模型' : '新建物模型'}</CardTitle>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setExportModal(true)}>
              <Download className="h-4 w-4 mr-1" /> 导出 JSON
            </Button>
            <Button variant="outline" size="sm" onClick={() => setExampleModal(true)}>
              JSON 示例
            </Button>
            <Button variant="outline" size="sm" onClick={openImportModal}>
              <FileJson className="h-4 w-4 mr-1" /> 导入 JSON
            </Button>
            <Button variant="outline" onClick={() => navigate('/thing-models')}>取消</Button>
            <Button onClick={handleSubmit} disabled={saving}>
              {saving ? '保存中...' : '保存'}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-8">
          {/* Basic info */}
          <div className="max-w-xl space-y-4">
            <div className="space-y-2">
              <Label htmlFor="tm-name">名称 *</Label>
              <Input id="tm-name" placeholder="如：温湿度传感器" value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="tm-desc">描述</Label>
              <Textarea id="tm-desc" rows={3} placeholder="物模型描述" value={description} onChange={(e) => setDescription(e.target.value)} />
            </div>
          </div>

          {/* Tabs */}
          <Tabs defaultValue="properties">
            <TabsList>
              <TabsTrigger value="properties">属性 ({properties.length})</TabsTrigger>
              <TabsTrigger value="events">事件 ({events.length})</TabsTrigger>
              <TabsTrigger value="services">服务 ({services.length})</TabsTrigger>
              <TabsTrigger value="modules">模块 ({enabledModules.length})</TabsTrigger>
            </TabsList>

            {/* Properties tab */}
            <TabsContent value="properties" className="space-y-4">
              <Button size="sm" onClick={openAddProp}>
                <Plus className="h-4 w-4" /> 添加属性
              </Button>
              {properties.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">暂无属性</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>标识符</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead>数据类型</TableHead>
                      <TableHead>枚举项</TableHead>
                      <TableHead>单位</TableHead>
                      <TableHead>读写</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {properties.map((p) => (
                      <TableRow key={p.id}>
                        <TableCell className="font-mono text-sm">{p.id}</TableCell>
                        <TableCell>{p.name}</TableCell>
                        <TableCell>{p.dataType}</TableCell>
                        <TableCell className="text-muted-foreground text-sm">
                          {p.dataType === 'enum' && p.enumValues?.length
                            ? p.enumValues.map((ev) => `${ev.value}:${ev.label}`).join(', ')
                            : '-'}
                        </TableCell>
                        <TableCell className="text-muted-foreground">{p.unit || '-'}</TableCell>
                        <TableCell>{p.accessMode}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="sm" onClick={() => openEditProp(p)}>
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setProperties(properties.filter((x) => x.id !== p.id))}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </TabsContent>

            {/* Events tab */}
            <TabsContent value="events" className="space-y-4">
              <Button size="sm" onClick={openAddEvent}>
                <Plus className="h-4 w-4" /> 添加事件
              </Button>
              {events.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">暂无事件</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>标识符</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead>参数数量</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {events.map((e) => (
                      <TableRow key={e.id}>
                        <TableCell className="font-mono text-sm">{e.id}</TableCell>
                        <TableCell>{e.name}</TableCell>
                        <TableCell>{e.params?.length || 0}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="sm" onClick={() => openEditEvent(e)}>
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setEvents(events.filter((x) => x.id !== e.id))}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </TabsContent>

            {/* Services tab */}
            <TabsContent value="services" className="space-y-4">
              <Button size="sm" onClick={openAddService}>
                <Plus className="h-4 w-4" /> 添加服务
              </Button>
              {services.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">暂无服务</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>标识符</TableHead>
                      <TableHead>名称</TableHead>
                      <TableHead>输入参数</TableHead>
                      <TableHead>输出参数</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {services.map((s) => (
                      <TableRow key={s.id}>
                        <TableCell className="font-mono text-sm">{s.id}</TableCell>
                        <TableCell>{s.name}</TableCell>
                        <TableCell>{s.inputParams?.length || 0}</TableCell>
                        <TableCell>{s.outputParams?.length || 0}</TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="sm" onClick={() => openEditService(s)}>
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setServices(services.filter((x) => x.id !== s.id))}>
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </TabsContent>

            {/* Modules tab */}
            <TabsContent value="modules" className="space-y-4">
              {availableModules.length === 0 ? (
                <p className="text-sm text-muted-foreground py-8 text-center">暂无可用模块</p>
              ) : (
                <div className="space-y-4">
                  {availableModules.map((mod) => {
                    const enabled = enabledModules.find((m) => m.id === mod.id);
                    const isEnabled = !!enabled;
                    const exposedProps = enabled?.config?.exposed_properties || [];
                    const exposedSvcs = enabled?.config?.exposed_services || [];

                    const toggleModule = () => {
                      if (isEnabled) {
                        setEnabledModules(enabledModules.filter((m) => m.id !== mod.id));
                      } else {
                        setEnabledModules([...enabledModules, { id: mod.id, config: { exposed_properties: [], exposed_services: [] } }]);
                      }
                    };
                    const toggleProp = (propId: string) => {
                      setEnabledModules(enabledModules.map((m) => {
                        if (m.id !== mod.id) return m;
                        const current = m.config.exposed_properties || [];
                        const next = current.includes(propId) ? current.filter((p) => p !== propId) : [...current, propId];
                        return { ...m, config: { ...m.config, exposed_properties: next } };
                      }));
                    };
                    const toggleSvc = (svcId: string) => {
                      setEnabledModules(enabledModules.map((m) => {
                        if (m.id !== mod.id) return m;
                        const current = m.config.exposed_services || [];
                        const next = current.includes(svcId) ? current.filter((s) => s !== svcId) : [...current, svcId];
                        return { ...m, config: { ...m.config, exposed_services: next } };
                      }));
                    };

                    return (
                      <Card key={mod.id} className={isEnabled ? 'border-primary/50' : ''}>
                        <CardHeader className="pb-3">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              <Checkbox checked={isEnabled} onCheckedChange={toggleModule} />
                              <div className="flex items-center gap-2">
                                {mod.id === 'voice' && <Mic className="h-4 w-4 text-primary" />}
                                <CardTitle className="text-base">{mod.name}</CardTitle>
                              </div>
                              {isEnabled && <Badge variant="secondary">已启用</Badge>}
                            </div>
                          </div>
                          {mod.description && (
                            <p className="text-sm text-muted-foreground ml-7">{mod.description}</p>
                          )}
                        </CardHeader>
                        {isEnabled && (
                          <CardContent className="space-y-4 pt-0">
                            <div className="space-y-2">
                              <Label className="text-sm font-medium">暴露属性（模块可操控的属性）</Label>
                              {properties.filter((p) => p.accessMode === 'rw').length === 0 ? (
                                <p className="text-xs text-muted-foreground">无可写属性（仅 rw 属性可被模块操控）</p>
                              ) : (
                                <div className="flex flex-wrap gap-3">
                                  {properties.filter((p) => p.accessMode === 'rw').map((p) => (
                                    <label key={p.id} className="flex items-center gap-2 cursor-pointer">
                                      <Checkbox checked={exposedProps.includes(p.id)} onCheckedChange={() => toggleProp(p.id)} />
                                      <span className="text-sm">{p.name} <span className="text-muted-foreground font-mono text-xs">({p.id})</span></span>
                                    </label>
                                  ))}
                                </div>
                              )}
                            </div>
                            <div className="space-y-2">
                              <Label className="text-sm font-medium">暴露服务（模块可调用的服务）</Label>
                              {services.length === 0 ? (
                                <p className="text-xs text-muted-foreground">无可用服务</p>
                              ) : (
                                <div className="flex flex-wrap gap-3">
                                  {services.map((s) => (
                                    <label key={s.id} className="flex items-center gap-2 cursor-pointer">
                                      <Checkbox checked={exposedSvcs.includes(s.id)} onCheckedChange={() => toggleSvc(s.id)} />
                                      <span className="text-sm">{s.name} <span className="text-muted-foreground font-mono text-xs">({s.id})</span></span>
                                    </label>
                                  ))}
                                </div>
                              )}
                            </div>
                          </CardContent>
                        )}
                      </Card>
                    );
                  })}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* ── Property Modal ── */}
      <Dialog open={propModal} onOpenChange={setPropModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingProp ? '编辑属性' : '添加属性'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：temperature" value={propForm.id || ''} onChange={(e) => setPropForm({ ...propForm, id: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>名称 *</Label>
              <Input placeholder="如：温度" value={propForm.name || ''} onChange={(e) => setPropForm({ ...propForm, name: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>数据类型 *</Label>
              <Select
                value={propForm.dataType || 'float'}
                onValueChange={(v) => {
                  // 切换类型时清除互斥字段
                  const next: Partial<Property> = { ...propForm, dataType: v as Property['dataType'] };
                  if (v === 'enum') { delete next.min; delete next.max; delete next.step; }
                  else { delete next.enumValues; }
                  setPropForm(next);
                }}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {dataTypes.map((dt) => <SelectItem key={dt.value} value={dt.value}>{dt.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>

            {/* 枚举值编辑器 */}
            {propForm.dataType === 'enum' && (
              <div className="space-y-2">
                <Label>枚举项 *</Label>
                <EnumEditor
                  enumValues={propForm.enumValues || []}
                  onChange={(ev) => setPropForm({ ...propForm, enumValues: ev })}
                />
              </div>
            )}

            {/* 数值范围（仅 int / float） */}
            {(propForm.dataType === 'int' || propForm.dataType === 'float') && (
              <div className="grid grid-cols-3 gap-3">
                <div className="space-y-2">
                  <Label>最小值</Label>
                  <Input type="number" value={propForm.min ?? ''} onChange={(e) => setPropForm({ ...propForm, min: e.target.value ? Number(e.target.value) : undefined })} />
                </div>
                <div className="space-y-2">
                  <Label>最大值</Label>
                  <Input type="number" value={propForm.max ?? ''} onChange={(e) => setPropForm({ ...propForm, max: e.target.value ? Number(e.target.value) : undefined })} />
                </div>
                <div className="space-y-2">
                  <Label>步长</Label>
                  <Input type="number" value={propForm.step ?? ''} onChange={(e) => setPropForm({ ...propForm, step: e.target.value ? Number(e.target.value) : undefined })} />
                </div>
              </div>
            )}

            <div className="space-y-2">
              <Label>单位</Label>
              <Input placeholder="如：°C" value={propForm.unit || ''} onChange={(e) => setPropForm({ ...propForm, unit: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>读写类型 *</Label>
              <Select value={propForm.accessMode || 'r'} onValueChange={(v) => setPropForm({ ...propForm, accessMode: v as Property['accessMode'] })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {accessModes.map((am) => <SelectItem key={am.value} value={am.value}>{am.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild><Button variant="outline">取消</Button></DialogClose>
            <Button onClick={saveProp}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Event Modal ── */}
      <Dialog open={eventModal} onOpenChange={setEventModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingEvent ? '编辑事件' : '添加事件'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：high_temp_alarm" value={eventFormData.id} onChange={(e) => setEventFormData({ ...eventFormData, id: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>名称 *</Label>
              <Input placeholder="如：高温告警" value={eventFormData.name} onChange={(e) => setEventFormData({ ...eventFormData, name: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>输出参数</Label>
              <ParamEditor params={eventFormData.params} onChange={(p) => setEventFormData({ ...eventFormData, params: p })} />
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild><Button variant="outline">取消</Button></DialogClose>
            <Button onClick={saveEvent}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Service Modal ── */}
      <Dialog open={serviceModal} onOpenChange={setServiceModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingService ? '编辑服务' : '添加服务'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：calibrate" value={serviceFormData.id} onChange={(e) => setServiceFormData({ ...serviceFormData, id: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>名称 *</Label>
              <Input placeholder="如：校准" value={serviceFormData.name} onChange={(e) => setServiceFormData({ ...serviceFormData, name: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>输入参数</Label>
              <ParamEditor params={serviceFormData.inputParams} onChange={(p) => setServiceFormData({ ...serviceFormData, inputParams: p })} />
            </div>
            <div className="space-y-2">
              <Label>输出参数</Label>
              <ParamEditor params={serviceFormData.outputParams} onChange={(p) => setServiceFormData({ ...serviceFormData, outputParams: p })} />
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild><Button variant="outline">取消</Button></DialogClose>
            <Button onClick={saveService}>确定</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Example JSON Modal ── */}
      <Dialog open={exampleModal} onOpenChange={setExampleModal}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>JSON 示例 — 智能温湿度传感器</DialogTitle>
          </DialogHeader>
          <pre className="bg-muted rounded-md p-4 text-xs font-mono overflow-auto max-h-[55vh] whitespace-pre-wrap break-all">
            {EXAMPLE_JSON}
          </pre>
          <p className="text-xs text-muted-foreground">
            可将此示例复制给 AI，描述你的设备特性，让 AI 仿照格式生成对应的物模型 JSON，再通过「导入 JSON」导入。
          </p>
          <DialogFooter className="gap-2">
            <DialogClose asChild><Button variant="outline">关闭</Button></DialogClose>
            <Button
              variant="outline"
              onClick={async () => {
                await navigator.clipboard.writeText(EXAMPLE_JSON);
                setExampleCopied(true);
                setTimeout(() => setExampleCopied(false), 2000);
              }}
            >
              {exampleCopied ? <Check className="h-4 w-4 mr-1 text-primary" /> : <Copy className="h-4 w-4 mr-1" />}
              {exampleCopied ? '已复制' : '复制示例'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Export JSON Modal ── */}
      <Dialog open={exportModal} onOpenChange={setExportModal}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>导出物模型 JSON</DialogTitle>
          </DialogHeader>
          <div className="relative">
            <pre className="bg-muted rounded-md p-4 text-xs font-mono overflow-auto max-h-[50vh] whitespace-pre-wrap break-all">
              {buildExportJSON()}
            </pre>
          </div>
          <p className="text-xs text-muted-foreground">
            可将此 JSON 提供给 AI，让其生成或修改后再通过「导入 JSON」导入。
          </p>
          <DialogFooter className="gap-2">
            <DialogClose asChild><Button variant="outline">关闭</Button></DialogClose>
            <Button variant="outline" onClick={handleExportCopy}>
              {exportCopied ? <Check className="h-4 w-4 mr-1 text-primary" /> : <Copy className="h-4 w-4 mr-1" />}
              {exportCopied ? '已复制' : '复制 JSON'}
            </Button>
            <Button onClick={handleExportDownload}>
              <Download className="h-4 w-4 mr-1" /> 下载 .json
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ── Import JSON Modal ── */}
      <Dialog open={importModal} onOpenChange={setImportModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>导入物模型 JSON</DialogTitle>
          </DialogHeader>

          {/* 切换方式 */}
          <div className="flex gap-2 flex-wrap">
            <Button
              size="sm"
              variant={importTab === 'text' ? 'default' : 'outline'}
              onClick={() => setImportTab('text')}
            >
              粘贴文本
            </Button>
            <Button
              size="sm"
              variant={importTab === 'file' ? 'default' : 'outline'}
              onClick={() => { setImportTab('file'); fileInputRef.current?.click(); }}
            >
              <Upload className="h-4 w-4 mr-1" /> 上传文件
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".json,application/json"
              className="hidden"
              onChange={handleFileChange}
            />
            <Button
              size="sm"
              variant="outline"
              className="ml-auto text-muted-foreground"
              onClick={() => handleImportTextChange(JSON_TEMPLATE)}
            >
              加载示例模板
            </Button>
          </div>

          {/* 文本区域 */}
          <Textarea
            rows={10}
            placeholder={'{\n  "name": "智能灯",\n  "properties": [...],\n  "events": [...],\n  "services": [...]\n}'}
            value={importText}
            onChange={(e) => handleImportTextChange(e.target.value)}
            className="font-mono text-xs"
          />

          {/* 错误提示 */}
          {importError && (
            <p className="text-sm text-destructive">{importError}</p>
          )}

          {/* 解析预览 */}
          {importPreview && (
            <div className="rounded-md bg-muted p-3 text-sm space-y-1">
              <p className="font-medium text-foreground">解析结果预览</p>
              {importPreview.name && <p>名称：<span className="font-mono">{importPreview.name}</span></p>}
              <p>属性：{importPreview.properties?.length ?? 0} 条</p>
              <p>事件：{importPreview.events?.length ?? 0} 条</p>
              <p>服务：{importPreview.services?.length ?? 0} 条</p>
              <p className="text-xs text-muted-foreground pt-1">确认导入后将覆盖当前表单内容（模块配置保持不变）</p>
            </div>
          )}

          <DialogFooter>
            <DialogClose asChild><Button variant="outline">取消</Button></DialogClose>
            <Button onClick={confirmImport} disabled={!importPreview}>
              确认导入
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
