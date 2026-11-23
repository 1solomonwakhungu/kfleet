import { useCallback, useEffect, useRef, useState } from 'react'
import { capture } from '@/lib/posthog'

const RESET_DELAY_MS = 1800

/**
 * Copies text and reports a short-lived `copied` flag for button feedback.
 * `event` is the PostHog event captured after a successful copy.
 */
export function useCopyToClipboard(event?: string) {
  const [copied, setCopied] = useState(false)
  const resetTimer = useRef<number | undefined>(undefined)

  useEffect(() => () => window.clearTimeout(resetTimer.current), [])

  const copy = useCallback(
    async (value: string) => {
      try {
        await navigator.clipboard.writeText(value)
      } catch {
        // The Clipboard API is missing or denied (insecure context, embedded
        // frame, permissions policy). Leave the UI unchanged so it never
        // claims a copy that did not happen.
        return false
      }

      if (event) {
        capture(event)
      }
      setCopied(true)
      window.clearTimeout(resetTimer.current)
      resetTimer.current = window.setTimeout(() => setCopied(false), RESET_DELAY_MS)
      return true
    },
    [event],
  )

  return { copied, copy }
}
