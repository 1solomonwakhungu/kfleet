import { useEffect, useId, useMemo, useRef, useState } from 'react';
import { AlertTriangle, ArrowDownToLine, Eraser, Loader2, Search, SquareTerminal, Wifi, WifiOff, X } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { usePodLogs } from '@/hooks/usePodLogs';
import { cn } from '@/lib/utils';
import type { PodInfo } from '@/types/resources';
import { ResourceState } from './ResourceTabState';

interface LogsTabProps {
  clusterId: string;
  pods: PodInfo[];
  selectedPod?: PodInfo;
  onSelectPod: (pod: PodInfo | undefined) => void;
}

export function LogsTab({ clusterId, pods, selectedPod, onSelectPod }: LogsTabProps) {
  const [autoScroll, setAutoScroll] = useState(true);
  const [wrapLines, setWrapLines] = useState(true);
  const [filter, setFilter] = useState('');
  const viewerRef = useRef<HTMLDivElement>(null);
  const podSelectId = useId();
  const logFilterId = useId();

  const { lines, connected, error, clear } = usePodLogs({
    clusterId,
    namespace: selectedPod?.namespace ?? '',
    pod: selectedPod?.name ?? '',
  });

  useEffect(() => {
    if (!autoScroll) return;
    const viewport = viewerRef.current?.querySelector<HTMLElement>('[data-radix-scroll-area-viewport]');
    if (viewport) viewport.scrollTop = viewport.scrollHeight;
  }, [lines, autoScroll]);

  const podKey = useMemo(
    () => (selectedPod ? `${selectedPod.namespace}/${selectedPod.name}` : undefined),
    [selectedPod],
  );
  const query = filter.trim().toLowerCase();
  const visibleLines = useMemo(
    () =>
      lines
        .map((line, index) => ({ line, number: index + 1 }))
        .filter(({ line }) => !query || line.toLowerCase().includes(query)),
    [lines, query],
  );

  if (pods.length === 0 && !selectedPod) {
    return (
      <ResourceState
        kind="empty"
        title="No pods available"
        description="Logs become available after the cluster reports at least one pod."
      />
    );
  }

  const status = connected ? 'Streaming' : error ? 'Reconnecting' : selectedPod ? 'Connecting' : 'Idle';
  const StatusIcon = connected ? Wifi : error ? WifiOff : selectedPod ? Loader2 : SquareTerminal;

  return (
    <section aria-label="Pod log viewer" className="overflow-hidden rounded-lg border border-border bg-surface">
      <div className="grid gap-3 border-b border-border p-3 lg:grid-cols-[minmax(14rem,1fr)_minmax(12rem,0.7fr)_auto] lg:items-end">
        <div className="min-w-0 space-y-1.5">
          <label className="text-xs font-medium text-muted" htmlFor={podSelectId}>Pod</label>
          <Select
            value={podKey ?? ''}
            onValueChange={(key) => {
              const pod = pods.find((candidate) => `${candidate.namespace}/${candidate.name}` === key);
              clear();
              setFilter('');
              setAutoScroll(true);
              onSelectPod(pod);
            }}
          >
            <SelectTrigger
              id={podSelectId}
              className="h-11 w-full border-border bg-background hover:border-blue-500/50 hover:bg-elevated focus:border-blue-500/60 focus:ring-2 focus:ring-blue-500"
              aria-label="Pod for log stream"
            >
              <SelectValue placeholder="Select a pod" />
            </SelectTrigger>
            <SelectContent className="border-border bg-elevated text-foreground">
              {pods.map((pod) => (
                <SelectItem
                  key={`${pod.namespace}/${pod.name}`}
                  value={`${pod.namespace}/${pod.name}`}
                  className="min-h-11 focus:bg-blue-500/15 focus:text-blue-200"
                >
                  {pod.namespace}/{pod.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="min-w-0 space-y-1.5">
          <label className="text-xs font-medium text-muted" htmlFor={logFilterId}>Find in logs</label>
          <div className="relative">
            <Search
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-blue-400"
              aria-hidden="true"
            />
            <Input
              id={logFilterId}
              type="search"
              value={filter}
              onChange={(event) => setFilter(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Escape' && filter) {
                  event.preventDefault();
                  setFilter('');
                }
              }}
              placeholder="Filter log output…"
              disabled={!selectedPod || lines.length === 0}
              aria-keyshortcuts="Escape"
              className="border-border bg-background pl-9 pr-10 hover:border-blue-500/50 hover:bg-elevated focus-visible:border-blue-500/60 focus-visible:outline-blue-500 [&::-webkit-search-cancel-button]:hidden"
            />
            {filter && (
              <button
                type="button"
                onClick={() => setFilter('')}
                className="absolute right-0 top-0 grid h-11 w-11 place-items-center rounded-md text-muted outline-none hover:text-blue-300 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:ring-offset-background active:text-blue-400"
                aria-label="Clear log filter"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            )}
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2 lg:justify-end" aria-label="Log viewer controls">
          <Button
            variant="outline"
            size="md"
            onClick={() => setWrapLines((current) => !current)}
            aria-pressed={wrapLines}
            className={cn(
              'border-border hover:border-blue-500/50 hover:text-blue-200',
              wrapLines && 'border-blue-500/50 bg-blue-500/10 text-blue-200',
            )}
          >
            <span className="font-mono text-xs" aria-hidden="true">↵</span>
            Wrap
          </Button>
          <Button
            variant="outline"
            size="md"
            onClick={() => setAutoScroll((current) => !current)}
            aria-pressed={autoScroll}
            className={cn(
              'border-border hover:border-blue-500/50 hover:text-blue-200',
              autoScroll && 'border-blue-500/50 bg-blue-500/10 text-blue-200',
            )}
          >
            <ArrowDownToLine className="h-4 w-4" aria-hidden="true" />
            Follow
          </Button>
          <Button
            variant="outline"
            size="md"
            onClick={clear}
            disabled={lines.length === 0}
            className="border-border hover:border-blue-500/50 hover:text-blue-200"
          >
            <Eraser className="h-4 w-4" aria-hidden="true" />
            Clear
          </Button>
        </div>
      </div>

      <div className="flex min-h-11 flex-wrap items-center justify-between gap-2 border-b border-border bg-background px-4 py-2 text-xs">
        <div
          className={cn(
            'flex items-center gap-2 font-medium',
            connected ? 'text-emerald-300' : error ? 'text-amber-300' : selectedPod ? 'text-blue-300' : 'text-muted',
          )}
          role="status"
          aria-live="polite"
        >
          <StatusIcon
            className={cn('h-3.5 w-3.5', selectedPod && !connected && !error && 'animate-spin')}
            aria-hidden="true"
          />
          <span>{status}</span>
          {error && <span className="font-normal text-muted">— {error}</span>}
        </div>
        <div className="flex items-center gap-3 font-mono tabular-nums text-muted">
          {query && <span>{visibleLines.length} matches</span>}
          <span>{lines.length} lines</span>
        </div>
      </div>

      <ScrollArea ref={viewerRef} className="h-[min(58dvh,32rem)] min-h-80 bg-background">
        <div
          role="log"
          aria-label={selectedPod ? `Logs for ${selectedPod.namespace}/${selectedPod.name}` : 'Pod logs'}
          aria-live="off"
          className={cn('min-h-80 p-3 font-mono text-xs leading-5 text-zinc-200', !wrapLines && 'w-max min-w-full')}
        >
          {!selectedPod ? (
            <div className="flex min-h-72 flex-col items-center justify-center gap-3 text-center text-muted">
              <SquareTerminal className="h-8 w-8 text-blue-400" aria-hidden="true" />
              <div>
                <p className="font-sans text-sm font-semibold text-foreground">Select a pod</p>
                <p className="mt-1 font-sans text-sm">Choose a pod above to open its live log stream.</p>
              </div>
            </div>
          ) : lines.length === 0 ? (
            <div className="flex min-h-72 flex-col items-center justify-center gap-3 text-center text-muted">
              {error ? (
                <AlertTriangle className="h-8 w-8 text-amber-300" aria-hidden="true" />
              ) : (
                <SquareTerminal className="h-8 w-8 text-blue-400" aria-hidden="true" />
              )}
              <div>
                <p className="font-sans text-sm font-semibold text-foreground">
                  {error ? 'Waiting for the log stream' : 'No log output yet'}
                </p>
                <p className="mt-1 max-w-md font-sans text-sm">
                  {error || 'The stream is connected; new output will appear here as the pod writes it.'}
                </p>
              </div>
            </div>
          ) : visibleLines.length === 0 ? (
            <div className="flex min-h-72 flex-col items-center justify-center gap-3 text-center text-muted">
              <Search className="h-8 w-8 text-blue-400" aria-hidden="true" />
              <div>
                <p className="font-sans text-sm font-semibold text-foreground">No matching log lines</p>
                <p className="mt-1 font-sans text-sm">No output contains “{filter.trim()}”.</p>
              </div>
            </div>
          ) : (
            visibleLines.map(({ line, number }) => (
              <div key={number} className="grid grid-cols-[3.25rem_minmax(0,1fr)] gap-3 hover:bg-blue-500/[0.06]">
                <span className="select-none text-right tabular-nums text-zinc-600" aria-hidden="true">{number}</span>
                <code className={cn('block text-zinc-200', wrapLines ? 'whitespace-pre-wrap break-all' : 'whitespace-pre')}>
                  {line || ' '}
                </code>
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </section>
  );
}
