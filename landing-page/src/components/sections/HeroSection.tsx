import { ArrowRight, ArrowUpRight, Star, Terminal } from 'lucide-react'
import { CopyButton } from '@/components/copy-button'
import { LinkButton } from '@/components/link-button'
import { Badge } from '@/components/ui/badge'
import { siteCopy } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'

// Intrinsic size of public/demo.gif, so the frame reserves its space before the GIF loads.
const DEMO_GIF_WIDTH = 900
const DEMO_GIF_HEIGHT = 563

export function HeroSection() {
  const { hero, links } = siteCopy
  const { announcement, quickCommand, demoFrame } = hero

  return (
    <section aria-labelledby="hero-heading" className="relative isolate overflow-hidden">
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 -z-10">
        <div className="hero-grid absolute inset-x-0 top-0 h-[46rem]" />
        <div className="hero-glow absolute inset-x-0 top-0 h-[38rem]" />
      </div>

      <div className="page-container flex flex-col items-center pt-14 pb-20 text-center sm:pt-24 sm:pb-28">
        <a
          href={announcement.href}
          target="_blank"
          rel="noreferrer"
          onClick={() => capture('release_notes_opened')}
          className="group inline-flex max-w-full items-center gap-2.5 rounded-full border border-border/60 bg-card/70 py-1 pr-3.5 pl-1 text-left text-[13px] text-muted-foreground backdrop-blur-sm transition-colors hover:border-border hover:text-foreground sm:text-sm"
        >
          <Badge variant="primary" className="font-mono">
            {announcement.badge}
          </Badge>
          <span>{announcement.text}</span>
          <ArrowRight
            aria-hidden="true"
            className="size-3.5 shrink-0 transition-transform duration-150 group-hover:translate-x-0.5"
          />
        </a>

        <h1
          id="hero-heading"
          className="mt-8 text-[clamp(2.375rem,5.2vw,3.75rem)] leading-[1.04] font-semibold tracking-[-0.035em] text-balance"
        >
          {hero.headline} <span className="block text-brand">{hero.headlineAccent}</span>
        </h1>

        <p className="mt-6 max-w-2xl text-base leading-relaxed text-pretty text-muted-foreground sm:text-lg">
          {hero.subheadline}
        </p>

        <div className="mt-9 flex w-full flex-col items-stretch justify-center gap-3 sm:w-auto sm:flex-row sm:items-center">
          <LinkButton link={hero.primaryCta} event="live_demo_cta_clicked" size="lg">
            {hero.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </LinkButton>
          <LinkButton
            link={hero.secondaryCta}
            event="hero_github_cta_clicked"
            variant="outline"
            size="lg"
          >
            <Star aria-hidden="true" />
            {hero.secondaryCta.label}
          </LinkButton>
          <LinkButton
            link={hero.tertiaryCta}
            event="quickstart_cta_clicked"
            variant="ghost"
            size="lg"
          >
            <Terminal aria-hidden="true" />
            {hero.tertiaryCta.label}
          </LinkButton>
        </div>

        <div className="mt-8 w-full max-w-4xl">
          <div className="flex items-center gap-2 rounded-lg border border-border/60 bg-surface/80 py-1 pr-1 pl-3.5 text-left backdrop-blur-sm">
            <span aria-hidden="true" className="font-mono text-[13px] text-muted-foreground select-none">
              $
            </span>
            <code
              className="scrollbar-none min-w-0 flex-1 overflow-x-auto py-1.5 font-mono text-[13px] whitespace-nowrap text-foreground/90"
            >
              {quickCommand.command}
            </code>
            <CopyButton
              value={quickCommand.command}
              label={quickCommand.label}
              event="hero_quick_command_copied"
            />
          </div>
          <p className="mt-2.5 text-xs text-muted-foreground">{quickCommand.label}</p>
        </div>

        <figure className="relative mt-14 w-full max-w-5xl sm:mt-20">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute -inset-x-6 -top-12 bottom-8 -z-10 rounded-[3rem] bg-primary/20 blur-3xl"
          />
          <div className="relative overflow-hidden rounded-xl border border-border bg-card shadow-2xl shadow-black/60">
            <div
              aria-hidden="true"
              className="absolute inset-x-0 top-0 z-10 h-px bg-linear-to-r from-transparent via-brand/70 to-transparent"
            />
            <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-3 border-b border-border/60 bg-muted/50 px-3.5 py-2.5">
              <div aria-hidden="true" className="flex gap-1.5">
                <span className="size-2.5 rounded-full bg-foreground/15" />
                <span className="size-2.5 rounded-full bg-foreground/15" />
                <span className="size-2.5 rounded-full bg-foreground/15" />
              </div>
              <div className="flex min-w-0 items-center gap-2 rounded-md border border-border/60 bg-background/70 px-3 py-1 font-mono text-xs text-foreground/90">
                <span aria-hidden="true" className="relative flex size-2 shrink-0">
                  <span className="absolute inline-flex size-full animate-ping rounded-full bg-success/70 motion-reduce:hidden" />
                  <span className="relative inline-flex size-2 rounded-full bg-success" />
                </span>
                <span className="truncate">{demoFrame.urlLabel}</span>
              </div>
              <p className="hidden justify-self-end text-xs text-muted-foreground lg:block">
                {demoFrame.statusLabel}
              </p>
            </div>
            <a
              href={links.liveDemoUrl}
              target="_blank"
              rel="noreferrer"
              onClick={() => capture('demo_gif_clicked')}
              className="block bg-surface focus-visible:-outline-offset-2"
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
            <p className="border-t border-border/60 px-3.5 py-2.5 text-left text-xs text-muted-foreground lg:hidden">
              {demoFrame.statusLabel}
            </p>
          </div>
          <figcaption className="mx-auto mt-5 flex max-w-3xl flex-col items-center gap-x-4 gap-y-2 text-sm text-pretty text-muted-foreground">
            <span>{demoFrame.caption}</span>
            <a
              href={links.liveDemoUrl}
              target="_blank"
              rel="noreferrer"
              onClick={() => capture('demo_caption_cta_clicked')}
              className="inline-flex items-center gap-1 rounded-sm font-medium text-brand underline-offset-4 hover:underline"
            >
              {demoFrame.urlLabel}
              <ArrowUpRight aria-hidden="true" className="size-3.5" />
            </a>
          </figcaption>
        </figure>
      </div>
    </section>
  )
}
