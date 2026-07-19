import { Card } from '../ui/card'

export function DashboardSkeleton() {
  return (
    <div className="animate-pulse" aria-label="Loading fleet dashboard" role="status">
      <span className="sr-only">Loading fleet dashboard</span>
      <div className="mt-6 grid grid-cols-2 gap-3 lg:grid-cols-4">
        {Array.from({ length: 4 }, (_, index) => (
          <Card key={index} className="h-28 border border-border p-4 sm:h-32 sm:p-5">
            <div className="h-3 w-16 rounded bg-elevated" />
            <div className="mt-4 h-8 w-20 rounded bg-elevated" />
            <div className="mt-3 h-3 w-24 max-w-full rounded bg-elevated" />
          </Card>
        ))}
      </div>

      <Card className="mt-5 border border-border p-4">
        <div className="h-4 w-24 rounded bg-elevated" />
        <div className="mt-4 grid gap-3 xl:grid-cols-[minmax(0,1fr)_15rem_15rem]">
          <div className="h-11 rounded-md bg-elevated" />
          <div className="h-11 rounded-md bg-elevated" />
          <div className="h-11 rounded-md bg-elevated" />
        </div>
      </Card>

      <div className="mt-5 h-5 w-36 rounded bg-elevated" />
      <div className="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
        {Array.from({ length: 8 }, (_, index) => (
          <Card key={index} className="h-[19rem] border border-border p-5">
            <div className="flex justify-between gap-4">
              <div className="h-5 w-2/3 rounded bg-elevated" />
              <div className="h-5 w-20 rounded bg-elevated" />
            </div>
            <div className="mt-6 grid grid-cols-2 gap-px overflow-hidden rounded-md bg-border p-px">
              {Array.from({ length: 4 }, (_, metricIndex) => (
                <div key={metricIndex} className="h-14 bg-elevated" />
              ))}
            </div>
            <div className="mt-5 h-3 w-14 rounded bg-elevated" />
            <div className="mt-3 h-7 w-4/5 rounded bg-elevated" />
            <div className="mt-5 h-px bg-border" />
            <div className="mt-4 h-4 w-1/2 rounded bg-elevated" />
          </Card>
        ))}
      </div>
    </div>
  )
}
