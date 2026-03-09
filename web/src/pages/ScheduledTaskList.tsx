import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Clock, Plus, Pencil, Trash2, Power } from 'lucide-react';
import { scheduledTaskApi, deviceApi } from '../api';
import type { ScheduledTask, Device } from '../api';
import { Card, CardContent } from '@/components/ui/card';
import { DataPagination } from '@/components/ui/data-pagination';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from '@/components/ui/dialog';

export default function ScheduledTaskList() {
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<ScheduledTask[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(true);
  const [filterDevice, setFilterDevice] = useState('all');
  const [deleteId, setDeleteId] = useState<string | null>(null);

  useEffect(() => { fetchDevices(); }, []);
  useEffect(() => { fetchTasks(); }, [page, pageSize, filterDevice]);

  const fetchDevices = async () => {
    try {
      const res = await deviceApi.list(1, 100);
      setDevices(res.data.list || []);
    } catch { /* silent */ }
  };

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const deviceId = filterDevice === 'all' ? undefined : filterDevice;
      const res = await scheduledTaskApi.list(page, pageSize, deviceId);
      setTasks(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取定时任务列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (task: ScheduledTask) => {
    try {
      await scheduledTaskApi.update(task.id, {
        device_id: task.device_id, name: task.name, cron_expr: task.cron_expr,
        action_type: task.action_type, property_id: task.property_id,
        service_id: task.service_id, value: task.value, enabled: !task.enabled,
      });
      toast.success(task.enabled ? '已禁用' : '已启用');
      fetchTasks();
    } catch { toast.error('操作失败'); }
  };

  const handleDelete = async () => {
    if (!deleteId) return;
    try {
      await scheduledTaskApi.delete(deleteId);
      toast.success('删除成功');
      setDeleteId(null);
      fetchTasks();
    } catch { toast.error('删除失败'); }
  };


  return (
    <>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Clock className="h-5 w-5 text-primary" />
            <h2 className="text-lg font-semibold">定时任务</h2>
            <Badge variant="secondary">{total} 条</Badge>
          </div>
          <div className="flex items-center gap-3">
            <Select value={filterDevice} onValueChange={(v) => { setFilterDevice(v); setPage(1); }}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="筛选设备" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部设备</SelectItem>
                {devices.map(d => (
                  <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button onClick={() => navigate('/scheduled-tasks/new')}>
              <Plus className="h-4 w-4 mr-1" /> 创建任务
            </Button>
          </div>
        </div>

        {loading ? (
          <div className="flex justify-center py-20"><Spinner size="lg" /></div>
        ) : tasks.length === 0 ? (
          <Card><CardContent className="py-16 text-center text-muted-foreground">暂无定时任务</CardContent></Card>
        ) : (
          <div className="rounded-lg border overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted/50 text-muted-foreground">
                  <th className="text-left px-4 py-3 font-medium">任务名称</th>
                  <th className="text-left px-4 py-3 font-medium">设备</th>
                  <th className="text-left px-4 py-3 font-medium">Cron</th>
                  <th className="text-left px-4 py-3 font-medium">动作</th>
                  <th className="text-left px-4 py-3 font-medium">状态</th>
                  <th className="text-left px-4 py-3 font-medium">最后执行</th>
                  <th className="text-right px-4 py-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map(t => (
                  <tr key={t.id} className="border-t">
                    <td className="px-4 py-3 font-medium">{t.name}</td>
                    <td className="px-4 py-3 text-muted-foreground">{t.device_name}</td>
                    <td className="px-4 py-3 font-mono text-xs">{t.cron_expr}</td>
                    <td className="px-4 py-3">
                      <Badge variant="outline" className="text-xs">
                        {t.action_type === 'property_set' ? '属性设置' : '服务调用'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant={t.enabled ? 'default' : 'secondary'} className="text-xs">
                        {t.enabled ? '启用' : '禁用'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {t.last_run_at ? new Date(t.last_run_at).toLocaleString() : '未执行'}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button variant="ghost" size="icon" onClick={() => handleToggle(t)}
                          title={t.enabled ? '禁用' : '启用'}>
                          <Power className={`h-4 w-4 ${t.enabled ? 'text-emerald-400' : 'text-muted-foreground'}`} />
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => navigate(`/scheduled-tasks/${t.id}/edit`)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" onClick={() => setDeleteId(t.id)}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            <DataPagination
              page={page} pageSize={pageSize} total={total}
              onPageChange={setPage}
              onPageSizeChange={(s) => setPageSize(s)}
            />
          </div>
        )}
      </div>

      <Dialog open={!!deleteId} onOpenChange={(open: boolean) => !open && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>删除后不可恢复，确定要删除这个定时任务吗？</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>取消</Button>
            <Button variant="destructive" onClick={handleDelete}>删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
