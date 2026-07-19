import { useEffect, useId, useState } from 'react';
import { Search, X } from 'lucide-react';
import { Input } from '@/components/ui/input';

interface SearchFilterProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

export function SearchFilter({ value, onChange, placeholder = 'Filter resources…', debounceMs = 250 }: SearchFilterProps) {
  const [draft, setDraft] = useState(value);
  const inputId = useId();

  useEffect(() => setDraft(value), [value]);

  useEffect(() => {
    if (draft === value) return;
    const timer = setTimeout(() => onChange(draft), debounceMs);
    return () => clearTimeout(timer);
  }, [draft, value, debounceMs, onChange]);

  return (
    <div className="relative w-[9.5rem] shrink-0 sm:w-64">
      <label htmlFor={inputId} className="sr-only">Filter resources</label>
      <Search
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-blue-400"
        aria-hidden="true"
      />
      <Input
        id={inputId}
        type="search"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Escape' && draft) {
            event.preventDefault();
            setDraft('');
          }
        }}
        placeholder={placeholder}
        aria-keyshortcuts="Escape"
        className="w-full border-border bg-surface pl-9 pr-10 hover:border-blue-500/50 hover:bg-elevated focus-visible:border-blue-500/60 focus-visible:outline-blue-500 [&::-webkit-search-cancel-button]:hidden"
      />
      {draft && (
        <button
          type="button"
          onClick={() => setDraft('')}
          className="absolute right-0 top-0 grid h-11 w-11 place-items-center rounded-md text-muted outline-none hover:text-blue-300 focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:ring-offset-background active:text-blue-400"
          aria-label="Clear resource filter"
        >
          <X className="h-4 w-4" aria-hidden="true" />
        </button>
      )}
    </div>
  );
}
