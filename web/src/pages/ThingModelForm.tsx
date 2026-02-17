import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ArrowLeft, Plus, Pencil, Trash2, Mic } from 'lucide-react';
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

  // --- Property handlers ---
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
    const prop = propForm as Property;
    if (editingProp) {
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

  // --- Event handlers ---
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

  // --- Service handlers ---
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

          {/* Tabs: properties / events / services */}
          <Tabs defaultValue="properties">
            <TabsList>
              <TabsTrigger value="properties">属性 ({properties.length})</TabsTrigger>
              <TabsTrigger value="events">事件 ({events.length})</TabsTrigger>
              <TabsTrigger value="services">服务 ({services.length})</TabsTrigger>
              <TabsTrigger value="modules">
                模块 ({enabledModules.length})
              </TabsTrigger>
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
                        const next = current.includes(propId)
                          ? current.filter((p) => p !== propId)
                          : [...current, propId];
                        return { ...m, config: { ...m.config, exposed_properties: next } };
                      }));
                    };

                    const toggleSvc = (svcId: string) => {
                      setEnabledModules(enabledModules.map((m) => {
                        if (m.id !== mod.id) return m;
                        const current = m.config.exposed_services || [];
                        const next = current.includes(svcId)
                          ? current.filter((s) => s !== svcId)
                          : [...current, svcId];
                        return { ...m, config: { ...m.config, exposed_services: next } };
                      }));
                    };

                    return (
                      <Card key={mod.id} className={isEnabled ? 'border-primary/50' : ''}>
                        <CardHeader className="pb-3">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                              <Checkbox
                                checked={isEnabled}
                                onCheckedChange={toggleModule}
                              />
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
                            {/* Exposed properties */}
                            <div className="space-y-2">
                              <Label className="text-sm font-medium">暴露属性（模块可操控的属性）</Label>
                              {properties.filter((p) => p.accessMode === 'rw').length === 0 ? (
                                <p className="text-xs text-muted-foreground">无可写属性（仅 rw 属性可被模块操控）</p>
                              ) : (
                                <div className="flex flex-wrap gap-3">
                                  {properties.filter((p) => p.accessMode === 'rw').map((p) => (
                                    <label key={p.id} className="flex items-center gap-2 cursor-pointer">
                                      <Checkbox
                                        checked={exposedProps.includes(p.id)}
                                        onCheckedChange={() => toggleProp(p.id)}
                                      />
                                      <span className="text-sm">{p.name} <span className="text-muted-foreground font-mono text-xs">({p.id})</span></span>
                                    </label>
                                  ))}
                                </div>
                              )}
                            </div>
                            {/* Exposed services */}
                            <div className="space-y-2">
                              <Label className="text-sm font-medium">暴露服务（模块可调用的服务）</Label>
                              {services.length === 0 ? (
                                <p className="text-xs text-muted-foreground">无可用服务</p>
                              ) : (
                                <div className="flex flex-wrap gap-3">
                                  {services.map((s) => (
                                    <label key={s.id} className="flex items-center gap-2 cursor-pointer">
                                      <Checkbox
                                        checked={exposedSvcs.includes(s.id)}
                                        onCheckedChange={() => toggleSvc(s.id)}
                                      />
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

      {/* Property Modal */}
      <Dialog open={propModal} onOpenChange={setPropModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingProp ? '编辑属性' : '添加属性'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：temperature" disabled={!!editingProp} value={propForm.id || ''} onChange={(e) => setPropForm({ ...propForm, id: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>名称 *</Label>
              <Input placeholder="如：温度" value={propForm.name || ''} onChange={(e) => setPropForm({ ...propForm, name: e.target.value })} />
            </div>
            <div className="space-y-2">
              <Label>数据类型 *</Label>
              <Select value={propForm.dataType || 'float'} onValueChange={(v) => setPropForm({ ...propForm, dataType: v as Property['dataType'] })}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {dataTypes.map((dt) => <SelectItem key={dt.value} value={dt.value}>{dt.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>单位</Label>
              <Input placeholder="如：°C" value={propForm.unit || ''} onChange={(e) => setPropForm({ ...propForm, unit: e.target.value })} />
            </div>
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

      {/* Event Modal */}
      <Dialog open={eventModal} onOpenChange={setEventModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingEvent ? '编辑事件' : '添加事件'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：high_temp_alarm" disabled={!!editingEvent} value={eventFormData.id} onChange={(e) => setEventFormData({ ...eventFormData, id: e.target.value })} />
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

      {/* Service Modal */}
      <Dialog open={serviceModal} onOpenChange={setServiceModal}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingService ? '编辑服务' : '添加服务'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>标识符 *</Label>
              <Input placeholder="如：calibrate" disabled={!!editingService} value={serviceFormData.id} onChange={(e) => setServiceFormData({ ...serviceFormData, id: e.target.value })} />
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
    </>
  );
}
