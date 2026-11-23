import type { ReactNode } from 'react'
import { CopyButton } from '@/components/copy-button'
import { tokenize, type CodeLanguage, type CodeTokenKind } from '@/lib/highlight'
import { cn } from '@/lib/utils'

const TOKEN_CLASS: Record<CodeTokenKind, string | undefined> = {
  key: 'text-[var(--t-blue)]',
  flag: 'text-[var(--t-blue)]',
  string: 'text-[var(--t-chip-mint-fg)]',
  variable: 'text-[var(--t-chip-warning-fg)]',
  punctuation: 'text-[var(--t-ink-muted)]',
  plain: undefined,
}

interface CodeBlockProps {
  readonly code: string
  readonly language?: CodeLanguage
  /** Header content: a file path, a step label, or a short caption. */
  readonly title: ReactNode
  /** What the copy button copies, for its accessible name. */
  readonly copyLabel: string
  readonly copyEvent: string
  /** Extra header controls rendered before the copy button. */
  readonly actions?: ReactNode
  readonly className?: string
}

export function CodeBlock({
  code,
  language = 'text',
  title,
  copyLabel,
  copyEvent,
  actions,
  className,
}: CodeBlockProps) {
  const isSingleLine = !code.includes('\n')

  return (
    <div
      className={cn(
        'min-w-0 overflow-hidden rounded-[18px] border border-[var(--t-border-subtle)] bg-[var(--t-elevated)] text-left',
        className,
      )}
    >
      <div className="flex min-h-11 items-center gap-2 border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)]/75 py-1.5 pr-2 pl-4">
        <div className="min-w-0 flex-1 truncate text-[12.5px] font-medium text-[var(--t-ink-muted)]">
          {title}
        </div>
        {actions}
        <CopyButton value={code} label={copyLabel} event={copyEvent} />
      </div>
      <pre
        className={cn(
          'scrollbar-subtle p-4 font-mono text-[12.5px] leading-[1.65] text-[var(--t-ink)] sm:text-[13px]',
          isSingleLine ? 'wrap-break-word whitespace-pre-wrap' : 'overflow-x-auto',
        )}
      >
        <code>
          {tokenize(code, language).map((token, index) => (
            <span key={index} className={TOKEN_CLASS[token.kind]}>
              {token.text}
            </span>
          ))}
        </code>
      </pre>
    </div>
  )
}
