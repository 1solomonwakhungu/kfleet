import type { ComponentProps } from 'react'
import { cn } from '@/lib/utils'

export function Section({ className, children, ...props }: ComponentProps<'section'>) {
  return (
    <section className={cn('page-container', className)} {...props}>
      {children}
    </section>
  )
}

interface SectionHeaderProps {
  readonly eyebrow: string
  readonly title: string
  readonly subtitle: string
  /** Id for the heading so the section can point at it with aria-labelledby. */
  readonly headingId: string
  readonly align?: 'left' | 'center'
  readonly className?: string
}

export function SectionHeader({
  eyebrow,
  title,
  subtitle,
  headingId,
  align = 'left',
  className,
}: SectionHeaderProps) {
  return (
    <div
      className={cn(
        'flex max-w-2xl flex-col gap-3',
        align === 'center' && 'mx-auto items-center text-center',
        className,
      )}
    >
      <p className="text-[13px] font-medium tracking-[-0.005em] text-[var(--t-blue)]">{eyebrow}</p>
      <h2 id={headingId} className="t-headline text-[30px] text-balance sm:text-[38px]">
        {title}
      </h2>
      <p className="t-body text-[15.5px] text-pretty sm:text-[16.5px]">{subtitle}</p>
    </div>
  )
}
