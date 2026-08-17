import { useCallback, useEffect, useRef, useState } from 'react'

import { adminApi } from '../lib/adminApi'
import type { AuditEvent } from '../types/admin'
import { isAbortError, messageFrom } from '../lib/errors'

/** The hub rejects limits above 1000 (internal/server/handlers_audit.go). */
export const maxAuditLimit = 1000
export const auditPageSize = 100

export interface AuditEventsState {
  events: AuditEvent[]
  loading: boolean
  error: string | null
  limit: number
  hasMore: boolean
  reload: () => Promise<void>
  loadMore: () => void
}

/**
 * Loads recent audit events, newest first. The hub exposes a single `limit`
 * parameter, so "load more" widens the requested window rather than paging
 * with a cursor.
 */
export function useAuditEvents(enabled = true): AuditEventsState {
  const [events, setEvents] = useState<AuditEvent[]>([])
  const [loading, setLoading] = useState(enabled)
  const [error, setError] = useState<string | null>(null)
  const [limit, setLimit] = useState(auditPageSize)
  const controllerRef = useRef<AbortController | null>(null)

  const load = useCallback(
    async (requested: number) => {
      if (!enabled) return
      controllerRef.current?.abort()
      const controller = new AbortController()
      controllerRef.current = controller
      setLoading(true)
      setError(null)

      try {
        const loaded = await adminApi.listAuditEvents(requested, controller.signal)
        if (controller.signal.aborted) return
        setEvents(loaded)
      } catch (caught) {
        if (!isAbortError(caught)) setError(messageFrom(caught, 'Audit events could not be loaded.'))
      } finally {
        if (!controller.signal.aborted) setLoading(false)
      }
    },
    [enabled],
  )

  useEffect(() => {
    void load(limit)
    return () => controllerRef.current?.abort()
  }, [load, limit])

  const reload = useCallback(() => load(limit), [load, limit])

  const loadMore = useCallback(() => {
    setLimit((current) => Math.min(current + auditPageSize, maxAuditLimit))
  }, [])

  return {
    events,
    loading,
    error,
    limit,
    hasMore: limit < maxAuditLimit && events.length >= limit,
    reload,
    loadMore,
  }
}
