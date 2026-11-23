import * as React from 'react'
import { ChevronRight } from 'lucide-react'
import { capture } from '@/lib/posthog'
import { cn } from '@/lib/utils'

export interface ChevronLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  readonly tone?: 'blue' | 'ink'
  readonly external?: boolean
  readonly event?: string
}

export function ChevronLink({
  tone = 'blue',
  external = false,
  event,
  className,
  onClick,
  children,
  ...props
}: ChevronLinkProps) {
  return (
    <a
      {...(external ? { target: '_blank', rel: 'noreferrer' } : {})}
      className={cn(
        'group inline-flex items-center gap-1 text-[14px] font-medium tracking-[-0.01em] cursor-pointer',
        tone === 'blue'
          ? 'text-[var(--t-blue)] hover:text-[var(--t-blue-hover)]'
          : 'text-[var(--t-ink)] hover:text-[var(--t-ink-secondary)]',
        className,
      )}
      onClick={(clickEvent) => {
        if (event) {
          capture(event)
        }
        onClick?.(clickEvent)
      }}
      {...props}
    >
      <span className="group-hover:underline underline-offset-4">{children}</span>
      <ChevronRight
        aria-hidden="true"
        className="h-4 w-4 transition-transform duration-200 [transition-timing-function:var(--t-ease)] group-hover:translate-x-0.5"
      />
    </a>
  )
}
