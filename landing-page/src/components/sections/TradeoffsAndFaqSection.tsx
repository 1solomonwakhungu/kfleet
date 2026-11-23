import { CornerDownRight } from 'lucide-react'
import { Section } from '@/components/section'
import { TonalAccordion } from '@/components/tonal/accordion'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import { siteCopy } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'

export function TradeoffsAndFaqSection() {
  const { tradeoffs, faq } = siteCopy

  return (
    <Section id={faq.id} aria-labelledby="faq-heading">
      <div className="grid gap-10 lg:grid-cols-12 lg:gap-12">
        <SurfaceTile tone="warm" padding="lg" className="flex flex-col gap-6 self-start lg:col-span-5">
          <div className="flex flex-col gap-2.5">
            <p className="text-[13px] font-medium text-[var(--t-ink-muted)]">{tradeoffs.eyebrow}</p>
            <h2 className="t-headline text-[26px] text-balance sm:text-[28px]">{tradeoffs.title}</h2>
            <p className="t-body text-[14.5px] text-pretty">{tradeoffs.subtitle}</p>
          </div>

          <ul className="flex flex-col gap-3">
            {tradeoffs.items.map((item) => (
              <li
                key={item.scenario}
                className="flex flex-col gap-1.5 rounded-[18px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] p-4"
              >
                <h3 className="text-[14.5px] font-medium tracking-[-0.01em] text-[var(--t-ink)]">
                  {item.scenario}
                </h3>
                <p className="flex items-start gap-2 text-[13.5px] font-medium text-[var(--t-blue)]">
                  <CornerDownRight aria-hidden="true" className="mt-0.5 h-4 w-4 shrink-0" />
                  <span>{item.recommendation}</span>
                </p>
                <p className="text-[13px] leading-[1.55] text-pretty text-[var(--t-ink-secondary)]">
                  {item.reason}
                </p>
              </li>
            ))}
          </ul>
        </SurfaceTile>

        <div className="lg:col-span-7">
          <div className="flex flex-col gap-3">
            <p className="text-[13px] font-medium text-[var(--t-blue)]">{faq.eyebrow}</p>
            <h2 id="faq-heading" className="t-headline text-[30px] text-balance sm:text-[38px]">
              {faq.title}
            </h2>
            <p className="t-body text-[15.5px] text-pretty sm:text-[16.5px]">{faq.subtitle}</p>
          </div>

          <TonalAccordion
            items={faq.items}
            onItemOpen={() => capture('faq_item_opened')}
            className="mt-8"
          />
        </div>
      </div>
    </Section>
  )
}
