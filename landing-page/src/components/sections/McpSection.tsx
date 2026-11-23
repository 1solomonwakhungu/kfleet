import { Check, Copy } from 'lucide-react'
import { CodeBlock } from '@/components/code-block'
import { Section, SectionHeader } from '@/components/section'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { siteCopy } from '@/content/siteCopy'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import type { CodeLanguage } from '@/lib/highlight'
import { capture } from '@/lib/posthog'

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
        className="flex w-full items-start gap-3 rounded-lg border border-border/60 bg-card/60 px-3.5 py-3 text-left text-sm leading-snug text-foreground/90 transition-colors hover:border-border hover:bg-accent"
      >
        <span className="sr-only">Copy prompt: </span>
        <span className="flex-1">{prompt}</span>
        <span
          aria-hidden="true"
          className="mt-0.5 flex shrink-0 items-center gap-1 text-xs text-muted-foreground [&_svg]:size-3.5"
        >
          {copied ? (
            <>
              <Check className="text-success" />
              <span className="text-success">Copied</span>
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

  return (
    <Section id={mcp.id} aria-labelledby="mcp-heading">
      <div className="grid gap-12 lg:grid-cols-2 lg:gap-14">
        <div className="flex min-w-0 flex-col gap-10">
          <SectionHeader
            headingId="mcp-heading"
            eyebrow={mcp.eyebrow}
            title={mcp.title}
            subtitle={mcp.subtitle}
          />

          <Tabs
            defaultValue={CLIENT_TABS[0].id}
            onValueChange={(value) => capture(`mcp_tab_${String(value)}`)}
          >
            <TabsList aria-label="MCP client" className="grid w-full grid-cols-1 sm:inline-flex sm:w-fit">
              {CLIENT_TABS.map((tab) => (
                <TabsTrigger key={tab.id} value={tab.id}>
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>
            {CLIENT_TABS.map((tab) => (
              <TabsContent key={tab.id} value={tab.id}>
                <CodeBlock
                  code={tab.code}
                  language={tab.language}
                  title={tab.blockTitle}
                  copyLabel={`${tab.label} setup`}
                  copyEvent="mcp_command_copied"
                />
              </TabsContent>
            ))}
          </Tabs>

          <div>
            <h3 className="text-sm font-medium text-foreground">Prompts to try</h3>
            <p className="mt-1 text-sm text-muted-foreground">Select a prompt to copy it.</p>
            <ul className="mt-4 flex flex-col gap-2">
              {mcp.samplePrompts.map((prompt) => (
                <SamplePrompt key={prompt} prompt={prompt} />
              ))}
            </ul>
          </div>
        </div>

        <Card className="min-w-0 self-start">
          <CardHeader className="pb-4">
            <CardTitle>{mcp.tools.length} read-only MCP tools</CardTitle>
          </CardHeader>
          <CardContent className="px-0 pb-0">
            <ul className="divide-y divide-border/60 border-t border-border/60">
              {mcp.tools.map((tool) => (
                <li key={tool.name} className="flex flex-col gap-1.5 px-6 py-3.5">
                  <div className="flex items-center justify-between gap-3">
                    <code className="font-mono text-[13px] font-medium text-foreground">{tool.name}</code>
                    <Badge variant="outline" size="sm">
                      {tool.category}
                    </Badge>
                  </div>
                  <p className="text-[13px] leading-snug text-pretty text-muted-foreground">
                    {tool.summary}
                  </p>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </div>
    </Section>
  )
}
