import { ArrowUpRight, Check, ChevronRight, Copy, Star } from 'lucide-react'
import { TonalLinkButton } from '@/components/tonal/button'
import { ChevronLink } from '@/components/tonal/chevron-link'
import { TonalChip } from '@/components/tonal/chip'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import { siteCopy } from '@/content/siteCopy'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { capture } from '@/lib/posthog'

// Intrinsic size of public/demo.gif, so the frame reserves its space before the GIF loads.
const DEMO_GIF_WIDTH = 900
const DEMO_GIF_HEIGHT = 563

export function HeroSection() {
  const { hero, links } = siteCopy
  const { announcement, quickCommand, demoFrame } = hero
  const { copied, copy } = useCopyToClipboard('hero_quick_command_copied')

  return (
    <section aria-labelledby="hero-heading" className="page-container pt-10 sm:pt-16">
      <div className="mx-auto flex max-w-3xl flex-col items-center text-center">
        <a
          href={announcement.href}
          target="_blank"
          rel="noreferrer"
          onClick={() => capture('release_notes_opened')}
          className="group mb-6 inline-flex max-w-full items-center gap-2.5 rounded-full border border-[var(--t-border-subtle)] bg-[var(--t-surface)] py-1 pr-3.5 pl-1.5 text-left text-[13px] font-medium text-[var(--t-ink-secondary)] transition-colors duration-200 hover:bg-[var(--t-surface-muted)] hover:text-[var(--t-ink)]"
        >
          <TonalChip tone="blue" className="font-mono text-[11.5px]">
            {announcement.badge}
          </TonalChip>
          <span className="truncate">{announcement.text}</span>
          <ChevronRight
            aria-hidden="true"
            className="h-4 w-4 shrink-0 text-[var(--t-ink-muted)] transition-transform duration-200 [transition-timing-function:var(--t-ease)] group-hover:translate-x-0.5 group-hover:text-[var(--t-ink)]"
          />
        </a>

        <h1
          id="hero-heading"
          className="t-display text-[40px] text-balance sm:text-[58px] lg:text-[64px]"
        >
          {hero.headline}{' '}
          <span className="block text-[var(--t-blue)]">{hero.headlineAccent}</span>
        </h1>

        <p className="t-body mt-5 max-w-2xl text-[16.5px] text-pretty sm:text-[18.5px]">
          {hero.subheadline}
        </p>

        <div className="mt-8 flex flex-wrap items-center justify-center gap-3.5">
          <TonalLinkButton link={hero.primaryCta} event="live_demo_cta_clicked" size="lg">
            {hero.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </TonalLinkButton>

          <TonalLinkButton
            link={hero.secondaryCta}
            event="hero_github_cta_clicked"
            variant="outline"
            size="lg"
          >
            <Star aria-hidden="true" />
            {hero.secondaryCta.label}
          </TonalLinkButton>

          <ChevronLink
            href={hero.tertiaryCta.href}
            event="quickstart_cta_clicked"
            className="px-2 py-2"
          >
            {hero.tertiaryCta.label}
          </ChevronLink>
        </div>

        <div className="mt-7 flex w-full max-w-3xl flex-col items-center">
          <button
            type="button"
            onClick={() => void copy(quickCommand.command)}
            aria-label={`Copy ${quickCommand.label}`}
            className="group inline-flex max-w-full cursor-pointer items-center gap-3 rounded-full border border-[var(--t-border-subtle)] bg-[var(--t-surface)] px-5 py-2.5 text-left font-mono text-[12.5px] text-[var(--t-ink-secondary)] transition-colors duration-200 hover:bg-[var(--t-surface-muted)] sm:text-[13px]"
          >
            <span aria-hidden="true" className="text-[var(--t-ink-muted)] select-none">
              $
            </span>
            <code className="scrollbar-none min-w-0 flex-1 overflow-x-auto whitespace-nowrap">
              {quickCommand.command}
            </code>
            {copied ? (
              <Check aria-hidden="true" className="h-4 w-4 shrink-0 text-[var(--t-chip-mint-fg)]" />
            ) : (
              <Copy
                aria-hidden="true"
                className="h-4 w-4 shrink-0 text-[var(--t-ink-muted)] transition-colors group-hover:text-[var(--t-ink)]"
              />
            )}
          </button>
          <span role="status" className="sr-only">
            {copied ? `Copied ${quickCommand.label}` : ''}
          </span>
          <p className="mt-2 text-[12.5px] text-[var(--t-ink-muted)]">{quickCommand.label}</p>
        </div>
      </div>

      {/* Tonal Browser Stage */}
      <figure className="mt-12 sm:mt-16">
        <SurfaceTile
          tone="cool"
          padding="md"
          className="rounded-[var(--t-radius-hero)] p-4 sm:p-7 lg:p-9"
        >
          <div className="overflow-hidden rounded-[20px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)]">
            <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-3 border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)] px-4 py-3">
              <div aria-hidden="true" className="flex gap-1.5">
                <span className="h-2.5 w-2.5 rounded-full bg-[var(--t-border)]" />
                <span className="h-2.5 w-2.5 rounded-full bg-[var(--t-border)]" />
                <span className="h-2.5 w-2.5 rounded-full bg-[var(--t-border)]" />
              </div>

              <div className="flex min-w-0 items-center gap-2 rounded-full border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] px-3.5 py-1 font-mono text-[12px] text-[var(--t-ink-secondary)]">
                <span
                  aria-hidden="true"
                  className="h-2 w-2 shrink-0 rounded-full bg-[var(--t-chip-mint-fg)]"
                />
                <span className="truncate">{demoFrame.urlLabel}</span>
              </div>

              <p className="hidden justify-self-end text-[12px] text-[var(--t-ink-muted)] lg:block">
                {demoFrame.statusLabel}
              </p>
            </div>

            <a
              href={links.liveDemoUrl}
              target="_blank"
              rel="noreferrer"
              onClick={() => capture('demo_gif_clicked')}
              className="block bg-[var(--t-surface-muted)] focus-visible:-outline-offset-2"
            >
              <img
                src={demoFrame.gifSrc}
                alt={demoFrame.gifAlt}
                width={DEMO_GIF_WIDTH}
                height={DEMO_GIF_HEIGHT}
                decoding="async"
                className="block h-auto w-full"
              />
              <span className="sr-only">. Opens the live demo at {demoFrame.urlLabel}.</span>
            </a>

            <p className="border-t border-[var(--t-border-subtle)] bg-[var(--t-surface)] px-4 py-2.5 text-left text-[12px] text-[var(--t-ink-muted)] lg:hidden">
              {demoFrame.statusLabel}
            </p>
          </div>
        </SurfaceTile>

        <figcaption className="mx-auto mt-5 flex max-w-3xl flex-col items-center gap-2 text-center text-[13.5px] text-pretty text-[var(--t-ink-muted)] sm:flex-row sm:justify-center sm:gap-3">
          <span>{demoFrame.caption}</span>
          <ChevronLink href={links.liveDemoUrl} external event="demo_caption_cta_clicked">
            {demoFrame.urlLabel}
          </ChevronLink>
        </figcaption>
      </figure>
    </section>
  )
}
