import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface NamespaceSelectorProps {
  namespaces: string[];
  value?: string;
  onChange: (namespace: string | undefined) => void;
}

const ALL = '__all__';

export function NamespaceSelector({ namespaces, value, onChange }: NamespaceSelectorProps) {
  return (
    <Select value={value ?? ALL} onValueChange={(v) => onChange(v === ALL ? undefined : v)}>
      <SelectTrigger className="w-48">
        <SelectValue placeholder="All namespaces" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={ALL}>All namespaces</SelectItem>
        {namespaces.map((ns) => (
          <SelectItem key={ns} value={ns}>
            {ns}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
