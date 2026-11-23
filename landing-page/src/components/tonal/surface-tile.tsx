import * as React from 'react'
import { cn } from '@/lib/utils'

export type SurfaceTone = 'white' | 'neutral' | 'warm' | 'cool' | 'sky' | 'mint' | 'sand'

export interface SurfaceTileProps extends React.HTMLAttributes<HTMLDivElement> {
  readonly tone?: SurfaceTone
  readonly bordered?: boolean
  readonly padding?: 'none' | 'sm' | 'md' | 'lg'
}

const TONE_MAP: Record<SurfaceTone, string> = {
  white: 'bg-[var(--t-elevated)]',
  neutral: 'bg-[var(--t-surface)]',
  warm: 'bg-[var(--t-surface-warm)]',
  cool: 'bg-[var(--t-surface-cool)]',
  sky: 'bg-[var(--t-surface-sky)]',
  mint: 'bg-[var(--t-surface-mint)]',
  sand: 'bg-[var(--t-surface-sand)]',
}

const PAD_MAP = {
  none: 'p-0',
  sm: 'p-5',
  md: 'p-6 sm:p-8',
  lg: 'p-7 sm:p-10',
}

export function SurfaceTile({
  tone = 'neutral',
  bordered = false,
  padding = 'md',
  className,
  ...props
}: SurfaceTileProps) {
  return (
    <div
      className={cn(
        'rounded-[var(--t-radius-card)] text-[var(--t-ink)] transition-colors duration-300 [transition-timing-function:var(--t-ease)]',
        TONE_MAP[tone],
        PAD_MAP[padding],
        (bordered || tone === 'white') && 'border border-[var(--t-border-subtle)]',
        className,
      )}
      {...props}
    />
  )
}
