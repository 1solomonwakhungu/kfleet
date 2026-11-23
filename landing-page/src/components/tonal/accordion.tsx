import * as React from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface AccordionItemData {
  readonly id: string
  readonly question: string
  readonly answer: string
}

export function TonalAccordion({
  items,
  onItemOpen,
  className,
}: {
  readonly items: readonly AccordionItemData[]
  readonly onItemOpen?: (id: string) => void
  readonly className?: string
}) {
  const [openIndex, setOpenIndex] = React.useState<number | null>(0)

  return (
    <div
      className={cn(
        'divide-y divide-[var(--t-border-subtle)] border-y border-[var(--t-border-subtle)]',
        className,
      )}
    >
      {items.map((item, idx) => {
        const isOpen = openIndex === idx
        const panelId = `faq-panel-${item.id}`
        return (
          <div key={item.id}>
            <button
              type="button"
              onClick={() => {
                const next = isOpen ? null : idx
                setOpenIndex(next)
                if (next !== null) {
                  onItemOpen?.(item.id)
                }
              }}
              aria-expanded={isOpen}
              aria-controls={panelId}
              className="group flex w-full cursor-pointer items-center justify-between gap-4 py-5 text-left"
            >
              <span className="text-[16px] font-medium tracking-[-0.015em] text-[var(--t-ink)] transition-colors group-hover:text-[var(--t-blue)]">
                {item.question}
              </span>
              <ChevronDown
                aria-hidden="true"
                className={cn(
                  'h-5 w-5 shrink-0 text-[var(--t-ink-muted)] transition-transform duration-250 [transition-timing-function:var(--t-ease)]',
                  isOpen && 'rotate-180 text-[var(--t-ink)]',
                )}
              />
            </button>
            <div
              id={panelId}
              className={cn(
                'grid transition-all duration-250 [transition-timing-function:var(--t-ease)]',
                isOpen ? 'grid-rows-[1fr] pb-5 opacity-100' : 'grid-rows-[0fr] opacity-0',
              )}
            >
              <div className="overflow-hidden">
                <p className="max-w-2xl text-[14.5px] leading-[1.65] text-pretty text-[var(--t-ink-secondary)]">
                  {item.answer}
                </p>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
