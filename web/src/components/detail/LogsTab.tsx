import { useEffect, useId, useMemo, useRef, useState } from 'react'
import { Button, FormControl, Select, Spinner, TextInput } from '@primer/react'
import {
  AlertIcon,
  BroadcastIcon,
  MoveToBottomIcon,
  SearchIcon,
  SyncIcon,
  TerminalIcon,
  TrashIcon,
  XIcon,
} from '@primer/octicons-react'

import { usePodLogs, type PodLogStatus } from '../../hooks/usePodLogs'
import type { PodInfo } from '../../types/resources'
import { ResourceState } from './ResourceTabState'
import resource from './resource.module.css'
import styles from './LogsTab.module.css'

const STREAM_LABELS: Record<PodLogStatus, string> = {
  idle: 'Idle',
  connecting: 'Connecting',
  streaming: 'Streaming',
  reconnecting: 'Reconnecting',
  ended: 'Stream ended',
  unavailable: 'Unavailable',
}

const STATUS_TONES: Record<PodLogStatus, string> = {
  idle: styles.toneMuted,
  connecting: styles.toneAccent,
  streaming: styles.toneSuccess,
  reconnecting: styles.toneAttention,
  ended: styles.toneMuted,
  unavailable: styles.toneDanger,
}

interface LogsTabProps {
  clusterId: string
  pods: PodInfo[]
  selectedPod?: PodInfo
  onSelectPod: (pod: PodInfo | undefined) => void
}

