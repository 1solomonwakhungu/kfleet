import * as React from 'react'
import { cn } from '@/lib/utils'

export type TonalChipTone = 'blue' | 'neutral' | 'mint' | 'warm' | 'warning' | 'danger'

export interface TonalChipProps extends React.HTMLAttributes<HTMLSpanElement> {
  readonly tone?: TonalChipTone
}

const CHIP_TONES: Record<TonalChipTone, string> = {
  blue: 'bg-[var(--t-blue-tint)] text-[var(--t-blue-hover)]',
  neutral: 'bg-[var(--t-surface-muted)] text-[var(--t-ink-secondary)]',
  mint: 'bg-[var(--t-surface-mint)] text-[var(--t-chip-mint-fg)]',
  warm: 'bg-[var(--t-surface-warm)] text-[var(--t-ink-secondary)]',
  warning: 'bg-[var(--t-chip-warning-bg)] text-[var(--t-chip-warning-fg)]',
  danger: 'bg-[var(--t-chip-danger-bg)] text-[var(--t-chip-danger-fg)]',
}

export function TonalChip({ tone = 'blue', className, ...props }: TonalChipProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-[12px] font-medium tracking-[-0.005em] [&_svg]:size-3.5 [&_svg]:shrink-0',
        CHIP_TONES[tone],
        className,
      )}
      {...props}
    />
  )
}
