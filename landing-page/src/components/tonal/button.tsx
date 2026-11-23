import * as React from 'react'
import { capture } from '@/lib/posthog'
import { cn } from '@/lib/utils'

export type TonalButtonVariant = 'primary' | 'dark' | 'outline' | 'tonal' | 'ghost'
export type TonalButtonSize = 'sm' | 'md' | 'lg' | 'icon'

const SIZE_STYLES: Record<TonalButtonSize, string> = {
  sm: 'px-4 py-2 text-[13px]',
  md: 'px-6 py-3 text-[14px]',
  lg: 'px-7 py-3.5 text-[15px]',
  icon: 'h-8 w-8 p-0 text-[13px]',
}

const VARIANT_STYLES: Record<TonalButtonVariant, string> = {
  primary: 'bg-[var(--t-cta-bg)] text-[var(--t-cta-fg,#ffffff)] hover:bg-[var(--t-cta-hover)]',
  dark: 'bg-[var(--t-dark-btn-bg,#202124)] text-[var(--t-dark-btn-fg,#ffffff)] hover:bg-[var(--t-dark-btn-hover,#3c4043)]',
  outline:
    'bg-transparent text-[var(--t-ink)] border border-[var(--t-border)] hover:bg-[var(--t-hover-overlay,rgba(32,33,36,0.04))]',
  tonal:
    'bg-[var(--t-blue-tint)] text-[var(--t-blue-hover)] hover:bg-[var(--t-blue-tint-hover,#d2e3fc)]',
  ghost:
    'bg-transparent text-[var(--t-ink-secondary)] hover:bg-[var(--t-surface-muted)] hover:text-[var(--t-ink)]',
}

const BASE_CLASSES =
  'inline-flex items-center justify-center gap-2 font-medium leading-none tracking-[-0.01em] cursor-pointer select-none rounded-[var(--t-radius-btn)] transition-all duration-200 [transition-timing-function:var(--t-ease)] active:scale-[0.98] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--t-blue)] disabled:opacity-50 disabled:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0'

export interface TonalButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  readonly variant?: TonalButtonVariant
  readonly size?: TonalButtonSize
}

export const TonalButton = React.forwardRef<HTMLButtonElement, TonalButtonProps>(
  ({ className, variant = 'primary', size = 'md', ...props }, ref) => (
    <button
      ref={ref}
      className={cn(BASE_CLASSES, SIZE_STYLES[size], VARIANT_STYLES[variant], className)}
      {...props}
    />
  ),
)

TonalButton.displayName = 'TonalButton'

interface LinkTarget {
  readonly href: string
  readonly external: boolean
}

export interface TonalLinkButtonProps
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'href' | 'target' | 'rel'> {
  readonly link: LinkTarget
  readonly event?: string
  readonly variant?: TonalButtonVariant
  readonly size?: TonalButtonSize
}

export function TonalLinkButton({
  link,
  event,
  variant = 'primary',
  size = 'md',
  className,
  onClick,
  children,
  ...props
}: TonalLinkButtonProps) {
  return (
    <a
      href={link.href}
      {...(link.external ? { target: '_blank', rel: 'noreferrer' } : {})}
      className={cn(BASE_CLASSES, SIZE_STYLES[size], VARIANT_STYLES[variant], className)}
      onClick={(clickEvent) => {
        if (event) {
          capture(event)
        }
        onClick?.(clickEvent)
      }}
      {...props}
    >
      {children}
    </a>
  )
}
