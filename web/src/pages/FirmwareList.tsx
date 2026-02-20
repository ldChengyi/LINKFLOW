import { useEffect, useState, useRef } from 'react';
import { toast } from 'sonner';
import { Upload, Trash2, Package } from 'lucide-react';
import { firmwareApi } from '../api';
import type { Firmware } from '../api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, DialogClose,
} from '@/components/ui/dialog';

function formatSize(bytes: number) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

export default function FirmwareList() {
  const [data, setData] = useState<Firmware[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Firmware | null>(null);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [name, setName] = useState('');
  const [version, setVersion] = useState('');
  const [description, setDescription] = useState('');
  const fileRef = useRef<HTMLInputElement>(null);
  const pageSize = 10;

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await firmwareApi.list(page, pageSize);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取固件列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, [page]);

  const handleUpload = async () => {
    const file = fileRef.current?.files?.[0];
    if (!name || !version || !file) {
      toast.error('请填写名称、版本号并选择文件');
      return;
    }
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append('name', name);
      fd.append('version', version);
      fd.append('description', description);
      fd.append('file', file);
      await firmwareApi.upload(fd);
      toast.success('上传成功');
      setUploadOpen(false);
      setName(''); setVersion(''); setDescription('');
      if (fileRef.current) fileRef.current.value = '';
      fetchData();
    } catch {
      toast.error('上传失败');
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await firmwareApi.delete(deleteTarget.id);
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
          <div className="flex items-center gap-3">
            <Package className="h-5 w-5 text-primary" />
            <CardTitle>固件管理</CardTitle>
          </div>
          <Button onClick={() => setUploadOpen(true)}>
            <Upload className="h-4 w-4" />
            上传固件
          </Button>
        </CardHeader>
        <CardContent>
          <div className="flex items-center mb-4">
            <span className="text-sm text-muted-foreground ml-auto">共 {total} 个固件</span>
          </div>

          {loading ? (
            <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
          ) : data.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <p>暂无固件</p>
              <Button variant="outline" className="mt-4" onClick={() => setUploadOpen(true)}>上传第一个固件</Button>
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>名称</TableHead>
                    <TableHead>版本</TableHead>
                    <TableHead>大小</TableHead>
                    <TableHead>SHA256</TableHead>
                    <TableHead>说明</TableHead>
                    <TableHead>上传时间</TableHead>
                    <TableHead className="text-right">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data.map((fw) => (
                    <TableRow key={fw.id}>
                      <TableCell className="font-medium">{fw.name}</TableCell>
                      <TableCell><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{fw.version}</code></TableCell>
                      <TableCell>{formatSize(fw.file_size)}</TableCell>
                      <TableCell><code className="text-xs text-muted-foreground">{fw.checksum.slice(0, 16)}...</code></TableCell>
                      <TableCell className="max-w-[200px] truncate">{fw.description || '-'}</TableCell>
                      <TableCell>{new Date(fw.created_at).toLocaleString()}</TableCell>
                      <TableCell className="text-right">
                        <Button variant="ghost" size="sm" className="text-destructive hover:text-destructive" onClick={() => setDeleteTarget(fw)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {totalPages > 1 && (
                <div className="flex items-center justify-between mt-4 pt-4 border-t border-border">
                  <span className="text-sm text-muted-foreground">共 {total} 条</span>
                  <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</Button>
                    <span className="text-sm text-muted-foreground px-2">{page} / {totalPages}</span>
                    <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>下一页</Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* 上传固件 Dialog */}
      <Dialog open={uploadOpen} onOpenChange={(open) => {
        setUploadOpen(open);
        if (!open) { setName(''); setVersion(''); setDescription(''); }
      }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>上传固件</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 pt-2">
            <div className="space-y-2">
              <Label>固件名称 *</Label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="如: ESP32-温控器" />
            </div>
            <div className="space-y-2">
              <Label>版本号 *</Label>
              <Input value={version} onChange={(e) => setVersion(e.target.value)} placeholder="如: 1.0.1" />
            </div>
            <div className="space-y-2">
              <Label>版本说明</Label>
              <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="可选" />
            </div>
            <div className="space-y-2">
              <Label>固件文件 *</Label>
              <Input type="file" ref={fileRef} accept=".bin,.hex,.elf" />
            </div>
            <div className="flex gap-3 pt-2">
              <Button onClick={handleUpload} disabled={uploading} className="flex-1">
                {uploading ? '上传中...' : '确认上传'}
              </Button>
              <Button variant="outline" onClick={() => setUploadOpen(false)} className="flex-1">取消</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* 删除确认 */}
      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除固件「{deleteTarget?.name} v{deleteTarget?.version}」吗？关联的 OTA 任务也将无法使用此固件。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild><Button variant="outline">取消</Button></DialogClose>
            <Button variant="destructive" onClick={handleDelete}>删除</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
