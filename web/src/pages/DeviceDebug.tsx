import { useEffect, useState, useRef } from 'react';
import { toast } from 'sonner';
import { Terminal, Send, Power, Wifi, Radio, History, RefreshCw, Mic } from 'lucide-react';
import { deviceApi, thingModelApi } from '../api';
import type { Device, ThingModel, Property, Service, DeviceLatestData, DebugLog } from '../api';
import { onWSMessage } from './Dashboard';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { DataPagination } from '@/components/ui/data-pagination';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';

type ConnectionType = 'real' | 'simulated' | 'offline';

export default function DeviceDebug() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [selectedId, setSelectedId] = useState('');
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null);
  const [thingModel, setThingModel] = useState<ThingModel | null>(null);
  const [latestData, setLatestData] = useState<DeviceLatestData | null>(null);
  const [loading, setLoading] = useState(false);
  const [propValues, setPropValues] = useState<Record<string, unknown>>({});
  const [svcParams, setSvcParams] = useState<Record<string, string>>({});
  const [sending, setSending] = useState<string | null>(null);
  const [simulating, setSimulating] = useState(false);
  const [connType, setConnType] = useState<ConnectionType>('offline');
  const [debugLogs, setDebugLogs] = useState<DebugLog[]>([]);
  const [loadingLogs, setLoadingLogs] = useState(false);
  const [logsTotal, setLogsTotal] = useState(0);
  const [logsPage, setLogsPage] = useState(1);
  const [logsPageSize, setLogsPageSize] = useState(20);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const [voiceText, setVoiceText] = useState('');
  const [voiceSending, setVoiceSending] = useState(false);
  const [voiceResults, setVoiceResults] = useState<Array<{ text: string; result: { success: boolean; message: string; action?: string; audio_url?: string }; time: string }>>([]);

  // 模拟上线心跳：每 2 分钟续期，停止时清理
  useEffect(() => {
    if (heartbeatRef.current) {
      clearInterval(heartbeatRef.current);
      heartbeatRef.current = null;
    }
    if (connType === 'simulated' && selectedId) {
      heartbeatRef.current = setInterval(() => {
        deviceApi.simulateHeartbeat(selectedId).catch(() => {
          // 续期失败（已过期），更新状态
          setConnType('offline');
          setSelectedDevice((prev) => prev ? { ...prev, status: 'offline' } : prev);
        });
      }, 2 * 60 * 1000);
    }
    return () => {
      if (heartbeatRef.current) {
        clearInterval(heartbeatRef.current);
        heartbeatRef.current = null;
      }
    };
  }, [connType, selectedId]);

  useEffect(() => { fetchDevices(); }, []);
  useEffect(() => { if (selectedId) fetchDetail(selectedId); }, [selectedId]);

  // WS real-time telemetry updates
  useEffect(() => {
    return onWSMessage((msg) => {
      if (msg.type === 'telemetry' && msg.data?.device_id === selectedId) {
        setLatestData((prev) => ({
          time: msg.data.time || new Date().toISOString(),
          payload: { ...(prev?.payload || {}), ...msg.data.payload },
          valid: msg.data.valid ?? true,
        }));
      }
      if (msg.type === 'device_status' && msg.data?.device_id === selectedId) {
        setSelectedDevice((prev) => prev ? { ...prev, status: msg.data.status } : prev);
        // 设备状态变化时重新查询连接类型
        fetchConnectionType(selectedId);
      }
      if (msg.type === 'voice_down' && msg.data?.device_id === selectedId) {
        setVoiceResults((prev) => [{
          text: msg.data.text || '',
          result: { success: msg.data.success, message: msg.data.message, action: msg.data.action, audio_url: msg.data.audio_url },
          time: new Date().toISOString(),
        }, ...prev].slice(0, 20));
      }
    });
  }, [selectedId]);

  const fetchDevices = async () => {
    try {
      const res = await deviceApi.list(1, 100);
      setDevices(res.data.list || []);
    } catch { toast.error('获取设备列表失败'); }
  };

  const fetchConnectionType = async (id: string) => {
    try {
      const res = await deviceApi.connectionType(id);
      setConnType(res.data.connection_type);
    } catch {
      setConnType('offline');
    }
  };

  const fetchDetail = async (id: string) => {
    setLoading(true);
    setThingModel(null);
    setLatestData(null);
    setPropValues({});
    setSvcParams({});
    setConnType('offline');
    setDebugLogs([]);
    setLogsPage(1);
    setVoiceText('');
    setVoiceResults([]);
    try {
      const [devRes, dataRes, connRes] = await Promise.all([
        deviceApi.get(id),
        deviceApi.latestData(id),
        deviceApi.connectionType(id),
      ]);
      setSelectedDevice(devRes.data);
      setLatestData(dataRes.data);
      setConnType(connRes.data.connection_type);
      if (devRes.data.model_id) {
        try {
          const mRes = await thingModelApi.get(devRes.data.model_id);
          setThingModel(mRes.data);
        } catch { /* silent */ }
      }
      fetchDebugLogs(id);
    } catch { toast.error('获取设备信息失败'); }
    finally { setLoading(false); }
  };

  const fetchDebugLogs = async (id: string) => {
    setLoadingLogs(true);
    try {
      const res = await deviceApi.debugLogs(id, logsPage, logsPageSize);
      setDebugLogs(res.data.list || []);
      setLogsTotal(res.data.total);
    } catch { /* silent */ }
    finally { setLoadingLogs(false); }
  };

  useEffect(() => {
    if (selectedId) fetchDebugLogs(selectedId);
  }, [logsPage, logsPageSize]);

  const rwProps = thingModel?.properties?.filter((p) => p.accessMode === 'rw') || [];
  const services = thingModel?.services || [];
  const hasVoiceModule = thingModel?.modules?.some((m) => m.id === 'voice') || false;
  const isOnline = connType !== 'offline';
  const isReal = connType === 'real';

  const handleAllPropertySet = async () => {
    setSending('__all__');
    try {
      const properties: Record<string, unknown> = {};
      for (const prop of rwProps) {
        const val = propValues[prop.id];
        if (prop.dataType === 'bool') {
          properties[prop.id] = val ?? false;
        } else if (val !== undefined && String(val) !== '') {
          if (prop.dataType === 'int') properties[prop.id] = parseInt(String(val), 10);
          else if (prop.dataType === 'float') properties[prop.id] = parseFloat(String(val));
          else if (prop.dataType === 'enum') properties[prop.id] = parseInt(String(val), 10);
          else properties[prop.id] = val;
        }
      }
      if (Object.keys(properties).length === 0) {
        toast.error('请至少设置一个属性值');
        return;
      }
      await deviceApi.debug(selectedId, { action_type: 'property_set', properties });
      toast.success('属性下发成功');
      fetchDebugLogs(selectedId);
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { msg?: string } } })?.response?.data?.msg || '下发失败';
      toast.error(msg);
    } finally { setSending(null); }
  };

  const handleServiceInvoke = async (svc: Service) => {
    setSending(svc.id);
    try {
      let params: unknown = {};
      if (svcParams[svc.id]) {
        try { params = JSON.parse(svcParams[svc.id]); } catch { toast.error('参数 JSON 格式错误'); setSending(null); return; }
      }
      await deviceApi.debug(selectedId, { action_type: 'service_invoke', service_id: svc.id, value: params });
      toast.success(`${svc.name} 调用成功`);
      fetchDebugLogs(selectedId);
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { msg?: string } } })?.response?.data?.msg || '调用失败';
      toast.error(msg);
    } finally { setSending(null); }
  };

  const handleSimulateToggle = async () => {
    if (!selectedId) return;
    setSimulating(true);
    try {
      if (isOnline) {
        await deviceApi.simulateOffline(selectedId);
        setSelectedDevice((prev) => prev ? { ...prev, status: 'offline' } : prev);
        setConnType('offline');
        toast.success('已模拟下线');
      } else {
        await deviceApi.simulateOnline(selectedId);
        setSelectedDevice((prev) => prev ? { ...prev, status: 'online' } : prev);
        setConnType('simulated');
        toast.success('已模拟上线');
      }
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { msg?: string } } })?.response?.data?.msg || '模拟操作失败';
      toast.error(msg);
    }
    finally { setSimulating(false); }
  };

  const handleVoiceSend = async () => {
    if (!voiceText.trim()) {
      toast.error('请输入语音指令文本');
      return;
    }
    setVoiceSending(true);
    const text = voiceText.trim();
    try {
      const res = await deviceApi.voiceDebug(selectedId, text);
      const d = res.data;
      setVoiceResults((prev) => [{
        text,
        result: { success: d.success, message: d.message, action: d.action, audio_url: d.audio_url },
        time: new Date().toISOString(),
      }, ...prev].slice(0, 20));
      setVoiceText('');
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { msg?: string } } })?.response?.data?.msg || '发送失败';
      toast.error(msg);
    } finally {
      setVoiceSending(false);
    }
  };

  const renderConnBadge = () => {
    switch (connType) {
      case 'real':
        return <Badge variant="default" className="gap-1"><Wifi className="h-3 w-3" />MQTT 连接</Badge>;
      case 'simulated':
        return <Badge variant="secondary" className="gap-1"><Radio className="h-3 w-3" />模拟在线</Badge>;
      default:
        return <Badge variant="outline">离线</Badge>;
    }
  };

  const renderPropInput = (prop: Property) => {
    const currentVal = latestData?.payload?.[prop.id];
    const currentLabel = currentVal !== undefined ? String(currentVal) : '--';

    switch (prop.dataType) {
      case 'bool':
        return (
          <div className="flex items-center gap-3">
            <span className="text-xs text-muted-foreground">当前: {currentLabel}</span>
            <Switch
              checked={propValues[prop.id] as boolean ?? (currentVal as boolean ?? false)}
              onCheckedChange={(v) => setPropValues((p) => ({ ...p, [prop.id]: v }))}
            />
          </div>
        );
      case 'enum':
        return (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground shrink-0">当前: {
              prop.enumValues?.find((e) => e.value === currentVal)?.label ?? currentLabel
            }</span>
            <Select value={String(propValues[prop.id] ?? '')} onValueChange={(v) => setPropValues((p) => ({ ...p, [prop.id]: v }))}>
              <SelectTrigger className="w-32 h-8"><SelectValue placeholder="选择" /></SelectTrigger>
              <SelectContent>
                {prop.enumValues?.map((e) => (
                  <SelectItem key={e.value} value={String(e.value)}>{e.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        );
      default:
        return (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground shrink-0">当前: {currentLabel}{prop.unit ? ` ${prop.unit}` : ''}</span>
            <Input
              type={prop.dataType === 'string' ? 'text' : 'number'}
              className="w-32 h-8"
              placeholder={prop.min !== undefined ? `${prop.min}~${prop.max}` : '输入值'}
              value={String(propValues[prop.id] ?? '')}
              onChange={(e) => setPropValues((p) => ({ ...p, [prop.id]: e.target.value }))}
            />
          </div>
        );
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-3">
            <Terminal className="h-5 w-5 text-primary" />
            <CardTitle>在线调试</CardTitle>
          </div>
          <div className="flex items-center gap-3">
            <Select value={selectedId} onValueChange={setSelectedId}>
              <SelectTrigger className="w-56"><SelectValue placeholder="选择设备" /></SelectTrigger>
              <SelectContent>
                {devices.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    <span className="flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full ${d.status === 'online' ? 'bg-emerald-500' : 'bg-gray-400'}`} />
                      {d.name}
                    </span>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedDevice && (
              <>
                {renderConnBadge()}
                <Button
                  size="sm"
                  variant={isOnline ? 'outline' : 'default'}
                  onClick={handleSimulateToggle}
                  disabled={simulating || isReal}
                  title={isReal ? '真实 MQTT 连接的设备不能模拟上下线' : ''}
                >
                  <Power className="h-4 w-4 mr-1" />
                  {simulating ? '处理中...' : isOnline ? '模拟下线' : '模拟上线'}
                </Button>
              </>
            )}
          </div>
        </CardHeader>
      </Card>

      {loading && <div className="flex justify-center py-12"><Spinner /></div>}

      {selectedDevice && !loading && !isOnline && (
        <Card><CardContent className="py-8 text-center text-muted-foreground">设备离线，请先点击「模拟上线」后再进行调试</CardContent></Card>
      )}

      {selectedDevice && !loading && isOnline && !thingModel && (
        <Card><CardContent className="py-8 text-center text-muted-foreground">该设备未绑定物模型，无法进行调试</CardContent></Card>
      )}

      {selectedDevice && !loading && isOnline && thingModel && (
        <Tabs defaultValue="properties" className="space-y-4">
          <TabsList>
            <TabsTrigger value="properties">属性设置</TabsTrigger>
            <TabsTrigger value="services">服务调用</TabsTrigger>
            <TabsTrigger value="history">调试历史</TabsTrigger>
            {hasVoiceModule && <TabsTrigger value="voice">语音调试</TabsTrigger>}
          </TabsList>

          {isReal && (
            <Card>
              <CardContent className="py-3 text-sm text-muted-foreground flex items-center gap-2">
                <Wifi className="h-4 w-4" />
                设备通过 MQTT 真实连接，下发指令后由设备硬件回传数据
              </CardContent>
            </Card>
          )}

          <TabsContent value="properties" className="space-y-4">
            {rwProps.length > 0 ? (
              <Card>
                <CardHeader><CardTitle className="text-base">属性设置</CardTitle></CardHeader>
                <CardContent className="space-y-3">
                  {rwProps.map((prop) => (
                    <div key={prop.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                      <div className="flex items-center gap-2 min-w-[120px]">
                        <span className="font-medium text-sm">{prop.name}</span>
                        <Badge variant="outline" className="text-xs">{prop.dataType}</Badge>
                      </div>
                      <div className="flex items-center gap-3">
                        {renderPropInput(prop)}
                      </div>
                    </div>
                  ))}
                  <div className="pt-2 flex justify-end">
                    <Button onClick={handleAllPropertySet} disabled={sending === '__all__'}>
                      {sending === '__all__' ? <Spinner className="h-4 w-4 mr-2" /> : <Send className="h-4 w-4 mr-2" />}
                      下发
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ) : (
              <Card><CardContent className="py-8 text-center text-muted-foreground">该物模型没有可写属性</CardContent></Card>
            )}
          </TabsContent>

          <TabsContent value="services" className="space-y-4">
            {services.length > 0 ? (
              <Card>
                <CardHeader><CardTitle className="text-base">服务调用</CardTitle></CardHeader>
                <CardContent className="space-y-3">
                  {services.map((svc) => (
                    <div key={svc.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                      <div className="min-w-[120px]">
                        <span className="font-medium text-sm">{svc.name}</span>
                        {svc.inputParams && svc.inputParams.length > 0 && (
                          <span className="text-xs text-muted-foreground ml-2">
                            参数: {svc.inputParams.map((p) => p.id).join(', ')}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-3">
                        {svc.inputParams && svc.inputParams.length > 0 && (
                          <Input
                            className="w-48 h-8 font-mono text-xs"
                            placeholder='{"key": "value"}'
                            value={svcParams[svc.id] || ''}
                            onChange={(e) => setSvcParams((p) => ({ ...p, [svc.id]: e.target.value }))}
                          />
                        )}
                        <Button size="sm" onClick={() => handleServiceInvoke(svc)} disabled={sending === svc.id}>
                          {sending === svc.id ? <Spinner className="h-4 w-4" /> : <Send className="h-4 w-4" />}
                          <span className="ml-1">调用</span>
                        </Button>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            ) : (
              <Card><CardContent className="py-8 text-center text-muted-foreground">该物模型没有可调用服务</CardContent></Card>
            )}
          </TabsContent>

          <TabsContent value="history" className="space-y-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                <CardTitle className="text-base flex items-center gap-2">
                  <History className="h-4 w-4" />
                  调试历史
                </CardTitle>
                <Button size="sm" variant="outline" onClick={() => fetchDebugLogs(selectedId)} disabled={loadingLogs}>
                  <RefreshCw className={`h-4 w-4 ${loadingLogs ? 'animate-spin' : ''}`} />
                </Button>
              </CardHeader>
              <CardContent>
                {loadingLogs ? (
                  <div className="flex justify-center py-8"><Spinner /></div>
                ) : debugLogs.length === 0 ? (
                  <div className="py-8 text-center text-muted-foreground text-sm">暂无调试记录</div>
                ) : (
                  <div className="space-y-2 max-h-96 overflow-y-auto">
                    {debugLogs.map((log) => (
                      <div key={log.id} className="border rounded-lg p-3 text-sm space-y-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <Badge variant={log.connection_type === 'real' ? 'default' : 'secondary'} className="gap-1">
                              {log.connection_type === 'real' ? <Wifi className="h-3 w-3" /> : <Radio className="h-3 w-3" />}
                              {log.connection_type === 'real' ? 'MQTT' : '模拟'}
                            </Badge>
                            <Badge variant={log.action_type === 'property_set' ? 'outline' : 'secondary'}>
                              {log.action_type === 'property_set' ? '属性设置' : log.action_type === 'voice_command' ? '语音指令' : '服务调用'}
                            </Badge>
                            <Badge variant={log.success ? 'default' : 'destructive'}>
                              {log.success ? '成功' : '失败'}
                            </Badge>
                          </div>
                          <span className="text-xs text-muted-foreground">
                            {new Date(log.created_at).toLocaleString('zh-CN')}
                          </span>
                        </div>
                        <div className="text-xs space-y-1">
                          <div className="font-mono bg-muted p-2 rounded">
                            <span className="text-muted-foreground">请求: </span>
                            {JSON.stringify(log.request)}
                          </div>
                          {log.response && (
                            <div className="font-mono bg-muted p-2 rounded">
                              <span className="text-muted-foreground">响应: </span>
                              {JSON.stringify(log.response)}
                            </div>
                          )}
                          {log.error_message && (
                            <div className="text-destructive">{log.error_message}</div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {logsTotal > 0 && (
                  <DataPagination
                    page={logsPage}
                    pageSize={logsPageSize}
                    total={logsTotal}
                    onPageChange={setLogsPage}
                    onPageSizeChange={setLogsPageSize}
                  />
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {hasVoiceModule && (
            <TabsContent value="voice" className="space-y-4">
              <Card>
                <CardHeader><CardTitle className="text-base flex items-center gap-2"><Mic className="h-4 w-4" />语音指令调试</CardTitle></CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center gap-3">
                    <Input
                      className="flex-1"
                      placeholder="输入语音指令文本，如：打开灯、调高亮度、设置温度到25"
                      value={voiceText}
                      onChange={(e) => setVoiceText(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter' && !voiceSending) handleVoiceSend(); }}
                    />
                    <Button onClick={handleVoiceSend} disabled={voiceSending || !voiceText.trim()}>
                      {voiceSending ? <Spinner className="h-4 w-4 mr-2" /> : <Send className="h-4 w-4 mr-2" />}
                      发送
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    通过 voice/up topic 发送文本指令，经语音处理链路（本地 NLP / Dify）解析并执行，结果通过 voice/down 实时返回
                  </p>
                </CardContent>
              </Card>

              {voiceResults.length > 0 && (
                <Card>
                  <CardHeader><CardTitle className="text-base">执行结果</CardTitle></CardHeader>
                  <CardContent>
                    <div className="space-y-2 max-h-96 overflow-y-auto">
                      {voiceResults.map((item, idx) => (
                        <div key={idx} className="border rounded-lg p-3 text-sm space-y-1">
                          <div className="flex items-center justify-between">
                            <Badge variant={item.result.success ? 'default' : 'destructive'}>
                              {item.result.success ? '成功' : '失败'}
                            </Badge>
                            <span className="text-xs text-muted-foreground">
                              {new Date(item.time).toLocaleString('zh-CN')}
                            </span>
                          </div>
                          <div className="text-muted-foreground">{item.result.message}</div>
                          {item.result.action && (
                            <div className="font-mono text-xs bg-muted p-2 rounded">{item.result.action}</div>
                          )}
                          {item.result.audio_url && (
                            <audio controls src={item.result.audio_url} className="w-full h-8 mt-1" />
                          )}
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>
          )}
        </Tabs>
      )}
    </div>
  );
}
