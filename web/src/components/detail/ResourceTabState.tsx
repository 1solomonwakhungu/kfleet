import type { ReactNode } from 'react'
import { Blankslate, SkeletonText } from '@primer/react/experimental'
import { AlertIcon, DatabaseIcon, SearchIcon } from '@primer/octicons-react'

import styles from './resource.module.css'

interface ResourceStateProps {
  kind: 'empty' | 'error'
  title: string
  description: string
}

/** Shared empty and error presentation for every cluster resource tab. */
export function ResourceState({ kind, title, description }: ResourceStateProps) {
  const Icon = kind === 'error' ? AlertIcon : SearchIcon

  return (
    <div className={styles.panel} role={kind === 'error' ? 'alert' : 'status'}>
      <Blankslate>
        <Blankslate.Visual>
          <Icon size={24} />
        </Blankslate.Visual>
        <Blankslate.Heading as="h3">{title}</Blankslate.Heading>
        <Blankslate.Description>{description}</Blankslate.Description>
      </Blankslate>
    </div>
  )
}

interface ResourceTableSkeletonProps {
  label: string
  columns: number
  rows?: number
}

export function ResourceTableSkeleton({ label, columns, rows = 6 }: ResourceTableSkeletonProps) {
  return (
    <section className={styles.panel} aria-busy="true" aria-label={label}>
      <div className={styles.panelHeader}>
        <DatabaseIcon size={16} className={styles.panelIcon} />
        <SkeletonText size="bodySmall" maxWidth="8rem" />
      </div>
      <div className={styles.scroll}>
        <table className={styles.table}>
          <caption className={styles.srOnly}>{label}</caption>
          <thead>
            <tr>
              {Array.from({ length: columns }, (_, index) => (
                <th key={index} scope="col">
                  <SkeletonText size="bodySmall" maxWidth="5rem" />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: rows }, (_, row) => (
              <tr key={row}>
                {Array.from({ length: columns }, (_, column) => (
                  <td key={column}>
                    <SkeletonText size="bodySmall" maxWidth={column === 0 ? '10rem' : '4rem'} />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}

interface ResourceTablePanelProps {
  label: string
  count: number
  noun: string
  children: ReactNode
}

export function ResourceTablePanel({ label, count, noun, children }: ResourceTablePanelProps) {
  return (
    <section className={styles.panel} aria-label={label}>
      <div className={styles.panelHeader}>
        <DatabaseIcon size={16} className={styles.panelIcon} />
        <span className={styles.panelCount}>{count}</span>
        <span className={styles.muted}>{count === 1 ? noun : `${noun}s`}</span>
      </div>
      <div className={styles.scroll}>{children}</div>
    </section>
  )
}
