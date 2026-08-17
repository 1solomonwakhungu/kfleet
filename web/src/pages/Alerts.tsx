import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, Flash, Heading, Label, Spinner, Text } from '@primer/react'
import { Blankslate } from '@primer/react/experimental'
import { AlertIcon, BellIcon, CheckIcon, StopIcon, SyncIcon, type Icon } from '@primer/octicons-react'

import { api } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import type { Alert, AlertDeliveryStatus, AlertStatus } from '../types/alert'
import layout from '../styles/layout.module.css'
import styles from './Alerts.module.css'

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

const statusVariants: Record<AlertStatus, 'danger' | 'accent' | 'success'> = {
  firing: 'danger',
  acknowledged: 'accent',
  resolved: 'success',
}

export default function AlertsPage() {
  const { user } = useAuth()
  const canAcknowledge = user?.role === 'admin' || user?.role === 'operator'
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
    if (!canAcknowledge) return
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
  }, [canAcknowledge])

  const summary = useMemo(() => ({
    firing: alerts.filter((alert) => alert.status === 'firing').length,
    acknowledged: alerts.filter((alert) => alert.status === 'acknowledged').length,
    deadLetter: alerts.filter((alert) => alert.deliveryStatus === 'dead_letter').length,
  }), [alerts])

  return (
    <main className={layout.page}>
      <header className={layout.pageHeader}>
        <div className={layout.pageHeaderText}>
          <Heading as="h1" variant="large">
            Fleet alerts
          </Heading>
          <Text className={layout.pageDescription}>
            Health alert history, acknowledgement state, and durable webhook delivery outcomes.
          </Text>
        </div>
        <Button leadingVisual={SyncIcon} disabled={loading || refreshing} onClick={() => void load(undefined, true)}>
          Refresh
        </Button>
      </header>

      <section className={`${layout.grid} ${styles.summary}`} aria-label="Alert summary">
        <SummaryCard label="Needs acknowledgement" value={summary.firing} icon={StopIcon} tone="danger" />
        <SummaryCard label="Acknowledged" value={summary.acknowledged} icon={CheckIcon} tone="accent" />
        <SummaryCard label="Dead letter" value={summary.deadLetter} icon={AlertIcon} tone="attention" />
      </section>

      {loadError && (
        <Flash variant="danger" role="alert" className={styles.flash}>
          {loadError}
        </Flash>
      )}

      {loading ? (
        <div className={`${layout.box} ${styles.loading}`}>
          <Spinner aria-label="Loading alert history" />
        </div>
      ) : alerts.length === 0 ? (
        <div className={layout.box}>
          <Blankslate>
            <Blankslate.Visual>
              <BellIcon size={24} />
            </Blankslate.Visual>
            <Blankslate.Heading as="h2">No fleet health alerts</Blankslate.Heading>
            <Blankslate.Description>
              Degraded and unreachable cluster events will appear here.
            </Blankslate.Description>
          </Blankslate>
        </div>
      ) : (
        <div className={layout.box}>
          <div className={layout.tableScroll}>
            <table className={layout.table} aria-label="Fleet alert history">
              <thead>
                <tr>
                  <th scope="col">Alert</th>
                  <th scope="col">State</th>
                  <th scope="col">Delivery</th>
                  <th scope="col">Triggered</th>
                  <th scope="col" className={styles.actionColumn}>
                    Action
                  </th>
                </tr>
              </thead>
              <tbody>
                {alerts.map((alert) => {
                  const isAcknowledging = acknowledging.has(alert.id)
                  const actionError = actionErrors[alert.id]

                  return (
                    <tr key={alert.id}>
                      <td>
                        <div className={styles.alertCell}>
                          <div className={styles.alertTitle}>
                            <Label variant={alert.severity === 'critical' ? 'danger' : 'attention'}>
                              {alert.severity}
                            </Label>
                            <Text weight="semibold">{alert.summary}</Text>
                          </div>
                          <span className={styles.meta}>{alert.ruleName}</span>
                          <span className={`${styles.meta} ${layout.mono} ${layout.truncate}`} title={alert.id}>
                            {alert.id}
                          </span>
                        </div>
                      </td>
                      <td>
                        <Label variant={statusVariants[alert.status]}>{statusLabels[alert.status]}</Label>
                        {alert.acknowledgedBy && <span className={styles.meta}>by {alert.acknowledgedBy}</span>}
                      </td>
                      <td>
                        <Text weight="medium">{deliveryLabels[alert.deliveryStatus]}</Text>
                        <span className={styles.meta}>
                          {alert.deliveryAttempts} {alert.deliveryAttempts === 1 ? 'attempt' : 'attempts'}
                        </span>
                        {alert.lastDeliveryError && (
                          <span className={`${styles.meta} ${styles.error}`} title={alert.lastDeliveryError}>
                            {alert.lastDeliveryError}
                          </span>
                        )}
                      </td>
                      <td>
                        <time dateTime={alert.triggeredAt} title={alert.triggeredAt} className={styles.nowrap}>
                          {formatTimestamp(alert.triggeredAt)}
                        </time>
                      </td>
                      <td className={styles.actionColumn}>
                        {alert.status === 'firing' ? (
                          <Button
                            variant={canAcknowledge ? 'primary' : 'default'}
                            disabled={isAcknowledging || !canAcknowledge}
                            leadingVisual={isAcknowledging ? undefined : CheckIcon}
                            onClick={() => void acknowledge(alert)}
                          >
                            {isAcknowledging ? 'Acknowledging...' : canAcknowledge ? 'Acknowledge' : 'View only'}
                          </Button>
                        ) : (
                          <span className={layout.muted}>{statusLabels[alert.status]}</span>
                        )}
                        {actionError && (
                          <p className={`${styles.meta} ${styles.error}`} role="alert">
                            {actionError}
                          </p>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </main>
  )
}

interface SummaryCardProps {
  label: string
  value: number
  icon: Icon
  tone: 'danger' | 'accent' | 'attention'
}

function SummaryCard({ label, value, icon: SummaryIcon, tone }: SummaryCardProps) {
  return (
    <div className={`${layout.box} ${layout.boxBody}`}>
      <div className={styles.summaryHead}>
        <div>
          <p className={layout.metricLabel}>{label}</p>
          <p className={layout.metricValue}>{value}</p>
        </div>
        <SummaryIcon size={24} className={styles[tone]} />
      </div>
    </div>
  )
}
