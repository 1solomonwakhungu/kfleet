import type { ComponentProps } from 'react'
import type { VariantProps } from 'class-variance-authority'
import { buttonVariants } from '@/components/ui/button-variants'
import { capture } from '@/lib/posthog'
import { cn } from '@/lib/utils'

interface LinkTarget {
  readonly href: string
  readonly external: boolean
}

type LinkButtonProps = Omit<ComponentProps<'a'>, 'href' | 'target' | 'rel'> &
  VariantProps<typeof buttonVariants> & {
    readonly link: LinkTarget
    /** PostHog event captured on click. */
    readonly event?: string
  }

/** An anchor styled as a button. External links open in a new tab. */
export function LinkButton({
  link,
  event,
  variant,
  size,
  className,
  onClick,
  children,
  ...props
}: LinkButtonProps) {
  return (
    <a
      href={link.href}
      {...(link.external ? { target: '_blank', rel: 'noreferrer' } : {})}
      className={cn(buttonVariants({ variant, size }), className)}
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
