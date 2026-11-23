import { ArrowUpRight, Info } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { LinkButton } from '@/components/link-button'
import { Section, SectionHeader } from '@/components/section'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { siteCopy, type QuickstartStep, type QuickstartTab } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'

const isUrl = (command: string) => /^https?:\/\//.test(command)

// The Helm steps carry their own "1. " prefix in the copy; the step marker
// already shows the number, so the prefix is dropped for display.
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
          <span className="flex items-center gap-2.5 text-[13px] text-foreground">
            <span
              aria-hidden="true"
              className="flex size-5 shrink-0 items-center justify-center rounded-full bg-primary/20 font-mono text-[11px] font-medium text-brand"
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
            <LinkButton
              link={{ href: step.command, external: true }}
              event="quickstart_live_demo_opened"
              variant="ghost"
              size="sm"
            >
              Open
              <ArrowUpRight aria-hidden="true" />
            </LinkButton>
          ) : undefined
        }
      />
      {step.note ? (
        <p className="flex items-start gap-2 px-1 text-[13px] leading-snug text-pretty text-muted-foreground">
          <Info aria-hidden="true" className="mt-0.5 size-3.5 shrink-0" />
          <span className="min-w-0 break-words">{step.note}</span>
        </p>
      ) : null}
    </li>
  )
}

export function QuickstartSection() {
  const { quickstart } = siteCopy

  return (
    <Section id={quickstart.id} aria-labelledby="quickstart-heading">
      <div className="grid gap-12 lg:grid-cols-[minmax(0,5fr)_minmax(0,8fr)] lg:gap-14">
        <SectionHeader
          headingId="quickstart-heading"
          eyebrow={quickstart.eyebrow}
          title={quickstart.title}
          subtitle={quickstart.subtitle}
          className="lg:sticky lg:top-24 lg:self-start"
        />

        <Tabs
          defaultValue={quickstart.tabs[0].id}
          onValueChange={(value) => capture(`quickstart_tab_${String(value)}`)}
          className="min-w-0 gap-6"
        >
          <TabsList aria-label="Quickstart methods" className="grid w-full grid-cols-2 sm:inline-flex sm:w-fit">
            {quickstart.tabs.map((tab) => (
              <TabsTrigger key={tab.id} value={tab.id}>
                {tab.label}
                <Badge variant="outline" size="sm" className="font-mono">
                  {tab.badge}
                </Badge>
              </TabsTrigger>
            ))}
          </TabsList>

          {quickstart.tabs.map((tab) => (
            <TabsContent key={tab.id} value={tab.id} className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <h3 className="text-lg font-semibold tracking-tight text-balance">{tab.title}</h3>
                <p className="text-sm leading-relaxed text-pretty text-muted-foreground">
                  {tab.description}
                </p>
              </div>
              <ol className="flex flex-col gap-5">
                {tab.steps.map((step, index) => (
                  <QuickstartStepBlock key={step.label} step={step} position={index + 1} tab={tab} />
                ))}
              </ol>
            </TabsContent>
          ))}
        </Tabs>
      </div>
    </Section>
  )
}
