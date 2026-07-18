import { type HTMLAttributes } from 'react'

import { cn } from '../../lib/utils'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: string
}

export function Badge({ className, ...props }: BadgeProps) {
  return (
    <span
      className={cn('inline-flex items-center whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-semibold', className)}
      {...props}
    />
  )
}
