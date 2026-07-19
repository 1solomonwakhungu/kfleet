import { Search, X } from 'lucide-react'

import { Button } from '../ui/button'
import { Card } from '../ui/card'
import { Input } from '../ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../ui/select'
import { cn } from '../../lib/utils'
import type { ClusterHealth } from '../../types/cluster'

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

const triggerClasses = 'h-11 border-border bg-background text-foreground shadow-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-background'
const itemClasses = 'data-[state=checked]:bg-blue-500/15 data-[state=checked]:text-blue-200 focus:bg-blue-600 focus:text-white'

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
    <Card className="mt-5 border border-border p-4" aria-labelledby="fleet-controls-heading">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-end">
        <div className="min-w-0 flex-1">
          <div className="mb-3 flex items-center justify-between gap-3">
            <h2 id="fleet-controls-heading" className="text-sm font-bold">Fleet controls</h2>
            <p className="whitespace-nowrap font-mono text-xs text-muted xl:hidden" aria-hidden="true">
              {resultCount} / {totalCount}
            </p>
          </div>
          <label htmlFor="cluster-search" className="mb-1.5 block text-xs font-semibold text-muted">
            Search clusters or labels
          </label>
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" aria-hidden="true" />
            <Input
              id="cluster-search"
              type="search"
              value={search}
              onChange={(event) => onSearchChange(event.target.value)}
              placeholder="production or region=us-east"
              className="pl-10 pr-10 focus-visible:border-blue-500 focus-visible:outline-blue-500"
            />
            {search && (
              <Button
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 h-9 min-h-9 w-9 -translate-y-1/2 px-0 text-muted hover:text-foreground focus-visible:outline-blue-500"
                aria-label="Clear cluster search"
                onClick={() => onSearchChange('')}
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </Button>
            )}
          </div>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 xl:w-[31rem]">
          <div>
            <label id="health-filter-label" className="mb-1.5 block text-xs font-semibold text-muted">Health</label>
            <Select value={health} onValueChange={(value) => onHealthChange(value as HealthFilter)}>
              <SelectTrigger
                aria-labelledby="health-filter-label"
                className={cn(triggerClasses, health !== 'all' && 'border-blue-500/70 bg-blue-500/10 text-blue-100')}
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="border-border bg-elevated text-foreground">
                <SelectItem value="all" className={itemClasses}>All health</SelectItem>
                <SelectItem value="healthy" className={itemClasses}>Healthy</SelectItem>
                <SelectItem value="degraded" className={itemClasses}>Degraded</SelectItem>
                <SelectItem value="unreachable" className={itemClasses}>Unreachable</SelectItem>
                <SelectItem value="unknown" className={itemClasses}>Unknown</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div>
            <label id="fleet-sort-label" className="mb-1.5 block text-xs font-semibold text-muted">Sort by</label>
            <Select value={sort} onValueChange={(value) => onSortChange(value as FleetSort)}>
              <SelectTrigger aria-labelledby="fleet-sort-label" className={triggerClasses}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="border-border bg-elevated text-foreground">
                <SelectItem value="health" className={itemClasses}>Health · needs attention</SelectItem>
                <SelectItem value="name" className={itemClasses}>Name · A–Z</SelectItem>
                <SelectItem value="heartbeat" className={itemClasses}>Heartbeat · newest</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className={cn('min-h-11 items-center justify-between gap-3 xl:justify-end', hasActiveControls ? 'flex' : 'hidden xl:flex')}>
          <p className="hidden whitespace-nowrap font-mono text-xs text-muted xl:block" aria-hidden="true">
            {resultCount} / {totalCount}
          </p>
          {hasActiveControls && (
            <Button variant="ghost" className="px-3 text-blue-400 hover:bg-blue-500/10 hover:text-blue-300" onClick={onReset}>
              <X className="h-4 w-4" aria-hidden="true" />
              Reset
            </Button>
          )}
        </div>
      </div>
    </Card>
  )
}
