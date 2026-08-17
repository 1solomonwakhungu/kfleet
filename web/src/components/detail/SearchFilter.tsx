import { useEffect, useState } from 'react'
import { TextInput } from '@primer/react'
import { SearchIcon, XIcon } from '@primer/octicons-react'

interface SearchFilterProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  debounceMs?: number
}

export function SearchFilter({
  value,
  onChange,
  placeholder = 'Filter resources…',
  debounceMs = 250,
}: SearchFilterProps) {
  const [draft, setDraft] = useState(value)

  useEffect(() => setDraft(value), [value])

  useEffect(() => {
    if (draft === value) return
    const timer = setTimeout(() => onChange(draft), debounceMs)
    return () => clearTimeout(timer)
  }, [draft, value, debounceMs, onChange])

  return (
    <TextInput
      type="search"
      aria-label="Filter resources"
      aria-keyshortcuts="Escape"
      value={draft}
      placeholder={placeholder}
      leadingVisual={SearchIcon}
      onChange={(event) => setDraft(event.target.value)}
      onKeyDown={(event) => {
        if (event.key === 'Escape' && draft) {
          event.preventDefault()
          setDraft('')
        }
      }}
      trailingAction={
        draft ? (
          <TextInput.Action icon={XIcon} aria-label="Clear resource filter" onClick={() => setDraft('')} />
        ) : undefined
      }
    />
  )
}
