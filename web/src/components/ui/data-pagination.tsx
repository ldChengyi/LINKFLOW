import { useState } from 'react';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';

interface DataPaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  className?: string;
}

function getPageNumbers(current: number, total: number): (number | 'ellipsis')[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }
  const pages: (number | 'ellipsis')[] = [1];
  if (current > 3) pages.push('ellipsis');
  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);
  for (let i = start; i <= end; i++) pages.push(i);
  if (current < total - 2) pages.push('ellipsis');
  pages.push(total);
  return pages;
}

export function DataPagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50, 100],
  className,
}: DataPaginationProps) {
  const [jumpValue, setJumpValue] = useState('');
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const pageNumbers = getPageNumbers(page, totalPages);

  const handleJump = () => {
    const target = parseInt(jumpValue, 10);
    if (!isNaN(target) && target >= 1 && target <= totalPages) {
      onPageChange(target);
    }
    setJumpValue('');
  };

  return (
    <div className={cn('flex items-center justify-between gap-4 px-1 pt-4 mt-4 border-t border-border flex-wrap', className)}>
      {/* 左：总数 */}
      <span className="text-sm text-muted-foreground shrink-0">
        共 <span className="font-medium text-foreground">{total}</span> 条记录
      </span>

      {/* 中：页码导航 */}
      <div className="flex items-center gap-1">
        <Button
          variant="outline" size="icon" className="h-8 w-8"
          onClick={() => onPageChange(1)} disabled={page <= 1}
          title="首页"
        >
          <ChevronsLeft className="h-4 w-4" />
        </Button>
        <Button
          variant="outline" size="icon" className="h-8 w-8"
          onClick={() => onPageChange(page - 1)} disabled={page <= 1}
          title="上一页"
        >
          <ChevronLeft className="h-4 w-4" />
        </Button>

        {pageNumbers.map((p, i) =>
          p === 'ellipsis' ? (
            <span key={`e-${i}`} className="w-8 text-center text-sm text-muted-foreground select-none">…</span>
          ) : (
            <Button
              key={p}
              variant={p === page ? 'default' : 'outline'}
              size="icon"
              className="h-8 w-8 text-sm"
              onClick={() => onPageChange(p)}
            >
              {p}
            </Button>
          )
        )}

        <Button
          variant="outline" size="icon" className="h-8 w-8"
          onClick={() => onPageChange(page + 1)} disabled={page >= totalPages}
          title="下一页"
        >
          <ChevronRight className="h-4 w-4" />
        </Button>
        <Button
          variant="outline" size="icon" className="h-8 w-8"
          onClick={() => onPageChange(totalPages)} disabled={page >= totalPages}
          title="末页"
        >
          <ChevronsRight className="h-4 w-4" />
        </Button>
      </div>

      {/* 右：每页条数 + 跳转 */}
      <div className="flex items-center gap-3 shrink-0">
        {onPageSizeChange && (
          <div className="flex items-center gap-1.5">
            <span className="text-sm text-muted-foreground">每页</span>
            <Select
              value={String(pageSize)}
              onValueChange={(v) => { onPageSizeChange(Number(v)); onPageChange(1); }}
            >
              <SelectTrigger className="h-8 w-[72px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {pageSizeOptions.map((s) => (
                  <SelectItem key={s} value={String(s)}>{s} 条</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        <div className="flex items-center gap-1.5">
          <span className="text-sm text-muted-foreground">跳转</span>
          <Input
            className="h-8 w-14 text-center px-1"
            value={jumpValue}
            onChange={(e) => setJumpValue(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleJump()}
            placeholder={String(page)}
          />
          <Button variant="outline" size="sm" className="h-8 px-3" onClick={handleJump}>
            GO
          </Button>
        </div>
      </div>
    </div>
  );
}
