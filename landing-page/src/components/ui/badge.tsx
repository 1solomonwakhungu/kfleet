import type { ComponentProps } from "react"
import type { VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"
import { badgeVariants } from "@/components/ui/badge-variants"

type BadgeProps = ComponentProps<"span"> & VariantProps<typeof badgeVariants>

function Badge({ className, variant, size, ...props }: BadgeProps) {
  return (
    <span
      data-slot="badge"
      className={cn(badgeVariants({ variant, size }), className)}
      {...props}
    />
  )
}

export { Badge }
export type { BadgeProps }
