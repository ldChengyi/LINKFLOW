import { useEffect, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { Radio, Plus, XCircle } from 'lucide-react';
import { otaTaskApi, firmwareApi, deviceApi } from '../api';
import type { OTATask, Firmware, Device } from '../api';
import { onWSMessage } from './Dashboard';
import { Button } from '@/components/ui/button';
import { DataPagination } from '@/components/ui/data-pagination';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';

const statusMap: Record<string, { label: string; color: string }> = {
  pending:      { label: '等待中', color: 'bg-gray-500' },
  pushing:      { label: '推送中', color: 'bg-blue-500' },
  downloading:  { label: '下载中', color: 'bg-cyan-500' },
  installing:   { label: '安装中', color: 'bg-yellow-500' },
  completed:    { label: '已完成', color: 'bg-emerald-500' },
  failed:       { label: '失败',   color: 'bg-red-500' },
  cancelled:    { label: '已取消', color: 'bg-gray-400' },
};

export default function OTATaskList() {
  const [data, setData] = useState<OTATask[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [filterDevice, setFilterDevice] = useState('');
  const [devices, setDevices] = useState<Device[]>([]);
  const [firmwares, setFirmwares] = useState<Firmware[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [selDevice, setSelDevice] = useState('');
  const [selFirmware, setSelFirmware] = useState('');
  const [creating, setCreating] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<OTATask | null>(null);
  const [pageSize, setPageSize] = useState(10);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await otaTaskApi.list(page, pageSize, filterDevice || undefined);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch { toast.error('获取OTA任务失败'); }
    finally { setLoading(false); }
  }, [page, pageSize, filterDevice]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    deviceApi.list(1, 200).then(r => setDevices(r.data.list || [])).catch(() => {});
    firmwareApi.list(1, 200).then(r => setFirmwares(r.data.list || [])).catch(() => {});
  }, []);

  // WS real-time OTA progress
  useEffect(() => {
    return onWSMessage((msg) => {
      if (msg.type === 'ota_progress') {
        const d = msg.data;
        setData(prev => prev.map(t =>
          t.id === d.task_id ? { ...t, status: d.status, progress: d.progress, error_msg: d.error_msg || t.error_msg } : t
        ));
      }
    });
  }, []);

  const handleCreate = async () => {
    if (!selDevice || !selFirmware) { toast.error('请选择设备和固件'); return; }
    setCreating(true);
    try {
      await otaTaskApi.create({ device_id: selDevice, firmware_id: selFirmware });
      toast.success('OTA任务创建成功');
      setCreateOpen(false);
      setSelDevice(''); setSelFirmware('');
      fetchData();
    } catch { toast.error('创建失败'); }
    finally { setCreating(false); }
  };

  const handleCancel = async () => {
    if (!cancelTarget) return;
    try {
      await otaTaskApi.cancel(cancelTarget.id);
      toast.success('已取消');
      setCancelTarget(null);
      fetchData();
    } catch { toast.error('取消失败'); }
  };

  const canCancel = (s: string) => s === 'pending' || s === 'pushing';

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-3">
            <Radio className="h-5 w-5 text-primary" />
            <CardTitle>OTA升级</CardTitle>
          </div>
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" /> 创建升级任务
          </Button>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4 mb-4">
            <Select value={filterDevice} onValueChange={(v) => { setFilterDevice(v === 'all' ? '' : v); setPage(1); }}>
              <SelectTrigger className="w-[220px]"><SelectValue placeholder="全部设备" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部设备</SelectItem>
                {devices.map(d => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
              </SelectContent>
            </Select>
            <span className="text-sm text-muted-foreground ml-auto">共 {total} 个任务</span>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无OTA任务</p>
              <Button variant="outline" className="mt-4" onClick={() => setCreateOpen(true)}>创建第一个升级任务</Button>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>设备</TableHead>
                    <TableHead>目标版本</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>进度</TableHead>
                    <TableHead>错误信息</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((t) => {
                    const st = statusMap[t.status] || statusMap.pending;
                    return (
                      <TableRow key={t.id}>
                        <TableCell className="font-medium">{t.device_name}</TableCell>
                        <TableCell><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{t.firmware_version}</code></TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center gap-1.5 text-xs font-medium text-white px-2 py-0.5 rounded-full ${st.color}`}>
                            {st.label}
                          </span>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2 min-w-[120px]">
                            <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                              <div className={`h-full rounded-full transition-all ${t.status === 'failed' ? 'bg-red-500' : 'bg-primary'}`} style={{ width: `${t.progress}%` }} />
                            </div>
                            <span className="text-xs text-muted-foreground w-8">{t.progress}%</span>
                          </div>
                        </TableCell>
                        <TableCell className="max-w-[200px] truncate text-xs text-muted-foreground">{t.error_msg || '-'}</TableCell>
                        <TableCell>{new Date(t.created_at).toLocaleString()}</TableCell>
                        <TableCell className="text-right">
                          {canCancel(t.status) && (
                            <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setCancelTarget(t)}>
                              <XCircle className="h-4 w-4" />
                            </Button>
                          )}
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>

              <DataPagination
                page={page} pageSize={pageSize} total={total}
                onPageChange={setPage}
                onPageSizeChange={(s) => setPageSize(s)}
              />
            </>
          )}
        </CardContent>
      </Card>

      {/* 创建OTA任务 Dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => {
        setCreateOpen(open);
        if (!open) { setSelDevice(''); setSelFirmware(''); }
      }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>创建OTA升级任务</DialogTitle></DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">目标设备 *</label>
              <Select value={selDevice} onValueChange={setSelDevice}>
                <SelectTrigger><SelectValue placeholder="选择设备" /></SelectTrigger>
                <SelectContent>
                  {devices.map(d => <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">目标固件 *</label>
              <Select value={selFirmware} onValueChange={setSelFirmware}>
                <SelectTrigger><SelectValue placeholder="选择固件版本" /></SelectTrigger>
                <SelectContent>
                  {firmwares.map(f => <SelectItem key={f.id} value={f.id}>{f.name} v{f.version}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            <div className="flex gap-3 pt-2">
              <Button onClick={handleCreate} disabled={creating} className="flex-1">
                {creating ? '创建中...' : '确认创建'}
              </Button>
              <Button variant="outline" onClick={() => setCreateOpen(false)} className="flex-1">取消</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 取消确认 */}
      <Dialog open={!!cancelTarget} onOpenChange={(open) => !open && setCancelTarget(null)}>
        <DialogContent>
          <DialogHeader><DialogTitle>确认取消</DialogTitle></DialogHeader>
          <p className="text-sm text-muted-foreground">确定要取消设备「{cancelTarget?.device_name}」的OTA升级任务吗？</p>
          <DialogFooter>
            <DialogClose asChild><Button variant="outline">返回</Button></DialogClose>
            <Button variant="destructive" onClick={handleCancel}>取消任务</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
