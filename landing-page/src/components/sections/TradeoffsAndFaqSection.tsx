import { CornerDownRight } from 'lucide-react'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Card, CardContent, CardDescription, CardHeader } from '@/components/ui/card'
import { Section } from '@/components/section'
import { siteCopy } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'

export function TradeoffsAndFaqSection() {
  const { tradeoffs, faq } = siteCopy

  return (
    <Section id={faq.id} aria-labelledby="faq-heading">
      <div className="grid gap-12 lg:grid-cols-12 lg:gap-10">
        <Card className="self-start lg:col-span-5">
          <CardHeader>
            <p className="text-sm font-medium text-brand">{tradeoffs.eyebrow}</p>
            <h2 className="text-2xl leading-tight font-semibold tracking-tight text-balance">
              {tradeoffs.title}
            </h2>
            <CardDescription>{tradeoffs.subtitle}</CardDescription>
          </CardHeader>
          <CardContent className="px-0 pb-0">
            <ul className="divide-y divide-border/60 border-t border-border/60">
              {tradeoffs.items.map((item) => (
                <li key={item.scenario} className="flex flex-col gap-2 px-6 py-5">
                  <h3 className="text-sm font-medium text-foreground">{item.scenario}</h3>
                  <p className="flex items-start gap-2 text-sm text-brand">
                    <CornerDownRight aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
                    {item.recommendation}
                  </p>
                  <p className="text-[13px] leading-relaxed text-pretty text-muted-foreground">
                    {item.reason}
                  </p>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <div className="lg:col-span-7">
          <div className="flex flex-col gap-4">
            <p className="text-sm font-medium text-brand">{faq.eyebrow}</p>
            <h2
              id="faq-heading"
              className="text-3xl leading-[1.1] font-semibold tracking-tight text-balance sm:text-4xl"
            >
              {faq.title}
            </h2>
            <p className="text-base leading-relaxed text-pretty text-muted-foreground">{faq.subtitle}</p>
          </div>

          <Accordion
            defaultValue={[faq.items[0].id]}
            hiddenUntilFound
            className="mt-8 border-t border-border/60"
            onValueChange={(openItems) => {
              if (openItems.length > 0) {
                capture('faq_item_opened')
              }
            }}
          >
            {faq.items.map((item) => (
              <AccordionItem key={item.id} value={item.id}>
                <AccordionTrigger>{item.question}</AccordionTrigger>
                <AccordionContent>
                  <p className="max-w-[65ch] text-pretty">{item.answer}</p>
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
      </div>
    </Section>
  )
}
