import { useEffect, useId, useRef, type ReactNode } from 'react'
import { AlertTriangle } from 'lucide-react'

import { Button } from '../ui/button'

interface ConfirmDialogProps {
  open: boolean
  title: string
  confirmLabel: string
  pending?: boolean
  error?: string | null
  children: ReactNode
  onConfirm: () => void
  onCancel: () => void
}

/**
 * Modal confirmation for destructive actions. Consequences are supplied by the
 * caller so each dialog can state exactly what will be lost.
 */
export function ConfirmDialog({
  open,
  title,
  confirmLabel,
  pending = false,
  error,
  children,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  const titleId = useId()
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (!open) return
    cancelRef.current?.focus()

    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !pending) onCancel()
    }
    document.addEventListener('keydown', closeOnEscape)
    return () => document.removeEventListener('keydown', closeOnEscape)
  }, [open, pending, onCancel])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-background/80 px-4 py-8">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="w-full max-w-lg rounded-lg border border-border bg-surface p-5 shadow-lg"
      >
        <div className="flex items-start gap-3">
          <span className="mt-0.5 grid size-9 shrink-0 place-items-center rounded-full bg-danger-soft text-danger">
            <AlertTriangle className="size-5" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <h2 id={titleId} className="font-display text-lg font-bold">
              {title}
            </h2>
            <div className="mt-2 space-y-2 text-sm text-muted">{children}</div>
          </div>
        </div>

        {error && (
          <p className="mt-4 rounded-md bg-danger-soft p-3 text-sm text-danger" role="alert">
            {error}
          </p>
        )}

        <div className="mt-5 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button ref={cancelRef} variant="outline" size="sm" disabled={pending} onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="danger" size="sm" disabled={pending} onClick={onConfirm}>
            {pending ? 'Working…' : confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
