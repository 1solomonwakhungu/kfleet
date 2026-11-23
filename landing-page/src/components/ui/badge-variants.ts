import { cva } from "class-variance-authority"

// Tinted fills rather than solid ones: on the dark canvas a solid status
// colour shouts, and these badges sit inside dense lists and tables.
export const badgeVariants = cva(
  "inline-flex w-fit shrink-0 items-center justify-center gap-1 rounded-full border font-medium whitespace-nowrap [&_svg]:pointer-events-none [&_svg]:size-3 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "border-border/60 bg-secondary text-secondary-foreground",
        primary: "border-primary/40 bg-primary/20 text-brand",
        outline: "border-border bg-transparent text-muted-foreground",
        success: "border-success/25 bg-success/10 text-success",
        warning: "border-warning/25 bg-warning/10 text-warning",
        destructive: "border-destructive/30 bg-destructive/10 text-destructive",
      },
      size: {
        sm: "h-5 px-1.5 text-[11px]",
        default: "h-6 px-2 text-xs",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)
