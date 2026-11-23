import * as React from 'react'
import { Info } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { Section, SectionHeader } from '@/components/section'
import { ChevronLink } from '@/components/tonal/chevron-link'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import { siteCopy, type QuickstartStep, type QuickstartTab } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'
import { cn } from '@/lib/utils'

const isUrl = (command: string) => /^https?:\/\//.test(command)

const stepLabel = (label: string) => label.replace(/^\d+\.\s*/, '')

function QuickstartStepBlock({
  step,
  position,
  tab,
}: {
  readonly step: QuickstartStep
  readonly position: number
  readonly tab: QuickstartTab
}) {
  const label = stepLabel(step.label)
  const opensInBrowser = isUrl(step.command)

  return (
    <li className="flex flex-col gap-2.5">
      <CodeBlock
        code={step.command}
        language={opensInBrowser ? 'text' : 'shell'}
        title={
          <span className="flex items-center gap-2.5 text-[13px] text-[var(--t-ink)]">
            <span
              aria-hidden="true"
              className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[var(--t-blue-tint)] font-mono text-[11px] font-medium text-[var(--t-blue-hover)]"
            >
              {position}
            </span>
            <span className="truncate">
              <span className="sr-only">Step {position}: </span>
              {label}
            </span>
          </span>
        }
        copyLabel={`${tab.label} step ${position}`}
        copyEvent="quickstart_command_copied"
        actions={
          opensInBrowser ? (
            <ChevronLink
              href={step.command}
              external
              event="quickstart_live_demo_opened"
              className="px-2 text-[13px]"
            >
              Open
            </ChevronLink>
          ) : undefined
        }
      />
      {step.note ? (
        <p className="flex items-start gap-2 px-1 text-[13px] leading-snug text-pretty text-[var(--t-ink-muted)]">
          <Info aria-hidden="true" className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="min-w-0 break-words">{step.note}</span>
        </p>
      ) : null}
    </li>
  )
}

export function QuickstartSection() {
  const { quickstart } = siteCopy
  const [activeId, setActiveId] = React.useState(quickstart.tabs[0].id)
  const activeTab = quickstart.tabs.find((tab) => tab.id === activeId) ?? quickstart.tabs[0]

  return (
    <Section id={quickstart.id} aria-labelledby="quickstart-heading">
      <div className="grid gap-10 lg:grid-cols-[minmax(0,5fr)_minmax(0,8fr)] lg:gap-12">
        <SectionHeader
          headingId="quickstart-heading"
          eyebrow={quickstart.eyebrow}
          title={quickstart.title}
          subtitle={quickstart.subtitle}
          className="lg:sticky lg:top-24 lg:self-start"
        />

        <SurfaceTile tone="neutral" padding="lg" className="flex min-w-0 flex-col gap-6">
          <div role="tablist" aria-label="Quickstart methods" className="flex flex-wrap gap-2">
            {quickstart.tabs.map((tab) => {
              const active = tab.id === activeTab.id
              return (
                <button
                  key={tab.id}
                  type="button"
                  role="tab"
                  aria-selected={active}
                  onClick={() => {
                    setActiveId(tab.id)
                    capture(`quickstart_tab_${tab.id}`)
                  }}
                  className={cn(
                    'inline-flex cursor-pointer items-center gap-2 rounded-full px-4 py-2 text-[13.5px] font-medium transition-all duration-200 [transition-timing-function:var(--t-ease)]',
                    active
                      ? 'bg-[var(--t-ink)] text-[var(--t-bg)]'
                      : 'border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] text-[var(--t-ink-secondary)] hover:bg-[var(--t-blue-tint)] hover:text-[var(--t-ink)]',
                  )}
                >
                  <span>{tab.label}</span>
                  <span
                    className={cn(
                      'rounded-full px-2 py-0.5 font-mono text-[11px]',
                      active
                        ? 'bg-[var(--t-bg)]/18 text-[var(--t-bg)]'
                        : 'bg-[var(--t-surface)] text-[var(--t-ink-muted)]',
                    )}
                  >
                    {tab.badge}
                  </span>
                </button>
              )
            })}
          </div>

          <div role="tabpanel" aria-label={activeTab.label} className="flex flex-col gap-5">
            <div className="flex flex-col gap-1.5">
              <h3 className="t-headline text-[22px] text-balance">{activeTab.title}</h3>
              <p className="t-body text-[14.5px] text-pretty">{activeTab.description}</p>
            </div>
            <ol className="flex flex-col gap-5">
              {activeTab.steps.map((step, index) => (
                <QuickstartStepBlock
                  key={step.label}
                  step={step}
                  position={index + 1}
                  tab={activeTab}
                />
              ))}
            </ol>
          </div>
        </SurfaceTile>
      </div>
    </Section>
  )
}
