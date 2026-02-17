import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { AlertTriangle } from 'lucide-react';
import { alertLogApi, deviceApi } from '../api';
import type { AlertLog, Device } from '../api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { onWSMessage } from './Dashboard';

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

export default function AlertLogList() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<AlertLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [devices, setDevices] = useState<Device[]>([]);
  const [filterDeviceId, setFilterDeviceId] = useState('');

  useEffect(() => {
    deviceApi.list(1, 100).then((res) => setDevices(res.data.list || [])).catch(() => {});
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await alertLogApi.list(page, pageSize, filterDeviceId || undefined);
      setData(res.data.list || []);
      setTotal(res.data.total);
    } catch {
      toast.error('获取告警历史失败');
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

  // 实时告警推送：新告警插入列表顶部（仅当匹配当前筛选设备时）
  useEffect(() => {
    return onWSMessage((msg) => {
      if (msg.type === 'alert') {
        const log = msg.data as AlertLog;
        if (!filterDeviceId || log.device_id === filterDeviceId) {
          setData((prev) => [log, ...prev].slice(0, pageSize));
          setTotal((prev) => prev + 1);
        }
      }
    });
  }, [pageSize, filterDeviceId]);

  const totalPages = Math.ceil(total / pageSize);

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <div className="flex items-center gap-3">
          <AlertTriangle className="h-5 w-5 text-primary" />
          <CardTitle>告警历史</CardTitle>
        </div>
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
            <p>暂无告警记录</p>
          </div>
        ) : (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>时间</TableHead>
                  <TableHead>规则</TableHead>
                  <TableHead>设备</TableHead>
                  <TableHead>属性</TableHead>
                  <TableHead>条件</TableHead>
                  <TableHead>实际值</TableHead>
                  <TableHead>级别</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="text-muted-foreground text-sm whitespace-nowrap">
                      {new Date(item.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell className="font-medium">{item.rule_name}</TableCell>
                    <TableCell>{item.device_name}</TableCell>
                    <TableCell>{item.property_name}</TableCell>
                    <TableCell className="font-mono text-sm">{item.operator} {item.threshold}</TableCell>
                    <TableCell className="font-mono text-sm font-bold">{item.actual_value}</TableCell>
                    <TableCell>
                      <Badge variant={severityColors[item.severity] as any || 'secondary'}>
                        {severityLabels[item.severity] || item.severity}
                      </Badge>
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
  );
}
