import { useEffect, useMemo, useRef, useState } from 'react'
import { TriangleDownIcon } from '@primer/octicons-react'
import styles from './SearchableSelect.module.css'

export interface SearchableSelectOption {
  value: string
  label: string
}

interface SearchableSelectProps {
  value: string
  options: SearchableSelectOption[]
  onChange: (value: string) => void
  ariaLabel: string
  id?: string
  placeholder?: string
  emptyMessage?: string
  disabled?: boolean
}

function normalize(text: string) {
  return text.trim().toLowerCase()
}

/**
 * Single-select combobox with type-to-search filtering. Renders a text input
 * that opens a filterable listbox; keyboard support: ArrowUp/ArrowDown move,
 * Enter selects, Escape closes.
 */
export function SearchableSelect({
  value,
  options,
  onChange,
  ariaLabel,
  id,
  placeholder = 'Search…',
  emptyMessage = 'No matches',
  disabled,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [activeIndex, setActiveIndex] = useState(0)
  const containerRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const listId = `${id ?? 'searchable-select'}-listbox`

  const selected = options.find((option) => option.value === value)

  const filtered = useMemo(() => {
    const needle = normalize(query)
    if (!needle) return options
    return options.filter((option) => normalize(option.label).includes(needle))
  }, [options, query])

  useEffect(() => {
    if (!open) return
    const onPointerDown = (event: PointerEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('pointerdown', onPointerDown)
    return () => document.removeEventListener('pointerdown', onPointerDown)
  }, [open])

  useEffect(() => {
    if (!open) setActiveIndex(0)
  }, [open, query])

  const openList = () => {
    if (disabled) return
    setQuery('')
    setActiveIndex(Math.max(0, options.findIndex((option) => option.value === value)))
    setOpen(true)
  }

  const closeList = () => {
    setOpen(false)
    setQuery('')
  }

  const commit = (option: SearchableSelectOption) => {
    setOpen(false)
    setQuery('')
    if (option.value !== value) onChange(option.value)
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (!open) {
      if (event.key === 'ArrowDown' || event.key === 'ArrowUp' || event.key === 'Enter') {
        event.preventDefault()
        openList()
      }
      return
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveIndex((index) => Math.min(index + 1, Math.max(filtered.length - 1, 0)))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveIndex((index) => Math.max(index - 1, 0))
    } else if (event.key === 'Enter') {
      event.preventDefault()
      const option = filtered[activeIndex]
      if (option) commit(option)
    } else if (event.key === 'Escape') {
      event.preventDefault()
      setOpen(false)
      setQuery('')
    }
  }

  return (
    <div ref={containerRef} className={styles.container}>
      <div className={styles.trigger}>
        <input
          id={id}
          type="text"
          role="combobox"
          aria-expanded={open}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-label={ariaLabel}
          disabled={disabled}
          value={open ? query : selected ? selected.label : ''}
          placeholder={placeholder}
          onChange={(event) => {
            setQuery(event.target.value)
            setActiveIndex(0)
            if (!open) setOpen(true)
          }}
          onFocus={openList}
          onClick={() => {
            // Only open on click when closed; clicking inside the input while
            // already open must not reset the typed query.
            if (!open) openList()
          }}
          onBlur={(event) => {
            if (!event.relatedTarget || !containerRef.current?.contains(event.relatedTarget as Node)) closeList()
          }}
          onKeyDown={handleKeyDown}
          className={styles.input}
          ref={inputRef}
          aria-activedescendant={open && filtered.length > 0 ? `${listId}-opt-${activeIndex}` : undefined}
        />
        <TriangleDownIcon className={styles.chevron} aria-hidden="true" />
      </div>
      {open && (
        <ul id={listId} role="listbox" aria-label={ariaLabel} className={styles.list}>
          {filtered.length === 0 ? (
            <li role="option" aria-selected="false" aria-disabled="true" className={styles.empty}>
              {emptyMessage}
            </li>
          ) : (
            filtered.map((option, index) => (
              <li key={option.value}>
                <button
                  type="button"
                  role="option"
                  id={`${listId}-opt-${index}`}
                  tabIndex={-1}
                  aria-selected={option.value === value}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => commit(option)}
                  onMouseEnter={() => setActiveIndex(index)}
                  className={`${styles.option} ${index === activeIndex ? styles.optionActive : ''}`}
                >
                  <span className={styles.optionLabel}>{option.label}</span>
                </button>
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  )
}
