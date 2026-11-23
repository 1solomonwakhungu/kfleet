import { ArrowUpRight, Star } from 'lucide-react'
import { Section } from '@/components/section'
import { TonalLinkButton } from '@/components/tonal/button'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import { siteCopy } from '@/content/siteCopy'

export function CtaSection() {
  const { cta } = siteCopy

  return (
    <Section aria-labelledby="cta-heading">
      <SurfaceTile
        tone="sky"
        padding="lg"
        className="flex flex-col items-center rounded-[var(--t-radius-hero)] px-6 py-14 text-center sm:px-12 sm:py-20"
      >
        <h2
          id="cta-heading"
          className="t-display max-w-3xl text-[32px] text-balance sm:text-[46px]"
        >
          {cta.headline}
        </h2>
        <p className="t-body mt-4 max-w-xl text-[16px] text-pretty sm:text-[17.5px]">
          {cta.description}
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3.5">
          <TonalLinkButton link={cta.primaryCta} event="cta_live_demo_clicked" size="lg">
            {cta.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </TonalLinkButton>
          <TonalLinkButton
            link={cta.secondaryCta}
            event="cta_github_clicked"
            variant="outline"
            size="lg"
            className="bg-[var(--t-elevated)]/70 hover:bg-[var(--t-elevated)]"
          >
            <Star aria-hidden="true" />
            {cta.secondaryCta.label}
          </TonalLinkButton>
        </div>
      </SurfaceTile>
    </Section>
  )
}
