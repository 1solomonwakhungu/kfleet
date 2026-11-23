import type { ComponentProps } from 'react'
import { cn } from '@/lib/utils'

export function Section({ className, children, ...props }: ComponentProps<'section'>) {
  return (
    <section className={cn('border-t border-border/60 py-20 sm:py-28', className)} {...props}>
      <div className="page-container">{children}</div>
    </section>
  )
}

interface SectionHeaderProps {
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  /** Id for the heading so the section can point at it with aria-labelledby. */
  readonly headingId: string
  readonly className?: string
}

export function SectionHeader({ eyebrow, title, subtitle, headingId, className }: SectionHeaderProps) {
  return (
    <div className={cn('flex max-w-2xl flex-col gap-4', className)}>
      <p className="text-sm font-medium text-brand">{eyebrow}</p>
      <h2
        id={headingId}
        className="text-3xl leading-[1.1] font-semibold tracking-tight text-balance sm:text-4xl"
      >
        {title}
      </h2>
      <p className="text-base leading-relaxed text-pretty text-muted-foreground sm:text-lg">
        {subtitle}
      </p>
    </div>
  )
}
