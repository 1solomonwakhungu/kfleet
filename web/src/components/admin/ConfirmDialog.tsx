import { useEffect, useRef, type ReactNode } from 'react'
import { Dialog, Flash } from '@primer/react'

import styles from './ConfirmDialog.module.css'

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
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (!open) return
    cancelRef.current?.focus()
  }, [open])

  if (!open) return null

  return (
    <Dialog
      title={title}
      onClose={() => {
        if (!pending) onCancel()
      }}
      footerButtons={[
        { buttonType: 'default', content: 'Cancel', disabled: pending, onClick: onCancel, ref: cancelRef },
        {
          buttonType: 'danger',
          content: pending ? 'Working…' : confirmLabel,
          disabled: pending,
          onClick: onConfirm,
        },
      ]}
    >
      <div className={styles.body}>{children}</div>
      {error && (
        <Flash variant="danger" role="alert" className={styles.error}>
          {error}
        </Flash>
      )}
    </Dialog>
  )
}
