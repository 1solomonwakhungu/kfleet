import { siteCopy } from '@/content/siteCopy'

export function MetricsStrip() {
  return (
    <section aria-label="Engineering facts" className="page-container">
      <dl className="grid grid-cols-1 gap-3.5 sm:grid-cols-2 lg:grid-cols-4">
        {siteCopy.metrics.map((metric) => (
          <div
            key={metric.label}
            className="flex flex-col justify-between gap-3 rounded-[20px] bg-[var(--t-surface)] p-5 transition-colors duration-200 [transition-timing-function:var(--t-ease)] sm:p-6"
          >
            <dt className="text-[13px] font-medium text-[var(--t-ink-muted)]">{metric.label}</dt>
            <div>
              <dd className="text-[32px] leading-none font-medium tracking-[-0.035em] text-[var(--t-ink)] tabular-nums sm:text-[36px]">
                {metric.value}
              </dd>
              <dd className="mt-2.5 text-[13px] leading-[1.45] text-pretty text-[var(--t-ink-secondary)]">
                {metric.detail}
              </dd>
            </div>
          </div>
        ))}
      </dl>
    </section>
  )
}
