import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { Search, RotateCw, Eye } from 'lucide-react';
import { auditLogApi } from '../api';
import type { AuditLog } from '../api';
import { Button } from '@/components/ui/button';
import { DataPagination } from '@/components/ui/data-pagination';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose,
} from '@/components/ui/dialog';

const resourceTypeMap: Record<string, { label: string; variant: string }> = {
  auth: { label: '认证', variant: 'outline' },
  thing_model: { label: '物模型', variant: 'secondary' },
  device: { label: '设备', variant: 'success' },
  scheduled_task: { label: '定时任务', variant: 'outline' },
  alert_rule: { label: '告警规则', variant: 'secondary' },
  alert_log: { label: '告警日志', variant: 'outline' },
};

export default function AuditLogList() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(15);

  const [category, setCategory] = useState('');
  const [action, setAction] = useState('');
  const [detailItem, setDetailItem] = useState<AuditLog | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await auditLogApi.list({
        page,
        page_size: pageSize,
        ...(category && { category }),
        ...(action && { action }),
      });
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取审计日志失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const handleSearch = () => {
    setPage(1);
    fetchData();
  };

  const handleReset = () => {
    setCategory('');
    setAction('');
    setPage(1);
    setTimeout(fetchData, 0);
  };


  const categoryLabel = (cat: string) => {
    switch (cat) {
      case 'api': return 'API';
      case 'device': return '设备';
      default: return cat;
    }
  };

  const categoryVariant = (cat: string) => {
    return cat === 'device' ? 'success' : 'secondary';
  };

  const statusColor = (code: number) => {
    if (code === 0) return '';
    if (code < 300) return 'text-green-400';
    if (code < 400) return 'text-amber-400';
    return 'text-destructive';
  };

  // 从 detail 中提取资源类型信息
  const getResourceType = (item: AuditLog) => {
    const rt = item.detail?.resource_type as string | undefined;
    if (rt && resourceTypeMap[rt]) return resourceTypeMap[rt];
    // 兼容旧数据：从 resource 路径推断
    if (item.resource?.includes('/thing-models')) return resourceTypeMap.thing_model;
    if (item.resource?.includes('/devices')) return resourceTypeMap.device;
    if (item.resource?.includes('/auth')) return resourceTypeMap.auth;
    return null;
  };

  // 获取资源名称（优先 detail 中的 resource_name / device_name）
  const getResourceName = (item: AuditLog): string => {
    if (item.detail?.resource_name) return String(item.detail.resource_name);
    if (item.detail?.device_name) return String(item.detail.device_name);
    if (item.detail?.name) return String(item.detail.name);
    if (item.detail?.email) return String(item.detail.email);
    return '';
  };

  return (
    <>
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle>审计日志</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-3 mb-4">
          <Select value={category} onValueChange={setCategory}>
            <SelectTrigger className="w-[130px]">
              <SelectValue placeholder="全部分类" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部分类</SelectItem>
              <SelectItem value="api">API 操作</SelectItem>
              <SelectItem value="device">设备事件</SelectItem>
            </SelectContent>
          </Select>

          <Input
            placeholder="搜索操作 (如 创建设备, 用户登录)"
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="max-w-[280px]"
            onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          />

          <Button variant="outline" size="sm" onClick={handleSearch}>
            <Search className="h-4 w-4 mr-1" />
            搜索
          </Button>
          <Button variant="ghost" size="sm" onClick={handleReset}>
            <RotateCw className="h-4 w-4 mr-1" />
            重置
          </Button>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-20 text-muted-foreground">加载中...</div>
        ) : data.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
            <p>暂无审计日志</p>
          </div>
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="whitespace-nowrap">分类</TableHead>
                  <TableHead className="whitespace-nowrap">操作</TableHead>
                  <TableHead className="whitespace-nowrap">资源类型</TableHead>
                  <TableHead className="whitespace-nowrap">资源名称</TableHead>
                  <TableHead className="whitespace-nowrap">状态码</TableHead>
                  <TableHead className="whitespace-nowrap">IP</TableHead>
                  <TableHead className="whitespace-nowrap">时间</TableHead>
                  <TableHead className="w-[50px]"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((item) => {
                  const rt = getResourceType(item);
                  const resourceName = getResourceName(item);
                  return (
                    <TableRow key={item.id}>
                      <TableCell className="whitespace-nowrap">
                        <Badge variant={categoryVariant(item.category)}>
                          {categoryLabel(item.category)}
                        </Badge>
                      </TableCell>
                      <TableCell className="font-medium">
                        {item.action}
                      </TableCell>
                      <TableCell className="whitespace-nowrap">
                        {rt ? (
                          <Badge variant={rt.variant as 'outline' | 'secondary' | 'success'}>
                            {rt.label}
                          </Badge>
                        ) : (
                          <span className="text-muted-foreground text-xs">-</span>
                        )}
                      </TableCell>
                      <TableCell>
                        {resourceName ? (
                          <span className="text-sm">{resourceName}</span>
                        ) : (
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded font-mono inline-block max-w-[180px] truncate align-middle" title={item.resource}>
                            {item.resource}
                          </code>
                        )}
                      </TableCell>
                      <TableCell>
                        {item.status_code > 0 && (
                          <span className={statusColor(item.status_code)}>{item.status_code}</span>
                        )}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm font-mono">
                        {item.ip || '-'}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-sm">
                        {new Date(item.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Button variant="ghost" size="sm" onClick={() => setDetailItem(item)}>
                          <Eye className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>

            <DataPagination
              page={page} pageSize={pageSize} total={total}
              onPageChange={setPage}
              onPageSizeChange={(s) => { setPageSize(s); setPage(1); }}
              pageSizeOptions={[15, 30, 50, 100]}
            />
          </>
        )}
      </CardContent>
    </Card>

    <Dialog open={!!detailItem} onOpenChange={(open) => !open && setDetailItem(null)}>
      <DialogContent className="sm:max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>审计日志详情</DialogTitle>
        </DialogHeader>
        {detailItem && (() => {
          const rt = getResourceType(detailItem);
          const resourceName = getResourceName(detailItem);
          return (
            <div className="space-y-4 text-sm">
              <div className="grid grid-cols-[90px_1fr] gap-x-4 gap-y-2.5">
                <span className="text-muted-foreground">分类</span>
                <span>
                  <Badge variant={categoryVariant(detailItem.category)}>
                    {categoryLabel(detailItem.category)}
                  </Badge>
                </span>
                <span className="text-muted-foreground">操作</span>
                <span className="font-medium">{detailItem.action}</span>
                {rt && (
                  <>
                    <span className="text-muted-foreground">资源类型</span>
                    <span>
                      <Badge variant={rt.variant as 'outline' | 'secondary' | 'success'}>
                        {rt.label}
                      </Badge>
                    </span>
                  </>
                )}
                {resourceName && (
                  <>
                    <span className="text-muted-foreground">资源名称</span>
                    <span>{resourceName}</span>
                  </>
                )}
                <span className="text-muted-foreground">请求路径</span>
                <code className="text-xs bg-muted px-1.5 py-0.5 rounded font-mono break-all">{detailItem.resource}</code>
                {detailItem.status_code > 0 && (
                  <>
                    <span className="text-muted-foreground">状态码</span>
                    <span className={statusColor(detailItem.status_code)}>{detailItem.status_code}</span>
                  </>
                )}
                {detailItem.latency_ms > 0 && (
                  <>
                    <span className="text-muted-foreground">耗时</span>
                    <span>{detailItem.latency_ms}ms</span>
                  </>
                )}
                <span className="text-muted-foreground">IP</span>
                <span className="font-mono">{detailItem.ip || '-'}</span>
                <span className="text-muted-foreground">时间</span>
                <span>{new Date(detailItem.created_at).toLocaleString()}</span>
              </div>
              {detailItem.detail && Object.keys(detailItem.detail).length > 0 && (
                <div className="space-y-1.5">
                  <span className="text-muted-foreground">请求详情</span>
                  <pre className="text-xs bg-muted p-3 rounded-lg overflow-auto max-h-[300px] font-mono">
                    {JSON.stringify(detailItem.detail, null, 2)}
                  </pre>
                </div>
              )}
              {detailItem.user_agent && (
                <div className="space-y-1.5">
                  <span className="text-muted-foreground">User-Agent</span>
                  <p className="text-xs text-muted-foreground break-all">{detailItem.user_agent}</p>
                </div>
              )}
            </div>
          );
        })()}
        <DialogClose asChild>
          <Button variant="outline" className="w-full">关闭</Button>
        </DialogClose>
      </DialogContent>
    </Dialog>
    </>
  );
}
