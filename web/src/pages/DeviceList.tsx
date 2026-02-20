import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Plus, Pencil, Trash2 } from 'lucide-react';
import { deviceApi } from '../api';
import type { Device } from '../api';
import { onWSMessage } from './Dashboard';
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

export default function DeviceList() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Device[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [deleteTarget, setDeleteTarget] = useState<Device | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await deviceApi.list(page, pageSize);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取设备列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  // 实时设备状态更新
  useEffect(() => {
    return onWSMessage((msg) => {
      if (msg.type === 'device_status') {
        const { device_id, status } = msg.data;
        setData((prev) => prev.map((d) => d.id === device_id ? { ...d, status } : d));
      }
    });
  }, []);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deviceApi.delete(deleteTarget.id);
      toast.success('删除成功');
      setDeleteTarget(null);
      fetchData();
    } catch {
      toast.error('删除失败');
    }
  };

  const maskSecret = (secret: string) => {
    if (secret.length <= 8) return secret;
    return secret.slice(0, 4) + '****' + secret.slice(-4);
  };


  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle>设备管理</CardTitle>
          <Button onClick={() => navigate('/devices/new')}>
            <Plus className="h-4 w-4" />
            新建设备
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无设备</p>
              <Button variant="outline" className="mt-4" onClick={() => navigate('/devices/new')}>
                添加第一个设备
              </Button>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>名称</TableHead>
                    <TableHead>绑定物模型</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>设备密钥</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <button
                          className="text-primary hover:underline cursor-pointer font-medium"
                          onClick={() => navigate(`/devices/${item.id}`)}
                        >
                          {item.name}
                        </button>
                      </TableCell>
                      <TableCell>
                        {item.model_name || <span className="text-muted-foreground">未绑定</span>}
                      </TableCell>
                      <TableCell>
                        <Badge variant={item.status === 'online' ? 'success' : 'secondary'}>
                          {item.status === 'online' ? '在线' : '离线'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                          {maskSecret(item.device_secret)}
                        </code>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {new Date(item.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm" onClick={() => navigate(`/devices/${item.id}/edit`)}>
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
              确定要删除设备「{deleteTarget?.name}」吗？此操作不可撤销。
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
