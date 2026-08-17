import type { ClusterHealth } from '../types/cluster'
import styles from './StatusDot.module.css'

interface StatusDotProps {
  health: ClusterHealth
}

/** Compact health indicator used beside cluster and fleet labels. */
export function StatusDot({ health }: StatusDotProps) {
  return <span className={`${styles.dot} ${styles[health]}`} aria-hidden="true" />
}
