import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { ClipboardList, CheckCircle2, XCircle } from 'lucide-react';
import { scheduledTaskLogApi, scheduledTaskApi, deviceApi } from '../api';
import type { ScheduledTaskLog, ScheduledTask, Device } from '../api';
import { DataPagination } from '@/components/ui/data-pagination';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';

export default function ScheduledTaskLogList() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ScheduledTaskLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [devices, setDevices] = useState<Device[]>([]);
  const [tasks, setTasks] = useState<ScheduledTask[]>([]);
  const [filterDeviceId, setFilterDeviceId] = useState('');
  const [filterTaskId, setFilterTaskId] = useState('');
  const [payloadDialog, setPayloadDialog] = useState<unknown>(null);

  useEffect(() => {
    deviceApi.list(1, 100).then((res) => setDevices(res.data.list || [])).catch(() => {});
    scheduledTaskApi.list(1, 100).then((res) => setTasks(res.data.list || [])).catch(() => {});
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await scheduledTaskLogApi.list(
        page,
        pageSize,
        filterTaskId || undefined,
        filterDeviceId || undefined,
      );
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取执行历史失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setPage(1);
  }, [filterDeviceId, filterTaskId]);

  useEffect(() => {
    fetchData();
  }, [page, pageSize, filterDeviceId, filterTaskId]);


  const payloadPreview = (payload: unknown): string => {
    try {
      const str = JSON.stringify(payload);
      return str.length > 60 ? str.slice(0, 60) + '...' : str;
    } catch {
      return String(payload);
    }
  };

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-3">
            <ClipboardList className="h-5 w-5 text-primary" />
            <CardTitle>执行历史</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3 mb-4 flex-wrap">
            <span className="text-sm text-muted-foreground whitespace-nowrap">筛选设备</span>
            <Select value={filterDeviceId} onValueChange={(v) => setFilterDeviceId(v === '_all' ? '' : v)}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="全部设备" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部设备</SelectItem>
                {devices.map((d) => (
                  <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <span className="text-sm text-muted-foreground whitespace-nowrap">筛选任务</span>
            <Select value={filterTaskId} onValueChange={(v) => setFilterTaskId(v === '_all' ? '' : v)}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="全部任务" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部任务</SelectItem>
                {tasks.map((t) => (
                  <SelectItem key={t.id} value={t.id}>{t.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>

            <span className="text-sm text-muted-foreground ml-auto">共 {total} 条</span>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无执行记录</p>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>执行时间</TableHead>
                    <TableHead>任务名</TableHead>
                    <TableHead>设备</TableHead>
                    <TableHead>动作类型</TableHead>
                    <TableHead>下发 Topic</TableHead>
                    <TableHead>Payload</TableHead>
                    <TableHead>状态</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="text-muted-foreground text-sm whitespace-nowrap">
                        {new Date(item.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="font-medium">{item.task_name}</TableCell>
                      <TableCell>{item.device_name}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-xs">
                          {item.action_type === 'property_set' ? '属性设置' : '服务调用'}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground max-w-[200px] truncate">
                        {item.topic}
                      </TableCell>
                      <TableCell>
                        <button
                          onClick={() => setPayloadDialog(item.payload)}
                          className="font-mono text-xs text-muted-foreground hover:text-foreground underline-offset-2 hover:underline max-w-[160px] truncate block text-left"
                        >
                          {payloadPreview(item.payload)}
                        </button>
                      </TableCell>
                      <TableCell>
                        {item.status === 'success' ? (
                          <div className="flex items-center gap-1 text-green-500 text-xs">
                            <CheckCircle2 className="h-3.5 w-3.5" />
                            成功
                          </div>
                        ) : (
                          <div className="flex flex-col gap-0.5">
                            <div className="flex items-center gap-1 text-destructive text-xs">
                              <XCircle className="h-3.5 w-3.5" />
                              失败
                            </div>
                            {item.error && (
                              <span className="text-xs text-muted-foreground truncate max-w-[140px]" title={item.error}>
                                {item.error}
                              </span>
                            )}
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
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

      {/* Payload 详情弹窗 */}
      <Dialog open={payloadDialog !== null} onOpenChange={(open) => !open && setPayloadDialog(null)}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Payload 详情</DialogTitle>
          </DialogHeader>
          <pre className="text-xs bg-muted rounded-md p-4 overflow-auto max-h-96 whitespace-pre-wrap break-all">
            {JSON.stringify(payloadDialog, null, 2)}
          </pre>
        </DialogContent>
      </Dialog>
    </>
  );
}
