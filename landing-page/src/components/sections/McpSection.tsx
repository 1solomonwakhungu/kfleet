import * as React from 'react'
import { Check, Copy } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { Section, SectionHeader } from '@/components/section'
import { TonalChip } from '@/components/tonal/chip'
import { SurfaceTile } from '@/components/tonal/surface-tile'
import { siteCopy } from '@/content/siteCopy'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import type { CodeLanguage } from '@/lib/highlight'
import { capture } from '@/lib/posthog'
import { cn } from '@/lib/utils'

interface McpClientTab {
  readonly id: string
  readonly label: string
  readonly blockTitle: string
  readonly language: CodeLanguage
  readonly code: string
}

const CLIENT_TABS: readonly McpClientTab[] = [
  {
    id: 'claude_code',
    label: 'Claude Code CLI',
    blockTitle: 'Terminal',
    language: 'shell',
    code: siteCopy.mcp.claudeCodeCommand,
  },
  {
    id: 'config_json',
    label: 'Claude Desktop / Cursor JSON',
    blockTitle: 'MCP client configuration',
    language: 'json',
    code: siteCopy.mcp.demoConfigJson,
  },
  {
    id: 'stdio',
    label: 'Stdio CLI',
    blockTitle: 'Terminal',
    language: 'shell',
    code: siteCopy.mcp.demoCommand,
  },
]

function SamplePrompt({ prompt }: { readonly prompt: string }) {
  const { copied, copy } = useCopyToClipboard('mcp_prompt_copied')

  return (
    <li>
      <button
        type="button"
        onClick={() => void copy(prompt)}
        className="flex w-full cursor-pointer items-start gap-3 rounded-[16px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] px-4 py-3 text-left text-[14px] leading-snug text-[var(--t-ink)] transition-colors duration-200 hover:border-[var(--t-blue)] hover:bg-[var(--t-surface)]"
      >
        <span className="sr-only">Copy prompt: </span>
        <span className="flex-1">{prompt}</span>
        <span
          aria-hidden="true"
          className="mt-0.5 flex shrink-0 items-center gap-1 text-[12px] text-[var(--t-ink-muted)] [&_svg]:size-3.5"
        >
          {copied ? (
            <>
              <Check className="text-[var(--t-chip-mint-fg)]" />
              <span className="font-medium text-[var(--t-chip-mint-fg)]">Copied</span>
            </>
          ) : (
            <Copy />
          )}
        </span>
      </button>
      <span role="status" className="sr-only">
        {copied ? 'Prompt copied' : ''}
      </span>
    </li>
  )
}

export function McpSection() {
  const { mcp } = siteCopy
  const [activeId, setActiveId] = React.useState(CLIENT_TABS[0].id)
  const activeTab = CLIENT_TABS.find((tab) => tab.id === activeId) ?? CLIENT_TABS[0]

  return (
    <Section id={mcp.id} aria-labelledby="mcp-heading">
      <div className="grid gap-10 lg:grid-cols-2 lg:gap-12">
        <div className="flex min-w-0 flex-col gap-7">
          <SectionHeader
            headingId="mcp-heading"
            eyebrow={mcp.eyebrow}
            title={mcp.title}
            subtitle={mcp.subtitle}
          />

          <SurfaceTile tone="cool" padding="md" className="flex flex-col gap-5">
            <div role="tablist" aria-label="MCP client" className="flex flex-wrap gap-2">
              {CLIENT_TABS.map((tab) => {
                const active = tab.id === activeTab.id
                return (
                  <button
                    key={tab.id}
                    type="button"
                    role="tab"
                    aria-selected={active}
                    onClick={() => {
                      setActiveId(tab.id)
                      capture(`mcp_tab_${tab.id}`)
                    }}
                    className={cn(
                      'cursor-pointer rounded-full px-4 py-2 text-[13px] font-medium transition-all duration-200 [transition-timing-function:var(--t-ease)]',
                      active
                        ? 'bg-[var(--t-ink)] text-[var(--t-bg)]'
                        : 'border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] text-[var(--t-ink-secondary)] hover:bg-[var(--t-blue-tint)] hover:text-[var(--t-ink)]',
                    )}
                  >
                    {tab.label}
                  </button>
                )
              })}
            </div>

            <div role="tabpanel" aria-label={activeTab.label}>
              <CodeBlock
                code={activeTab.code}
                language={activeTab.language}
                title={activeTab.blockTitle}
                copyLabel={`${activeTab.label} setup`}
                copyEvent="mcp_command_copied"
              />
            </div>
          </SurfaceTile>

          <SurfaceTile tone="neutral" padding="md" className="flex flex-col gap-4">
            <div>
              <h3 className="text-[15px] font-medium tracking-[-0.01em] text-[var(--t-ink)]">
                Prompts to try
              </h3>
              <p className="mt-1 text-[13.5px] text-[var(--t-ink-muted)]">
                Select a prompt to copy it.
              </p>
            </div>
            <ul className="flex flex-col gap-2.5">
              {mcp.samplePrompts.map((prompt) => (
                <SamplePrompt key={prompt} prompt={prompt} />
              ))}
            </ul>
          </SurfaceTile>
        </div>

        <SurfaceTile
          tone="white"
          bordered
          padding="none"
          className="min-w-0 self-start overflow-hidden"
        >
          <div className="flex items-center justify-between gap-3 border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)] px-6 py-4">
            <h3 className="text-[16px] font-medium tracking-[-0.015em] text-[var(--t-ink)]">
              {mcp.tools.length} read-only MCP tools
            </h3>
            <TonalChip tone="blue">stdio</TonalChip>
          </div>
          <ul className="divide-y divide-[var(--t-border-subtle)]">
            {mcp.tools.map((tool) => (
              <li
                key={tool.name}
                className="flex flex-col gap-1.5 px-6 py-4 transition-colors duration-150 hover:bg-[var(--t-surface)]/70"
              >
                <div className="flex items-center justify-between gap-3">
                  <code className="font-mono text-[13px] font-medium text-[var(--t-ink)]">
                    {tool.name}
                  </code>
                  <TonalChip tone="neutral" className="px-2.5 py-0.5 text-[11px]">
                    {tool.category}
                  </TonalChip>
                </div>
                <p className="text-[13.5px] leading-[1.5] text-pretty text-[var(--t-ink-secondary)]">
                  {tool.summary}
                </p>
              </li>
            ))}
          </ul>
        </SurfaceTile>
      </div>
    </Section>
  )
}