export function LogsTab({ clusterId, pods, selectedPod, onSelectPod }: LogsTabProps) {
  const [autoScroll, setAutoScroll] = useState(true)
  const [wrapLines, setWrapLines] = useState(true)
  const [filter, setFilter] = useState('')
  const viewerRef = useRef<HTMLDivElement>(null)
  const podSelectId = useId()

  const { lines, status: streamStatus, error, clear, retry } = usePodLogs({
    clusterId,
    namespace: selectedPod?.namespace ?? '',
    pod: selectedPod?.name ?? '',
  })

  useEffect(() => {
    if (!autoScroll) return
    const viewport = viewerRef.current
    if (viewport) viewport.scrollTop = viewport.scrollHeight
  }, [lines, autoScroll])

  const podKey = useMemo(
    () => (selectedPod ? `${selectedPod.namespace}/${selectedPod.name}` : undefined),
    [selectedPod],
  )
  const query = filter.trim().toLowerCase()
  const visibleLines = useMemo(
    () =>
      lines
        .map((line, index) => ({ line, number: index + 1 }))
        .filter(({ line }) => !query || line.toLowerCase().includes(query)),
    [lines, query],
  )

  if (pods.length === 0 && !selectedPod) {
    return (
      <ResourceState
        kind="empty"
        title="No pods available"
        description="Logs become available after the cluster reports at least one pod."
      />
    )
  }

  const unavailable = streamStatus === 'unavailable'

  return (
    <section className={resource.panel} aria-label="Pod log viewer">
      <div className={styles.controls}>
        <FormControl id={podSelectId}>
          <FormControl.Label>Pod</FormControl.Label>
          <Select
            aria-label="Pod for log stream"
            value={podKey ?? ''}
            onChange={(event) => {
              const pod = pods.find((candidate) => `${candidate.namespace}/${candidate.name}` === event.target.value)
              clear()
              setFilter('')
              setAutoScroll(true)
              onSelectPod(pod)
            }}
          >
            <Select.Option value="">Select a pod</Select.Option>
            {pods.map((pod) => (
              <Select.Option key={`${pod.namespace}/${pod.name}`} value={`${pod.namespace}/${pod.name}`}>
                {pod.namespace}/{pod.name}
              </Select.Option>
            ))}
          </Select>
        </FormControl>

        <FormControl disabled={!selectedPod || lines.length === 0}>
          <FormControl.Label>Find in logs</FormControl.Label>
          <TextInput
            type="search"
            block
            value={filter}
            placeholder="Filter log output…"
            leadingVisual={SearchIcon}
            aria-keyshortcuts="Escape"
            onChange={(event) => setFilter(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Escape' && filter) {
                event.preventDefault()
                setFilter('')
              }
            }}
            trailingAction={
              filter ? (
                <TextInput.Action icon={XIcon} aria-label="Clear log filter" onClick={() => setFilter('')} />
              ) : undefined
            }
          />
        </FormControl>

        <div className={styles.actions} aria-label="Log viewer controls">
          <Button aria-pressed={wrapLines} onClick={() => setWrapLines((current) => !current)}>
            Wrap
          </Button>
          <Button
            aria-pressed={autoScroll}
            leadingVisual={MoveToBottomIcon}
            onClick={() => setAutoScroll((current) => !current)}
          >
            Follow
          </Button>
          <Button leadingVisual={TrashIcon} disabled={lines.length === 0} onClick={clear}>
            Clear
          </Button>
        </div>
      </div>

      <div className={styles.statusBar}>
        <div className={`${styles.status} ${STATUS_TONES[streamStatus]}`} role="status" aria-live="polite">
          {streamStatus === 'connecting' ? (
            <Spinner size="small" />
          ) : unavailable ? (
            <AlertIcon size={12} />
          ) : streamStatus === 'streaming' ? (
            <BroadcastIcon size={12} />
          ) : (
            <TerminalIcon size={12} />
          )}
          <span>{STREAM_LABELS[streamStatus]}</span>
          {error && <span className={resource.muted}>— {error}</span>}
          {unavailable && (
            <Button size="small" leadingVisual={SyncIcon} onClick={retry}>
              Retry
            </Button>
          )}
        </div>
        <div className={styles.counts}>
          {query && <span>{visibleLines.length} matches</span>}
          <span>{lines.length} lines</span>
        </div>
      </div>

      <div className={styles.viewer} ref={viewerRef}>
        <div
          role="log"
          aria-label={selectedPod ? `Logs for ${selectedPod.namespace}/${selectedPod.name}` : 'Pod logs'}
          aria-live="off"
          className={`${styles.log} ${wrapLines ? '' : styles.logNoWrap}`}
        >
          {!selectedPod ? (
            <LogPlaceholder
              icon={<TerminalIcon size={24} />}
              title="Select a pod"
              description="Choose a pod above to open its live log stream."
            />
          ) : lines.length === 0 ? (
            <LogPlaceholder
              icon={unavailable || error ? <AlertIcon size={24} /> : <TerminalIcon size={24} />}
              title={
                unavailable
                  ? 'Log streaming is unavailable'
                  : error
                    ? 'Waiting for the log stream'
                    : streamStatus === 'ended'
                      ? 'No log output'
                      : 'No log output yet'
              }
              description={
                error ||
                (streamStatus === 'ended'
                  ? 'The pod produced no output for this stream.'
                  : 'The stream is connected; new output will appear here as the pod writes it.')
              }
              action={
                unavailable ? (
                  <Button leadingVisual={SyncIcon} onClick={retry}>
                    Try again
                  </Button>
                ) : undefined
              }
            />
          ) : visibleLines.length === 0 ? (
            <LogPlaceholder
              icon={<SearchIcon size={24} />}
              title="No matching log lines"
              description={`No output contains “${filter.trim()}”.`}
            />
          ) : (
            visibleLines.map(({ line, number }) => (
              <div key={number} className={styles.line}>
                <span className={styles.lineNumber} aria-hidden="true">
                  {number}
                </span>
                <code className={styles.lineContent}>{line || ' '}</code>
              </div>
            ))
          )}
        </div>
      </div>
    </section>
  )
}

interface LogPlaceholderProps {
  icon: React.ReactNode
  title: string
  description: string
  action?: React.ReactNode
}

function LogPlaceholder({ icon, title, description, action }: LogPlaceholderProps) {
  return (
    <div className={styles.placeholder}>
      <span className={styles.placeholderIcon}>{icon}</span>
      <p className={styles.placeholderTitle}>{title}</p>
      <p className={styles.placeholderText}>{description}</p>
      {action}
    </div>
  )
}
