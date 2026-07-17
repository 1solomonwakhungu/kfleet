import { forwardRef, type InputHTMLAttributes } from 'react'

import { cn } from '../../lib/utils'

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(function Input(
  { className, type = 'text', ...props },
  ref,
) {
  return (
    <input
      ref={ref}
      type={type}
      className={cn(
        'h-11 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground outline outline-2 outline-transparent outline-offset-1 placeholder:text-muted hover:bg-surface focus-visible:border-border focus-visible:outline-focus disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    />
  )
})
