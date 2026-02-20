import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Plus, Pencil, Trash2, Bell } from 'lucide-react';
import { alertRuleApi, deviceApi } from '../api';
import type { AlertRule, Device } from '../api';
import { Button } from '@/components/ui/button';
import { DataPagination } from '@/components/ui/data-pagination';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose,
} from '@/components/ui/dialog';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';

const severityColors: Record<string, string> = {
  critical: 'destructive',
  warning: 'warning',
  info: 'secondary',
};

const severityLabels: Record<string, string> = {
  critical: '严重',
  warning: '警告',
  info: '信息',
};

export default function AlertRuleList() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AlertRule[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [deleteTarget, setDeleteTarget] = useState<AlertRule | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [filterDeviceId, setFilterDeviceId] = useState('');

  useEffect(() => {
    deviceApi.list(1, 100).then((res) => setDevices(res.data.list || [])).catch(() => {});
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await alertRuleApi.list(page, pageSize, filterDeviceId || undefined);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取告警规则失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setPage(1);
  }, [filterDeviceId]);

  useEffect(() => {
    fetchData();
  }, [page, pageSize, filterDeviceId]);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await alertRuleApi.delete(deleteTarget.id);
      toast.success('删除成功');
      setDeleteTarget(null);
      fetchData();
    } catch {
      toast.error('删除失败');
    }
  };


  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <div className="flex items-center gap-3">
            <Bell className="h-5 w-5 text-primary" />
            <CardTitle>告警规则</CardTitle>
          </div>
          <Button onClick={() => navigate('/alert-rules/new')}>
            <Plus className="h-4 w-4" />
            新建规则
          </Button>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3 mb-4">
            <span className="text-sm text-muted-foreground whitespace-nowrap">筛选设备</span>
            <Select value={filterDeviceId} onValueChange={(v) => setFilterDeviceId(v === '_all' ? '' : v)}>
              <SelectTrigger className="w-[220px]">
                <SelectValue placeholder="全部设备" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部设备</SelectItem>
                {devices.map((d) => (
                  <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <span className="text-sm text-muted-foreground ml-auto">共 {total} 条</span>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无告警规则</p>
              <Button variant="outline" className="mt-4" onClick={() => navigate('/alert-rules/new')}>
                创建第一条规则
              </Button>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>规则名称</TableHead>
                    <TableHead>设备</TableHead>
                    <TableHead>属性</TableHead>
                    <TableHead>条件</TableHead>
                    <TableHead>级别</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell>{item.device_name || item.device_id.slice(0, 8)}</TableCell>
                      <TableCell><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{item.property_id}</code></TableCell>
                      <TableCell className="font-mono text-sm">{item.operator} {item.threshold}</TableCell>
                      <TableCell>
                        <Badge variant={severityColors[item.severity] as any || 'secondary'}>
                          {severityLabels[item.severity] || item.severity}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={item.enabled ? 'success' : 'secondary'}>
                          {item.enabled ? '启用' : '禁用'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm" onClick={() => navigate(`/alert-rules/${item.id}/edit`)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setDeleteTarget(item)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
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

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除告警规则「{deleteTarget?.name}」吗？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">取消</Button>
            </DialogClose>
            <Button variant="destructive" onClick={handleDelete}>删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
