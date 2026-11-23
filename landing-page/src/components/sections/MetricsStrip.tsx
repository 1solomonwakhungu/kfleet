import { siteCopy } from '@/content/siteCopy'

export function MetricsStrip() {
  return (
    <section aria-label="Engineering facts" className="border-t border-border/60">
      <div className="page-container">
        {/* The 1px gaps over a tinted grid draw the hairlines between cells. */}
        <dl className="grid grid-cols-2 gap-px border-x border-border/60 bg-border/60 lg:grid-cols-4">
          {siteCopy.metrics.map((metric) => (
            <div key={metric.label} className="flex flex-col gap-1.5 bg-background px-4 py-7 sm:px-6 sm:py-9">
              <dt className="order-2 text-sm font-medium text-foreground">{metric.label}</dt>
              <dd className="order-1 font-mono text-3xl font-medium tracking-tight text-foreground sm:text-4xl">
                {metric.value}
              </dd>
              <dd className="order-3 text-[13px] leading-snug text-pretty text-muted-foreground">
                {metric.detail}
              </dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  )
}
