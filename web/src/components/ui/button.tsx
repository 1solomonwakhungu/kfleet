import { forwardRef, type ButtonHTMLAttributes } from 'react'

import { cn } from '../../lib/utils'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
}

const variants = {
  default: 'bg-accent text-accent-foreground hover:brightness-105',
  outline: 'border border-border bg-transparent text-foreground hover:bg-elevated',
  ghost: 'bg-transparent text-muted hover:bg-elevated hover:text-foreground',
  danger: 'bg-danger text-danger-foreground hover:brightness-110',
} as const

const sizes = {
  sm: 'min-h-9 px-3 text-sm',
  md: 'min-h-11 px-4 text-sm',
} as const

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = 'default', size = 'md', type = 'button', ...props },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type}
      className={cn(
        'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md font-semibold transition-[background-color,color,transform,filter] duration-150 ease-out active:translate-y-px disabled:cursor-not-allowed disabled:opacity-50',
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  )
})
