import { SkeletonBox } from '@primer/react'
import { SkeletonText } from '@primer/react/experimental'

import layout from '../../styles/layout.module.css'
import styles from './DashboardSkeleton.module.css'

export function DashboardSkeleton() {
  return (
    <div role="status" aria-label="Loading fleet dashboard">
      <span className={layout.srOnly}>Loading fleet dashboard</span>

      <div className={`${layout.grid} ${layout.grid4}`}>
        {Array.from({ length: 4 }, (_, index) => (
          <div key={index} className={`${layout.box} ${layout.boxBody} ${styles.tile}`}>
            <SkeletonText size="bodySmall" maxWidth="40%" />
            <SkeletonText size="titleMedium" maxWidth="55%" />
            <SkeletonText size="bodySmall" maxWidth="70%" />
          </div>
        ))}
      </div>

      <div className={`${layout.box} ${layout.boxBody} ${styles.controls}`}>
        <SkeletonBox height="32px" />
        <SkeletonBox height="32px" />
        <SkeletonBox height="32px" />
      </div>

      <div className={`${layout.grid} ${layout.grid3} ${styles.cards}`}>
        {Array.from({ length: 6 }, (_, index) => (
          <div key={index} className={`${layout.box} ${layout.boxBody} ${styles.tile}`}>
            <SkeletonText size="titleSmall" maxWidth="60%" />
            <SkeletonBox height="76px" />
            <SkeletonText size="bodySmall" maxWidth="80%" />
            <SkeletonText size="bodySmall" maxWidth="45%" />
          </div>
        ))}
      </div>
    </div>
  )
}
