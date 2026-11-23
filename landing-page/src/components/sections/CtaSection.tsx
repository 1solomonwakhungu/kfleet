import { ArrowUpRight, Star } from 'lucide-react'
import { LinkButton } from '@/components/link-button'
import { Section } from '@/components/section'
import { Card } from '@/components/ui/card'
import { siteCopy } from '@/content/siteCopy'

export function CtaSection() {
  const { cta } = siteCopy

  return (
    <Section aria-labelledby="cta-heading">
      <Card className="relative isolate items-center overflow-hidden border-primary/40 px-6 py-14 text-center sm:px-12 sm:py-20">
        <div aria-hidden="true" className="hero-glow pointer-events-none absolute inset-0 -z-10 opacity-80" />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent via-brand/80 to-transparent"
        />
        <h2
          id="cta-heading"
          className="max-w-3xl text-3xl leading-[1.1] font-semibold tracking-tight text-balance sm:text-4xl"
        >
          {cta.headline}
        </h2>
        <p className="mt-4 max-w-xl text-base leading-relaxed text-pretty text-muted-foreground sm:text-lg">
          {cta.description}
        </p>
        <div className="mt-8 flex w-full flex-col items-stretch justify-center gap-3 sm:w-auto sm:flex-row">
          <LinkButton link={cta.primaryCta} event="cta_live_demo_clicked" size="lg">
            {cta.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </LinkButton>
          <LinkButton link={cta.secondaryCta} event="cta_github_clicked" variant="outline" size="lg">
            <Star aria-hidden="true" />
            {cta.secondaryCta.label}
          </LinkButton>
        </div>
      </Card>
    </Section>
  )
}
