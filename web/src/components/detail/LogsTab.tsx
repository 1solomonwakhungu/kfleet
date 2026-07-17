import { useEffect, useMemo, useRef, useState } from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { usePodLogs } from '@/hooks/usePodLogs';
import type { PodInfo } from '@/types/resources';

interface LogsTabProps {
  clusterId: string;
  pods: PodInfo[];
  selectedPod?: PodInfo;
  onSelectPod: (pod: PodInfo | undefined) => void;
}

export function LogsTab({ clusterId, pods, selectedPod, onSelectPod }: LogsTabProps) {
  const [container, setContainer] = useState<string | undefined>(undefined);
  const [autoScroll, setAutoScroll] = useState(true);
  const bottomRef = useRef<HTMLDivElement>(null);

  const { lines, connected, error, clear } = usePodLogs({
    clusterId,
    namespace: selectedPod?.namespace ?? '',
    pod: selectedPod?.name ?? '',
    container,
  });

  useEffect(() => {
    if (autoScroll) bottomRef.current?.scrollIntoView({ block: 'end' });
  }, [lines, autoScroll]);

  const podKey = useMemo(() => (selectedPod ? `${selectedPod.namespace}/${selectedPod.name}` : undefined), [selectedPod]);

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-2">
        <Select
          value={podKey}
          onValueChange={(key) => {
            const pod = pods.find((p) => `${p.namespace}/${p.name}` === key);
            onSelectPod(pod);
            setContainer(undefined);
          }}
        >
          <SelectTrigger className="w-64">
            <SelectValue placeholder="Select a pod" />
          </SelectTrigger>
          <SelectContent>
            {pods.map((pod) => (
              <SelectItem key={`${pod.namespace}/${pod.name}`} value={`${pod.namespace}/${pod.name}`}>
                {pod.namespace}/{pod.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <span className="text-xs text-muted-foreground">{connected ? 'streaming' : error ?? 'idle'}</span>

        <div className="ml-auto flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setAutoScroll((v) => !v)}>
            {autoScroll ? 'Auto-scroll: on' : 'Auto-scroll: off'}
          </Button>
          <Button variant="outline" size="sm" onClick={clear}>
            Clear
          </Button>
        </div>
      </div>

      <ScrollArea className="h-[480px] rounded-md border border-border bg-black/40">
        <pre className="whitespace-pre-wrap break-all p-3 font-mono text-xs text-zinc-200">
          {selectedPod ? lines.join('\n') || 'No log output yet.' : 'Select a pod to view logs.'}
        </pre>
        <div ref={bottomRef} />
      </ScrollArea>
    </div>
  );
}
