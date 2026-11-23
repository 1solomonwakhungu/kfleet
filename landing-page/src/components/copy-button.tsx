import { Check, Copy } from 'lucide-react'
import { TonalButton } from '@/components/tonal/button'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { cn } from '@/lib/utils'

interface CopyButtonProps {
  /** Text placed on the clipboard. */
  readonly value: string
  /** What is being copied, e.g. "Docker command". Used for the accessible name. */
  readonly label: string
  /** PostHog event captured after a successful copy. */
  readonly event: string
  readonly className?: string
}

export function CopyButton({ value, label, event, className }: CopyButtonProps) {
  const { copied, copy } = useCopyToClipboard(event)

  return (
    <>
      <TonalButton
        type="button"
        variant="ghost"
        size="icon"
        aria-label={`Copy ${label}`}
        className={cn(
          copied && 'text-[var(--t-chip-mint-fg)] hover:text-[var(--t-chip-mint-fg)]',
          className,
        )}
        onClick={() => void copy(value)}
      >
        {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
      </TonalButton>
      <span role="status" className="sr-only">
        {copied ? `Copied ${label}` : ''}
      </span>
    </>
  )
}
