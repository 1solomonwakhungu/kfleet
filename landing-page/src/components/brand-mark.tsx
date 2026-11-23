import { cn } from '@/lib/utils'

/** The kfleet mark. Decorative: it always sits next to the written name. */
export function BrandMark({ className }: { readonly className?: string }) {
  return (
    <img
      src="/brand/kfleet-mark.svg"
      alt=""
      width={28}
      height={28}
      className={cn('size-7 shrink-0', className)}
    />
  )
}
