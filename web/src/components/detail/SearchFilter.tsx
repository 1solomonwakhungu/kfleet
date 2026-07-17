import { useEffect, useState } from 'react';
import { Input } from '@/components/ui/input';

interface SearchFilterProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

export function SearchFilter({ value, onChange, placeholder = 'Filter by name…', debounceMs = 250 }: SearchFilterProps) {
  const [draft, setDraft] = useState(value);

  useEffect(() => setDraft(value), [value]);

  useEffect(() => {
    const timer = setTimeout(() => onChange(draft), debounceMs);
    return () => clearTimeout(timer);
  }, [draft, debounceMs, onChange]);

  return <Input value={draft} onChange={(e) => setDraft(e.target.value)} placeholder={placeholder} className="w-64" />;
}
