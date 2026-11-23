import type { ReactNode } from 'react'
import { CopyButton } from '@/components/copy-button'
import { tokenize, type CodeLanguage, type CodeTokenKind } from '@/lib/highlight'
import { cn } from '@/lib/utils'

const TOKEN_CLASS: Record<CodeTokenKind, string | undefined> = {
  key: 'text-brand',
  flag: 'text-brand',
  string: 'text-success/90',
  variable: 'text-warning/90',
  punctuation: 'text-muted-foreground/70',
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
  // A one-line command wraps so it is readable without scrolling. Multi-line
  // snippets keep their line breaks and indentation and scroll sideways.
  const isSingleLine = !code.includes('\n')

  return (
    <div
      className={cn(
        'min-w-0 overflow-hidden rounded-lg border border-border/60 bg-surface text-left',
        className,
      )}
    >
      <div className="flex min-h-10 items-center gap-2 border-b border-border/60 py-1 pr-1 pl-3.5">
        <div className="min-w-0 flex-1 truncate text-xs text-muted-foreground">{title}</div>
        {actions}
        <CopyButton value={code} label={copyLabel} event={copyEvent} />
      </div>
      <pre
        className={cn(
          'scrollbar-subtle p-3.5 font-mono text-xs leading-relaxed text-foreground/90 sm:text-[13px]',
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
