/* Hallmark · pre-emit critique: P5 H4 E4 S5 R5 V4 */
/* Hallmark · component: resource-tabs · genre: modern-minimal · theme: existing Midnight
 * states: default · hover · focus · active · disabled · loading · error · success
 * contrast: inherited project tokens + semantic status tokens
 */
import type { ReactNode } from 'react';
import { AlertTriangle, Database, SearchX } from 'lucide-react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { cn } from '@/lib/utils';

interface ResourceStateProps {
  kind: 'empty' | 'error';
  title: string;
  description: string;
}

export function ResourceState({ kind, title, description }: ResourceStateProps) {
  const Icon = kind === 'error' ? AlertTriangle : SearchX;

  return (
    <div
      role={kind === 'error' ? 'alert' : 'status'}
      className={cn(
        'flex min-h-48 flex-col items-center justify-center gap-3 rounded-lg border border-dashed px-5 py-10 text-center',
        kind === 'error' ? 'border-red-500/40 bg-red-500/5' : 'border-border bg-surface',
      )}
    >
      <div
        className={cn(
          'grid h-10 w-10 place-items-center rounded-md border',
          kind === 'error'
            ? 'border-red-500/30 bg-red-500/10 text-red-300'
            : 'border-border bg-background text-muted',
        )}
      >
        <Icon className="h-5 w-5" aria-hidden="true" />
      </div>
      <div className="max-w-md space-y-1">
        <p className="font-semibold text-foreground">{title}</p>
        <p className="text-sm leading-6 text-muted">{description}</p>
      </div>
    </div>
  );
}

interface ResourceTableSkeletonProps {
  label: string;
  columns: number;
  rows?: number;
}

export function ResourceTableSkeleton({ label, columns, rows = 6 }: ResourceTableSkeletonProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-border bg-surface" aria-busy="true" aria-label={label}>
      <div className="flex h-12 items-center gap-3 border-b border-border px-4">
        <Database className="h-4 w-4 text-blue-400" aria-hidden="true" />
        <div className="h-3 w-24 animate-pulse rounded bg-elevated" />
      </div>
      <Table className="min-w-[720px]">
        <caption className="sr-only">{label}</caption>
        <TableHeader>
          <TableRow>
            {Array.from({ length: columns }, (_, index) => (
              <TableHead key={index} scope="col">
                <div className="h-2.5 w-16 animate-pulse rounded bg-elevated" />
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: rows }, (_, row) => (
            <TableRow key={row}>
              {Array.from({ length: columns }, (_, column) => (
                <TableCell key={column}>
                  <div
                    className={cn(
                      'h-3 animate-pulse rounded bg-elevated',
                      column === 0 ? 'w-36' : column % 2 === 0 ? 'w-14' : 'w-24',
                    )}
                  />
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

interface ResourceTablePanelProps {
  label: string;
  count: number;
  noun: string;
  children: ReactNode;
  className?: string;
}

export function ResourceTablePanel({ label, count, noun, children, className }: ResourceTablePanelProps) {
  return (
    <section aria-label={label} className={cn('overflow-hidden rounded-lg border border-border bg-surface', className)}>
      <div className="flex h-12 items-center justify-between gap-4 border-b border-border px-4">
        <div className="flex min-w-0 items-center gap-2 text-sm">
          <Database className="h-4 w-4 shrink-0 text-blue-400" aria-hidden="true" />
          <span className="font-semibold tabular-nums text-foreground">{count}</span>
          <span className="truncate text-muted">{count === 1 ? noun : `${noun}s`}</span>
        </div>
        <span className="shrink-0 text-xs text-muted sm:hidden">Scroll for details</span>
      </div>
      {children}
    </section>
  );
}
