import { Accordion as AccordionPrimitive } from "@base-ui/react/accordion"
import { ChevronDown } from "lucide-react"

import { cn } from "@/lib/utils"

type WithClassName<Props> = Omit<Props, "className"> & { className?: string }

function Accordion({
  className,
  ...props
}: WithClassName<AccordionPrimitive.Root.Props>) {
  return (
    <AccordionPrimitive.Root
      data-slot="accordion"
      className={cn("flex flex-col", className)}
      {...props}
    />
  )
}

function AccordionItem({
  className,
  ...props
}: WithClassName<AccordionPrimitive.Item.Props>) {
  return (
    <AccordionPrimitive.Item
      data-slot="accordion-item"
      className={cn("border-b border-border/60 last:border-b-0", className)}
      {...props}
    />
  )
}

function AccordionTrigger({
  className,
  children,
  ...props
}: WithClassName<AccordionPrimitive.Trigger.Props>) {
  return (
    <AccordionPrimitive.Header className="flex">
      <AccordionPrimitive.Trigger
        data-slot="accordion-trigger"
        className={cn(
          "group/accordion-trigger flex flex-1 items-start justify-between gap-4 rounded-sm py-5 text-left text-base font-medium text-foreground",
          "transition-colors duration-150 hover:text-brand",
          className
        )}
        {...props}
      >
        {children}
        <ChevronDown
          aria-hidden="true"
          className="mt-1 size-4 shrink-0 text-muted-foreground transition-transform duration-200 ease-out group-data-panel-open/accordion-trigger:rotate-180"
        />
      </AccordionPrimitive.Trigger>
    </AccordionPrimitive.Header>
  )
}

function AccordionContent({
  className,
  children,
  ...props
}: WithClassName<AccordionPrimitive.Panel.Props>) {
  return (
    <AccordionPrimitive.Panel
      data-slot="accordion-content"
      className="h-(--accordion-panel-height) overflow-hidden text-sm transition-[height] duration-200 ease-out data-ending-style:h-0 data-starting-style:h-0 motion-reduce:transition-none"
      {...props}
    >
      <div className={cn("pb-5 leading-relaxed text-muted-foreground", className)}>
        {children}
      </div>
    </AccordionPrimitive.Panel>
  )
}

export { Accordion, AccordionItem, AccordionTrigger, AccordionContent }
