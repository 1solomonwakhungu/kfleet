import { useId, useMemo } from 'react';
import { Layers3 } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface NamespaceSelectorProps {
  namespaces: string[];
  value?: string;
  onChange: (namespace: string | undefined) => void;
}

const ALL = '__all__';

export function NamespaceSelector({ namespaces, value, onChange }: NamespaceSelectorProps) {
  const triggerId = useId();
  const options = useMemo(
    () => Array.from(new Set(namespaces.filter(Boolean))).sort((a, b) => a.localeCompare(b)),
    [namespaces],
  );

  return (
    <div className="relative w-32 shrink-0 sm:w-48">
      <label htmlFor={triggerId} className="sr-only">Namespace</label>
      <Layers3
        className="pointer-events-none absolute left-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-blue-400"
        aria-hidden="true"
      />
      <Select value={value ?? ALL} onValueChange={(selected) => onChange(selected === ALL ? undefined : selected)}>
        <SelectTrigger
          id={triggerId}
          aria-label="Namespace"
          className="h-11 w-full border-border bg-surface pl-9 hover:border-blue-500/50 hover:bg-elevated focus:border-blue-500/60 focus:ring-2 focus:ring-blue-500"
        >
          <SelectValue placeholder="All namespaces" />
        </SelectTrigger>
        <SelectContent className="border-border bg-elevated text-foreground">
          <SelectItem value={ALL} className="min-h-11 focus:bg-blue-500/15 focus:text-blue-200">All namespaces</SelectItem>
          {options.map((namespace) => (
            <SelectItem
              key={namespace}
              value={namespace}
              className="min-h-11 focus:bg-blue-500/15 focus:text-blue-200"
            >
              {namespace}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
