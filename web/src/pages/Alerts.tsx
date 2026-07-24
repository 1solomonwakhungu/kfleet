import { useCallback, useEffect, useMemo, useState } from 'react'
import { AlertTriangle, BellRing, Check, CircleAlert, LoaderCircle, RefreshCw, type LucideIcon } from 'lucide-react'

import { api } from '../lib/api'
import type { Alert, AlertDeliveryStatus, AlertStatus } from '../types/alert'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent } from '../components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table'
import { cn } from '../lib/utils'

const timestampFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function formatTimestamp(value?: string) {
  if (!value) return 'Not available'
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? 'Not available' : timestampFormatter.format(timestamp)
}

const deliveryLabels: Record<AlertDeliveryStatus, string> = {
  pending: 'Pending',
  retrying: 'Retrying',
  delivered: 'Delivered',
  dead_letter: 'Dead letter',
  disabled: 'Disabled',
}

const statusLabels: Record<AlertStatus, string> = {
  firing: 'Firing',
  acknowledged: 'Acknowledged',
  resolved: 'Resolved',
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [loadError, setLoadError] = useState('')
  const [acknowledging, setAcknowledging] = useState<ReadonlySet<string>>(new Set())
  const [actionErrors, setActionErrors] = useState<Readonly<Record<string, string>>>({})

  const load = useCallback(async (signal?: AbortSignal, background = false) => {
    if (background) setRefreshing(true)
    else setLoading(true)
    try {
      const nextAlerts = await api.listAlerts(undefined, signal)
      setAlerts(nextAlerts)
      setLoadError('')
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return
      setLoadError(error instanceof Error ? error.message : 'Failed to load alert history')
    } finally {
      if (background) setRefreshing(false)
      else setLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    void load(controller.signal)
    return () => controller.abort()
  }, [load])

  const acknowledge = useCallback(async (alert: Alert) => {
    setAcknowledging((current) => new Set(current).add(alert.id))
    setActionErrors((current) => {
      const next = { ...current }
      delete next[alert.id]
      return next
    })
    try {
      const updated = await api.acknowledgeAlert(alert.id)
      setAlerts((current) => current.map((item) => item.id === updated.id ? updated : item))
    } catch (error) {
      setActionErrors((current) => ({
        ...current,
        [alert.id]: error instanceof Error ? error.message : 'Failed to acknowledge alert',
      }))
    } finally {
      setAcknowledging((current) => {
        const next = new Set(current)
        next.delete(alert.id)
        return next
      })
    }
  }, [])

  const summary = useMemo(() => ({
    firing: alerts.filter((alert) => alert.status === 'firing').length,
    acknowledged: alerts.filter((alert) => alert.status === 'acknowledged').length,
    deadLetter: alerts.filter((alert) => alert.deliveryStatus === 'dead_letter').length,
  }), [alerts])

  return (
    <main className="mx-auto w-full max-w-[96rem] px-4 py-8 sm:px-6 lg:px-8">
      <header className="flex flex-col justify-between gap-5 border-b border-border pb-7 sm:flex-row sm:items-end">
        <div>
          <div className="flex items-center gap-2 font-mono text-xs font-semibold uppercase tracking-[0.14em] text-accent">
            <BellRing className="size-4" aria-hidden="true" />
            Operations
          </div>
          <h1 className="mt-3 font-display text-3xl font-bold tracking-tight sm:text-4xl">Fleet alerts</h1>
          <p className="mt-2 max-w-2xl text-muted">
            Health alert history, acknowledgement state, and durable webhook delivery outcomes.
          </p>
        </div>
        <Button
          variant="outline"
          disabled={loading || refreshing}
          onClick={() => void load(undefined, true)}
        >
          <RefreshCw className={cn('size-4', refreshing && 'animate-spin')} aria-hidden="true" />
          Refresh
        </Button>
      </header>

      <section className="grid gap-3 py-6 sm:grid-cols-3" aria-label="Alert summary">
        <SummaryCard label="Needs acknowledgement" value={summary.firing} icon={CircleAlert} tone="danger" />
        <SummaryCard label="Acknowledged" value={summary.acknowledged} icon={Check} tone="accent" />
        <SummaryCard label="Dead letter" value={summary.deadLetter} icon={AlertTriangle} tone="warning" />
      </section>

      {loadError && (
        <div className="mb-5 rounded-md border border-red-900 bg-red-950/40 p-4 text-sm text-red-200" role="alert">
          {loadError}
        </div>
      )}

      {loading ? (
        <Card className="grid min-h-64 place-items-center ring-1 ring-inset ring-border">
          <LoaderCircle className="size-7 animate-spin text-accent" aria-label="Loading alert history" />
        </Card>
      ) : alerts.length === 0 ? (
        <Card className="ring-1 ring-inset ring-border">
          <CardContent className="grid min-h-64 place-items-center p-6 text-center">
            <div>
              <span className="mx-auto grid size-12 place-items-center rounded-full bg-blue-950 text-blue-300 ring-1 ring-inset ring-blue-800">
                <BellRing className="size-6" aria-hidden="true" />
              </span>
              <p className="mt-4 font-display text-xl font-bold">No fleet health alerts</p>
              <p className="mt-2 text-muted">Degraded and unreachable cluster events will appear here.</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <Card className="overflow-hidden ring-1 ring-inset ring-border">
          <Table aria-label="Fleet alert history">
            <TableHeader>
              <TableRow>
                <TableHead scope="col">Alert</TableHead>
                <TableHead scope="col">State</TableHead>
                <TableHead scope="col">Delivery</TableHead>
                <TableHead scope="col">Triggered</TableHead>
                <TableHead scope="col" className="text-right">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {alerts.map((alert) => {
                const isAcknowledging = acknowledging.has(alert.id)
                const actionError = actionErrors[alert.id]
                return (
                  <TableRow key={alert.id}>
                    <TableCell>
                      <div className="min-w-64">
                        <div className="flex flex-wrap items-center gap-2">
                          <Badge className={alert.severity === 'critical'
                            ? 'bg-red-950 text-red-200 ring-1 ring-inset ring-red-800'
                            : 'bg-amber-950 text-amber-200 ring-1 ring-inset ring-amber-800'}
                          >
                            {alert.severity}
                          </Badge>
                          <p className="font-semibold text-foreground">{alert.summary}</p>
                        </div>
                        <p className="mt-2 text-xs text-muted">{alert.ruleName}</p>
                        <p className="mt-1 max-w-72 truncate font-mono text-xs text-muted" title={alert.id}>{alert.id}</p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge className={cn(
                        'ring-1 ring-inset',
                        alert.status === 'firing' && 'bg-red-950 text-red-200 ring-red-800',
                        alert.status === 'acknowledged' && 'bg-blue-950 text-blue-200 ring-blue-800',
                        alert.status === 'resolved' && 'bg-emerald-950 text-emerald-200 ring-emerald-800',
                      )}>
                        {statusLabels[alert.status]}
                      </Badge>
                      {alert.acknowledgedBy && (
                        <p className="mt-2 text-xs text-muted">by {alert.acknowledgedBy}</p>
                      )}
                    </TableCell>
                    <TableCell>
                      <p className="text-sm font-medium">{deliveryLabels[alert.deliveryStatus]}</p>
                      <p className="mt-1 text-xs text-muted">
                        {alert.deliveryAttempts} {alert.deliveryAttempts === 1 ? 'attempt' : 'attempts'}
                      </p>
                      {alert.lastDeliveryError && (
                        <p className="mt-1 max-w-64 truncate text-xs text-danger" title={alert.lastDeliveryError}>
                          {alert.lastDeliveryError}
                        </p>
                      )}
                    </TableCell>
                    <TableCell>
                      <time dateTime={alert.triggeredAt} title={alert.triggeredAt} className="whitespace-nowrap text-sm">
                        {formatTimestamp(alert.triggeredAt)}
                      </time>
                    </TableCell>
                    <TableCell className="min-w-48 text-right">
                      {alert.status === 'firing' ? (
                        <Button
                          size="sm"
                          disabled={isAcknowledging}
                          onClick={() => void acknowledge(alert)}
                        >
                          {isAcknowledging
                            ? <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
                            : <Check className="size-4" aria-hidden="true" />}
                          {isAcknowledging ? 'Acknowledging...' : 'Acknowledge'}
                        </Button>
                      ) : (
                        <span className="text-sm text-muted">{statusLabels[alert.status]}</span>
                      )}
                      {actionError && <p className="mt-2 text-xs text-danger" role="alert">{actionError}</p>}
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </Card>
      )}
    </main>
  )
}

function SummaryCard({
  label,
  value,
  icon: Icon,
  tone,
}: {
  label: string
  value: number
  icon: LucideIcon
  tone: 'danger' | 'accent' | 'warning'
}) {
  return (
    <Card className="p-4 ring-1 ring-inset ring-border">
      <div className="flex items-center justify-between gap-4">
        <div>
          <p className="text-sm text-muted">{label}</p>
          <p className="mt-1 font-display text-3xl font-bold">{value}</p>
        </div>
        <Icon className={cn(
          'size-6',
          tone === 'danger' && 'text-red-400',
          tone === 'accent' && 'text-blue-400',
          tone === 'warning' && 'text-amber-400',
        )} aria-hidden="true" />
      </div>
    </Card>
  )
}
