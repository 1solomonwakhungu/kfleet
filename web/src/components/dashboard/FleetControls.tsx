import { Button, FormControl, Select, TextInput } from '@primer/react'
import { SearchIcon, XIcon } from '@primer/octicons-react'

import type { ClusterHealth } from '../../types/cluster'
import layout from '../../styles/layout.module.css'
import styles from './FleetControls.module.css'

export type HealthFilter = 'all' | ClusterHealth
export type FleetSort = 'health' | 'name' | 'heartbeat'

interface FleetControlsProps {
  search: string
  onSearchChange: (value: string) => void
  health: HealthFilter
  onHealthChange: (value: HealthFilter) => void
  sort: FleetSort
  onSortChange: (value: FleetSort) => void
  resultCount: number
  totalCount: number
  hasActiveControls: boolean
  onReset: () => void
}

export function FleetControls({
  search,
  onSearchChange,
  health,
  onHealthChange,
  sort,
  onSortChange,
  resultCount,
  totalCount,
  hasActiveControls,
  onReset,
}: FleetControlsProps) {
  return (
    <section className={`${layout.box} ${layout.boxBody} ${styles.controls}`} aria-labelledby="fleet-controls-heading">
      <h2 id="fleet-controls-heading" className={layout.srOnly}>
        Fleet controls
      </h2>

      <div className={styles.searchField}>
        <FormControl>
          <FormControl.Label>Search clusters or labels</FormControl.Label>
          <TextInput
            type="search"
            block
            value={search}
            leadingVisual={SearchIcon}
            placeholder="production or region=us-east"
            onChange={(event) => onSearchChange(event.target.value)}
            trailingAction={
              search ? (
                <TextInput.Action icon={XIcon} aria-label="Clear cluster search" onClick={() => onSearchChange('')} />
              ) : undefined
            }
          />
        </FormControl>
      </div>

      <FormControl>
        <FormControl.Label>Health</FormControl.Label>
        <Select value={health} onChange={(event) => onHealthChange(event.target.value as HealthFilter)}>
          <Select.Option value="all">All health</Select.Option>
          <Select.Option value="healthy">Healthy</Select.Option>
          <Select.Option value="degraded">Degraded</Select.Option>
          <Select.Option value="unreachable">Unreachable</Select.Option>
          <Select.Option value="unknown">Unknown</Select.Option>
        </Select>
      </FormControl>

      <FormControl>
        <FormControl.Label>Sort by</FormControl.Label>
        <Select value={sort} onChange={(event) => onSortChange(event.target.value as FleetSort)}>
          <Select.Option value="health">Health · needs attention</Select.Option>
          <Select.Option value="name">Name · A–Z</Select.Option>
          <Select.Option value="heartbeat">Heartbeat · newest</Select.Option>
        </Select>
      </FormControl>

      <div className={styles.summary}>
        <span className={`${layout.mono} ${layout.muted}`} aria-hidden="true">
          {resultCount} / {totalCount}
        </span>
        {hasActiveControls && (
          <Button variant="invisible" leadingVisual={XIcon} onClick={onReset}>
            Reset
          </Button>
        )}
      </div>
    </section>
  )
}
