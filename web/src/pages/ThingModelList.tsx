import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Plus, Pencil, Trash2, Mic } from 'lucide-react';
import { thingModelApi } from '../api';
import type { ThingModel } from '../api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose,
} from '@/components/ui/dialog';

export default function ThingModelList() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ThingModel[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [deleteTarget, setDeleteTarget] = useState<ThingModel | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await thingModelApi.list(page, pageSize);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取物模型列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await thingModelApi.delete(deleteTarget.id);
      toast.success('删除成功');
      setDeleteTarget(null);
      fetchData();
    } catch {
      toast.error('删除失败');
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle>物模型管理</CardTitle>
          <Button onClick={() => navigate('/thing-models/new')}>
            <Plus className="h-4 w-4" />
            新建物模型
          </Button>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无物模型</p>
              <Button variant="outline" className="mt-4" onClick={() => navigate('/thing-models/new')}>
                创建第一个物模型
              </Button>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>名称</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead>属性</TableHead>
                    <TableHead>事件</TableHead>
                    <TableHead>服务</TableHead>
                    <TableHead>模块</TableHead>
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
                          onClick={() => navigate(`/thing-models/${item.id}`)}
                        >
                          {item.name}
                        </button>
                      </TableCell>
                      <TableCell className="text-muted-foreground max-w-[200px] truncate">
                        {item.description || '-'}
                      </TableCell>
                      <TableCell><Badge variant="info">{item.properties?.length || 0} 个</Badge></TableCell>
                      <TableCell><Badge variant="warning">{item.events?.length || 0} 个</Badge></TableCell>
                      <TableCell><Badge variant="success">{item.services?.length || 0} 个</Badge></TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {item.modules?.length > 0 ? item.modules.map((m) => (
                            <Badge key={m.id} variant="outline" className="gap-1">
                              {m.id === 'voice' && <Mic className="h-3 w-3" />}
                              {m.id}
                            </Badge>
                          )) : <span className="text-muted-foreground text-sm">-</span>}
                        </div>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {new Date(item.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button variant="ghost" size="sm" onClick={() => navigate(`/thing-models/${item.id}/edit`)}>
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

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4 pt-4 border-t border-border">
                  <span className="text-sm text-muted-foreground">共 {total} 条</span>
                  <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                      上一页
                    </Button>
                    <span className="text-sm text-muted-foreground px-2">{page} / {totalPages}</span>
                    <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
                      下一页
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Delete confirmation dialog */}
      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除物模型「{deleteTarget?.name}」吗？此操作不可撤销。
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
