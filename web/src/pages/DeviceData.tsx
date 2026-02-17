import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { Activity, RefreshCw, Copy, Check, Mic, Blocks } from 'lucide-react';
import { deviceApi, thingModelApi } from '../api';
import type { Device, ThingModel, DeviceLatestData, Property } from '../api';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';

export default function DeviceData() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedId, setSelectedId] = useState('');
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [thingModel, setThingModel] = useState<ThingModel | null>(null);
  const [latestData, setLatestData] = useState<DeviceLatestData | null>(null);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    fetchDevices();
  }, []);

  useEffect(() => {
    if (selectedId) fetchDeviceDetail(selectedId);
  }, [selectedId]);

  const fetchDevices = async () => {
    try {
      const res = await deviceApi.list(1, 100);
      setDevices(res.data.list || []);
    } catch {
      toast.error('获取设备列表失败');
    }
  };

  const fetchDeviceDetail = async (id: string) => {
    setLoading(true);
    setThingModel(null);
    setLatestData(null);
    try {
      const [deviceRes, dataRes] = await Promise.all([
        deviceApi.get(id),
        deviceApi.latestData(id),
      ]);
      setSelectedDevice(deviceRes.data);
      setLatestData(dataRes.data);

      if (deviceRes.data.model_id) {
        try {
          const modelRes = await thingModelApi.get(deviceRes.data.model_id);
          setThingModel(modelRes.data);
        } catch { /* silent */ }
      }
    } catch {
      toast.error('获取设备数据失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    if (!selectedId) return;
    setRefreshing(true);
    try {
      const res = await deviceApi.latestData(selectedId);
      setLatestData(res.data);
    } catch {
      toast.error('刷新失败');
    } finally {
      setRefreshing(false);
    }
  };

  const properties = thingModel?.properties || [];

  return (
    <div className="space-y-6">
      {/* 顶部：设备选择 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Activity className="h-5 w-5 text-primary" />
          <h2 className="text-lg font-semibold">设备数据</h2>
        </div>
        <div className="flex items-center gap-3">
          <Select value={selectedId} onValueChange={setSelectedId}>
            <SelectTrigger className="w-64">
              <SelectValue placeholder="选择设备" />
            </SelectTrigger>
            <SelectContent>
              {devices.map((d) => (
                <SelectItem key={d.id} value={d.id}>
                  <span className="flex items-center gap-2">
                    <span className={`h-2 w-2 rounded-full ${d.status === 'online' ? 'bg-emerald-400' : 'bg-muted-foreground/40'}`} />
                    {d.name}
                  </span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {selectedId && (
            <Button variant="outline" size="icon" onClick={handleRefresh} disabled={refreshing}>
              <RefreshCw className={`h-4 w-4 ${refreshing ? 'animate-spin' : ''}`} />
            </Button>
          )}
        </div>
      </div>

      {!selectedId && (
        <div className="text-center py-20 text-muted-foreground">
          请从右上角下拉框选择一个设备
        </div>
      )}

      {selectedId && loading && (
        <div className="flex items-center justify-center py-20">
          <Spinner size="lg" />
        </div>
      )}

      {selectedId && !loading && selectedDevice && (
        <Tabs defaultValue="properties">
          <TabsList>
            <TabsTrigger value="properties">属性数据</TabsTrigger>
            <TabsTrigger value="model">物模型详情</TabsTrigger>
            {thingModel?.modules && thingModel.modules.length > 0 && (
              <TabsTrigger value="modules">模块 ({thingModel.modules.length})</TabsTrigger>
            )}
            <TabsTrigger value="api">接口详情</TabsTrigger>
          </TabsList>

          {/* Tab 1: 属性数据 */}
          <TabsContent value="properties">
            {!thingModel ? (
              <Card>
                <CardContent className="py-10 text-center text-muted-foreground">
                  该设备未绑定物模型，无法展示属性数据
                </CardContent>
              </Card>
            ) : !latestData ? (
              <Card>
                <CardContent className="py-10 text-center text-muted-foreground">
                  暂无上报数据
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center gap-3 text-sm text-muted-foreground">
                  <span>上报时间：{new Date(latestData.time).toLocaleString()}</span>
                  {!latestData.valid && <Badge variant="destructive">校验异常</Badge>}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {properties.map((prop) => (
                    <PropertyCard
                      key={prop.id}
                      property={prop}
                      value={latestData.payload[prop.id]}
                      error={latestData.errors?.[prop.id]}
                    />
                  ))}

                  {/* 未知字段 */}
                  {Object.keys(latestData.payload)
                    .filter((key) => !properties.some((p) => p.id === key))
                    .map((key) => (
                      <PropertyCard
                        key={key}
                        property={{ id: key, name: key, dataType: 'string', accessMode: 'r' }}
                        value={latestData.payload[key]}
                        error={latestData.errors?.[key] || '未知属性'}
                        unknown
                      />
                    ))}
                </div>
              </div>
            )}
          </TabsContent>

          {/* Tab 2: 物模型详情 */}
          <TabsContent value="model">
            {!thingModel ? (
              <Card>
                <CardContent className="py-10 text-center text-muted-foreground">
                  该设备未绑定物模型
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-6">
                {/* 属性 */}
                <Card>
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">属性 Properties</CardTitle>
                      <Badge variant="secondary">{properties.length} 项</Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {properties.length === 0 ? (
                      <p className="text-sm text-muted-foreground">暂无属性定义</p>
                    ) : (
                      <>
                        {/* 属性列表 */}
                        <div className="rounded-lg border overflow-hidden">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="bg-muted/50 text-muted-foreground">
                                <th className="text-left px-4 py-2 font-medium">标识符</th>
                                <th className="text-left px-4 py-2 font-medium">名称</th>
                                <th className="text-left px-4 py-2 font-medium">类型</th>
                                <th className="text-left px-4 py-2 font-medium">范围</th>
                                <th className="text-left px-4 py-2 font-medium">单位</th>
                                <th className="text-left px-4 py-2 font-medium">读写</th>
                              </tr>
                            </thead>
                            <tbody>
                              {properties.map((p) => (
                                <tr key={p.id} className="border-t">
                                  <td className="px-4 py-2 font-mono text-xs">{p.id}</td>
                                  <td className="px-4 py-2">{p.name}</td>
                                  <td className="px-4 py-2">
                                    <Badge variant="outline" className="text-xs">{p.dataType}</Badge>
                                    {p.dataType === 'enum' && p.enumValues && (
                                      <span className="ml-2 text-xs text-muted-foreground">
                                        {p.enumValues.map((e) => `${e.value}=${e.label}`).join(', ')}
                                      </span>
                                    )}
                                  </td>
                                  <td className="px-4 py-2 text-xs text-muted-foreground">
                                    {p.min !== undefined && p.max !== undefined ? `${p.min} ~ ${p.max}` : '-'}
                                    {p.step !== undefined && ` (步长 ${p.step})`}
                                  </td>
                                  <td className="px-4 py-2 text-xs">{p.unit || '-'}</td>
                                  <td className="px-4 py-2">
                                    <Badge variant={p.accessMode === 'rw' ? 'default' : 'secondary'} className="text-xs">
                                      {p.accessMode === 'rw' ? '读写' : '只读'}
                                    </Badge>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>

                        {/* 上报格式 */}
                        <div className="space-y-2">
                          <div className="flex items-center gap-2">
                            <ApiRow method="PUB" path={`devices/${selectedDevice.id}/telemetry/up`} desc="设备上报属性" />
                          </div>
                          <CodeBlock
                            json={Object.fromEntries(
                              properties.map((p) => [p.id, sampleValue(p)])
                            )}
                          />
                        </div>

                        {/* 下行格式（仅 rw 属性） */}
                        {properties.some((p) => p.accessMode === 'rw') && (
                          <div className="space-y-2">
                            <div className="flex items-center gap-2">
                              <ApiRow method="SUB" path={`devices/${selectedDevice.id}/telemetry/down`} desc="平台下发属性设置" />
                            </div>
                            <CodeBlock
                              json={Object.fromEntries(
                                properties.filter((p) => p.accessMode === 'rw').map((p) => [p.id, sampleValue(p)])
                              )}
                            />
                          </div>
                        )}
                      </>
                    )}
                  </CardContent>
                </Card>

                {/* 事件 */}
                <Card>
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">事件 Events</CardTitle>
                      <Badge variant="secondary">{thingModel.events?.length || 0} 项</Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {!thingModel.events?.length ? (
                      <p className="text-sm text-muted-foreground">暂无事件定义</p>
                    ) : (
                      thingModel.events.map((evt) => (
                        <div key={evt.id} className="space-y-2 rounded-lg border p-4">
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-sm">{evt.name}</span>
                            <span className="font-mono text-xs text-muted-foreground">{evt.id}</span>
                          </div>
                          {evt.params && evt.params.length > 0 && (
                            <div className="rounded border overflow-hidden">
                              <table className="w-full text-sm">
                                <thead>
                                  <tr className="bg-muted/50 text-muted-foreground">
                                    <th className="text-left px-3 py-1.5 font-medium text-xs">参数标识</th>
                                    <th className="text-left px-3 py-1.5 font-medium text-xs">名称</th>
                                    <th className="text-left px-3 py-1.5 font-medium text-xs">类型</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {evt.params.map((p) => (
                                    <tr key={p.id} className="border-t">
                                      <td className="px-3 py-1.5 font-mono text-xs">{p.id}</td>
                                      <td className="px-3 py-1.5 text-xs">{p.name || '-'}</td>
                                      <td className="px-3 py-1.5"><Badge variant="outline" className="text-xs">{p.dataType}</Badge></td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            </div>
                          )}
                          {/* 事件上报格式 */}
                          <div className="space-y-1">
                            <ApiRow method="PUB" path={`devices/${selectedDevice.id}/event`} desc="设备上报事件" />
                            <CodeBlock
                              json={{
                                id: evt.id,
                                time: new Date().toISOString(),
                                params: Object.fromEntries(
                                  (evt.params || []).map((p) => [p.id, sampleValueByType(p.dataType)])
                                ),
                              }}
                            />
                          </div>
                        </div>
                      ))
                    )}
                  </CardContent>
                </Card>

                {/* 服务 */}
                <Card>
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-base">服务 Services</CardTitle>
                      <Badge variant="secondary">{thingModel.services?.length || 0} 项</Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {!thingModel.services?.length ? (
                      <p className="text-sm text-muted-foreground">暂无服务定义</p>
                    ) : (
                      thingModel.services.map((svc) => (
                        <div key={svc.id} className="space-y-3 rounded-lg border p-4">
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-sm">{svc.name}</span>
                            <span className="font-mono text-xs text-muted-foreground">{svc.id}</span>
                          </div>

                          {/* 入参 */}
                          {svc.inputParams && svc.inputParams.length > 0 && (
                            <div>
                              <p className="text-xs text-muted-foreground mb-1">输入参数</p>
                              <div className="rounded border overflow-hidden">
                                <table className="w-full text-sm">
                                  <thead>
                                    <tr className="bg-muted/50 text-muted-foreground">
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">参数标识</th>
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">名称</th>
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">类型</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {svc.inputParams.map((p) => (
                                      <tr key={p.id} className="border-t">
                                        <td className="px-3 py-1.5 font-mono text-xs">{p.id}</td>
                                        <td className="px-3 py-1.5 text-xs">{p.name || '-'}</td>
                                        <td className="px-3 py-1.5"><Badge variant="outline" className="text-xs">{p.dataType}</Badge></td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
                            </div>
                          )}

                          {/* 出参 */}
                          {svc.outputParams && svc.outputParams.length > 0 && (
                            <div>
                              <p className="text-xs text-muted-foreground mb-1">输出参数</p>
                              <div className="rounded border overflow-hidden">
                                <table className="w-full text-sm">
                                  <thead>
                                    <tr className="bg-muted/50 text-muted-foreground">
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">参数标识</th>
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">名称</th>
                                      <th className="text-left px-3 py-1.5 font-medium text-xs">类型</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {svc.outputParams.map((p) => (
                                      <tr key={p.id} className="border-t">
                                        <td className="px-3 py-1.5 font-mono text-xs">{p.id}</td>
                                        <td className="px-3 py-1.5 text-xs">{p.name || '-'}</td>
                                        <td className="px-3 py-1.5"><Badge variant="outline" className="text-xs">{p.dataType}</Badge></td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              </div>
                            </div>
                          )}

                          {/* 服务调用下发格式 */}
                          <div className="space-y-1">
                            <ApiRow method="SUB" path={`devices/${selectedDevice.id}/service/invoke`} desc="平台下发服务调用" />
                            <CodeBlock
                              json={{
                                id: 'req_001',
                                service: svc.id,
                                params: Object.fromEntries(
                                  (svc.inputParams || []).map((p) => [p.id, sampleValueByType(p.dataType)])
                                ),
                              }}
                            />
                          </div>

                          {/* 服务回复格式 */}
                          <div className="space-y-1">
                            <ApiRow method="PUB" path={`devices/${selectedDevice.id}/service/reply`} desc="设备回复执行结果" />
                            <CodeBlock
                              json={{
                                id: 'req_001',
                                service: svc.id,
                                code: 200,
                                data: Object.fromEntries(
                                  (svc.outputParams || []).map((p) => [p.id, sampleValueByType(p.dataType)])
                                ),
                              }}
                            />
                          </div>
                        </div>
                      ))
                    )}
                  </CardContent>
                </Card>

                {/* 语音模块 */}
                {thingModel.modules?.some((m) => m.id === 'voice') && (() => {
                  const voiceMod = thingModel.modules.find((m) => m.id === 'voice')!;
                  const exposedProps = voiceMod.config?.exposed_properties || [];
                  const exposedSvcs = voiceMod.config?.exposed_services || [];
                  return (
                    <Card className="border-primary/30">
                      <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                          <CardTitle className="text-base">语音模块 Voice</CardTitle>
                          <Badge variant="default">已启用</Badge>
                        </div>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div className="space-y-2">
                          <p className="text-sm font-medium">可操控属性</p>
                          {exposedProps.length === 0 ? (
                            <p className="text-xs text-muted-foreground">未配置暴露属性</p>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {exposedProps.map((pid) => {
                                const prop = properties.find((p) => p.id === pid);
                                return (
                                  <Badge key={pid} variant="outline">
                                    {prop ? `${prop.name} (${pid})` : pid}
                                  </Badge>
                                );
                              })}
                            </div>
                          )}
                        </div>
                        <div className="space-y-2">
                          <p className="text-sm font-medium">可调用服务</p>
                          {exposedSvcs.length === 0 ? (
                            <p className="text-xs text-muted-foreground">未配置暴露服务</p>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {exposedSvcs.map((sid) => {
                                const svc = thingModel.services?.find((s) => s.id === sid);
                                return (
                                  <Badge key={sid} variant="outline">
                                    {svc ? `${svc.name} (${sid})` : sid}
                                  </Badge>
                                );
                              })}
                            </div>
                          )}
                        </div>
                        {/* 语音上报格式 */}
                        <div className="space-y-1">
                          <ApiRow method="PUB" path={`devices/${selectedDevice.id}/voice/up`} desc="语音文本上报" />
                          <CodeBlock json={{ text: "打开灯" }} />
                        </div>
                        <div className="space-y-1">
                          <ApiRow method="SUB" path={`devices/${selectedDevice.id}/voice/down`} desc="语音执行结果" />
                          <CodeBlock json={{ success: true, message: "已执行", action: "set_property" }} />
                        </div>
                      </CardContent>
                    </Card>
                  );
                })()}
              </div>
            )}
          </TabsContent>

          {/* Tab: 模块 */}
          {thingModel?.modules && thingModel.modules.length > 0 && (
            <TabsContent value="modules">
              <div className="space-y-4">
                {thingModel.modules.map((mod) => {
                  const exposedProps = mod.config?.exposed_properties || [];
                  const exposedSvcs = mod.config?.exposed_services || [];
                  const Icon = mod.id === 'voice' ? Mic : Blocks;

                  return (
                    <Card key={mod.id}>
                      <CardHeader className="pb-3">
                        <div className="flex items-center gap-3">
                          <div className="p-2 rounded-lg bg-primary/10">
                            <Icon className="h-5 w-5 text-primary" />
                          </div>
                          <div>
                            <CardTitle className="text-base">{mod.id === 'voice' ? '语音控制模块' : mod.id}</CardTitle>
                            <p className="text-xs text-muted-foreground mt-0.5">模块ID: {mod.id}</p>
                          </div>
                          <Badge variant="default" className="ml-auto">已启用</Badge>
                        </div>
                      </CardHeader>
                      <CardContent className="space-y-4">
                        {/* 暴露的属性 */}
                        <div className="space-y-2">
                          <p className="text-sm font-medium">可操控属性</p>
                          {exposedProps.length === 0 ? (
                            <p className="text-xs text-muted-foreground">未配置暴露属性</p>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {exposedProps.map((pid) => {
                                const prop = properties.find((p) => p.id === pid);
                                return (
                                  <Badge key={pid} variant="outline" className="gap-1.5">
                                    <span>{prop ? prop.name : pid}</span>
                                    <span className="text-muted-foreground font-mono text-xs">{pid}</span>
                                    {prop && <span className="text-muted-foreground">· {prop.dataType}{prop.unit ? ` (${prop.unit})` : ''}</span>}
                                  </Badge>
                                );
                              })}
                            </div>
                          )}
                        </div>

                        {/* 暴露的服务 */}
                        <div className="space-y-2">
                          <p className="text-sm font-medium">可调用服务</p>
                          {exposedSvcs.length === 0 ? (
                            <p className="text-xs text-muted-foreground">未配置暴露服务</p>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {exposedSvcs.map((sid) => {
                                const svc = thingModel.services?.find((s) => s.id === sid);
                                return (
                                  <Badge key={sid} variant="outline" className="gap-1.5">
                                    <span>{svc ? svc.name : sid}</span>
                                    <span className="text-muted-foreground font-mono text-xs">{sid}</span>
                                  </Badge>
                                );
                              })}
                            </div>
                          )}
                        </div>

                        {/* 模块专属 MQTT Topic */}
                        {mod.id === 'voice' && (
                          <div className="space-y-2 pt-2 border-t">
                            <p className="text-sm font-medium">MQTT Topic</p>
                            <ApiRow method="PUB" path={`devices/${selectedDevice.id}/voice/up`} desc="语音文本上报" />
                            <ApiRow method="SUB" path={`devices/${selectedDevice.id}/voice/down`} desc="执行结果返回" />
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-2">
                              <div>
                                <p className="text-xs text-muted-foreground mb-1">上报 Payload</p>
                                <CodeBlock json={{ text: "打开灯" }} />
                              </div>
                              <div>
                                <p className="text-xs text-muted-foreground mb-1">返回 Payload</p>
                                <CodeBlock json={{ success: true, message: "已执行", action: "set_property" }} />
                              </div>
                            </div>
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  );
                })}
              </div>
            </TabsContent>
          )}

          {/* Tab 3: 接口详情 */}
          <TabsContent value="api">
            <div className="space-y-4">
              {/* MQTT 接口 */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">MQTT 接口</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <ApiRow
                    method="PUB"
                    path={`devices/${selectedDevice.id}/telemetry/up`}
                    desc="设备上报遥测数据（设备端发布）"
                  />
                  <ApiRow
                    method="PUB"
                    path={`devices/${selectedDevice.id}/event`}
                    desc="设备事件上报（设备端发布）"
                  />
                  <ApiRow
                    method="SUB"
                    path={`devices/${selectedDevice.id}/service/invoke`}
                    desc="服务调用下发（设备端订阅）"
                  />
                  <ApiRow
                    method="PUB"
                    path={`devices/${selectedDevice.id}/service/reply`}
                    desc="服务执行结果回传（设备端发布）"
                  />
                  {thingModel?.modules?.some((m) => m.id === 'voice') && (
                    <>
                      <div className="border-t my-2" />
                      <p className="text-xs text-muted-foreground px-3">语音模块</p>
                      <ApiRow
                        method="PUB"
                        path={`devices/${selectedDevice.id}/voice/up`}
                        desc="语音文本上报（设备端发布）"
                      />
                      <ApiRow
                        method="SUB"
                        path={`devices/${selectedDevice.id}/voice/down`}
                        desc="语音执行结果回传（设备端订阅）"
                      />
                    </>
                  )}
                </CardContent>
              </Card>

              {/* MQTT 认证信息 */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">MQTT 认证参数</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="bg-muted rounded-lg p-4 font-mono text-sm space-y-1">
                    <p><span className="text-muted-foreground">Host: </span>localhost</p>
                    <p><span className="text-muted-foreground">Port: </span>1883</p>
                    <p><span className="text-muted-foreground">Client ID: </span>{selectedDevice.id}</p>
                    <p><span className="text-muted-foreground">Username: </span>{selectedDevice.id}</p>
                    <p><span className="text-muted-foreground">Password: </span>{selectedDevice.device_secret}</p>
                  </div>
                </CardContent>
              </Card>

              {/* Payload 示例 */}
              {thingModel && properties.length > 0 && (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-base">上报 Payload 示例</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto">
                      {JSON.stringify(
                        Object.fromEntries(
                          properties.map((p) => [
                            p.id,
                            p.dataType === 'int' ? 0
                              : p.dataType === 'float' ? 0.0
                              : p.dataType === 'bool' ? false
                              : p.dataType === 'enum' ? (p.enumValues?.[0]?.value ?? 0)
                              : '',
                          ])
                        ),
                        null,
                        2
                      )}
                    </pre>
                  </CardContent>
                </Card>
              )}
            </div>
          </TabsContent>
        </Tabs>
      )}
    </div>
  );
}

/* ---- 子组件 ---- */

function PropertyCard({
  property,
  value,
  error,
  unknown,
}: {
  property: Property;
  value: unknown;
  error?: string;
  unknown?: boolean;
}) {
  const hasValue = value !== undefined;

  return (
    <Card className={error ? 'border-destructive/40' : ''}>
      <CardContent className="pt-4 pb-4 space-y-2">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium">{property.name}</span>
          {error ? (
            <Badge variant="destructive" className="text-xs">{unknown ? '未知属性' : '异常'}</Badge>
          ) : hasValue ? (
            <Badge variant="success" className="text-xs">正常</Badge>
          ) : (
            <Badge variant="secondary" className="text-xs">未上报</Badge>
          )}
        </div>
        <div className="flex items-baseline gap-1">
          <span className="text-2xl font-bold font-mono">
            {hasValue ? String(value) : '--'}
          </span>
          {property.unit && (
            <span className="text-sm text-muted-foreground">{property.unit}</span>
          )}
        </div>
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <span>{property.id} · {property.dataType}</span>
          {property.min !== undefined && property.max !== undefined && (
            <span>范围 {property.min} ~ {property.max}</span>
          )}
        </div>
        {error && (
          <p className="text-xs text-destructive">{error}</p>
        )}
      </CardContent>
    </Card>
  );
}

function ApiRow({ method, path, desc }: { method: string; path: string; desc: string }) {
  const colors: Record<string, string> = {
    GET: 'bg-emerald-500/15 text-emerald-400',
    PUT: 'bg-amber-500/15 text-amber-400',
    DELETE: 'bg-red-500/15 text-red-400',
    POST: 'bg-blue-500/15 text-blue-400',
    PUB: 'bg-violet-500/15 text-violet-400',
    SUB: 'bg-cyan-500/15 text-cyan-400',
  };

  return (
    <div className="flex items-center gap-3 py-2 px-3 rounded-lg bg-muted/50">
      <span className={`text-xs font-bold px-2 py-0.5 rounded ${colors[method] || 'bg-muted text-foreground'}`}>
        {method}
      </span>
      <code className="text-sm font-mono flex-1 break-all">{path}</code>
      <span className="text-xs text-muted-foreground shrink-0">{desc}</span>
    </div>
  );
}

function CodeBlock({ json }: { json: unknown }) {
  const [copied, setCopied] = useState(false);
  const text = JSON.stringify(json, null, 2);

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="relative group">
      <pre className="bg-muted rounded-lg p-4 text-sm overflow-x-auto font-mono">{text}</pre>
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-1.5 rounded-md bg-background/80 border opacity-0 group-hover:opacity-100 transition-opacity"
      >
        {copied ? <Check className="h-3.5 w-3.5 text-emerald-400" /> : <Copy className="h-3.5 w-3.5 text-muted-foreground" />}
      </button>
    </div>
  );
}

function sampleValue(p: Property): unknown {
  switch (p.dataType) {
    case 'int': return p.min !== undefined ? p.min : 0;
    case 'float': return p.min !== undefined ? p.min : 0.0;
    case 'bool': return false;
    case 'enum': return p.enumValues?.[0]?.value ?? 0;
    default: return '';
  }
}

function sampleValueByType(dataType: string): unknown {
  switch (dataType) {
    case 'int': return 0;
    case 'float': return 0.0;
    case 'bool': return false;
    default: return '';
  }
}
