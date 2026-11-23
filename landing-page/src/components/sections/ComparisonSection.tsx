import { Check, Minus } from 'lucide-react'
import { Section, SectionHeader } from '@/components/section'
import { siteCopy } from '@/content/siteCopy'
import { cn } from '@/lib/utils'

function ComparisonValue({ value, isKfleet }: { readonly value: string; readonly isKfleet: boolean }) {
  const isEmptyValue = value === 'None' || value === 'No'

  if (isKfleet && isEmptyValue) {
    return (
      <span className="inline-flex items-center gap-1.5 font-medium text-[var(--t-ink)]">
        <Check aria-hidden="true" className="h-4 w-4 shrink-0 text-[var(--t-chip-mint-fg)]" />
        {value}
      </span>
    )
  }

  if (value === 'None') {
    return (
      <span className="inline-flex items-center gap-1.5 text-[var(--t-ink-muted)]">
        <Minus aria-hidden="true" className="h-3.5 w-3.5 shrink-0" />
        {value}
      </span>
    )
  }

  return value
}

export function ComparisonSection() {
  const { comparison } = siteCopy
  const [capabilityColumn, ...productColumns] = comparison.columns

  return (
    <Section id={comparison.id} aria-labelledby="compare-heading">
      <SectionHeader
        headingId="compare-heading"
        eyebrow={comparison.eyebrow}
        title={comparison.title}
        subtitle={comparison.subtitle}
      />

      <div className="mt-10 overflow-hidden rounded-[var(--t-radius-card)] border border-[var(--t-border-subtle)] bg-[var(--t-bg)]">
        <div
          role="region"
          aria-labelledby="compare-heading"
          tabIndex={0}
          className="scrollbar-subtle overflow-x-auto focus-visible:-outline-offset-2"
        >
          <table className="w-full min-w-[60rem] border-collapse text-left">
            <thead>
              <tr className="border-b border-[var(--t-border-subtle)] bg-[var(--t-surface)] text-[13px] font-medium text-[var(--t-ink-muted)]">
                <th scope="col" className="w-[16%] py-4 pr-4 pl-6 font-medium">
                  {capabilityColumn}
                </th>
                {productColumns.map((column, index) => (
                  <th
                    key={column}
                    scope="col"
                    className={cn(
                      'px-4 py-4 font-medium',
                      index === 0
                        ? 'bg-[var(--t-surface-sky)]/70 font-medium text-[var(--t-blue-hover)] shadow-[inset_0_2px_0_0_var(--t-blue)]'
                        : 'text-[var(--t-ink)]',
                    )}
                  >
                    {column}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-[var(--t-border-subtle)] text-[13.5px]">
              {comparison.rows.map((row) => (
                <tr
                  key={row.capability}
                  className="transition-colors duration-150 hover:bg-[var(--t-surface)]/55"
                >
                  <th
                    scope="row"
                    className="py-4 pr-4 pl-6 align-top text-[14px] font-medium whitespace-normal text-[var(--t-ink-muted)]"
                  >
                    {row.capability}
                  </th>
                  {row.values.map((value, index) => (
                    <td
                      key={productColumns[index]}
                      className={cn(
                        'px-4 py-4 align-top leading-snug',
                        index === 0
                          ? 'bg-[var(--t-surface-sky)]/45 font-medium text-[var(--t-ink)]'
                          : 'text-[var(--t-ink-secondary)]',
                      )}
                    >
                      <ComparisonValue value={value} isKfleet={index === 0} />
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="border-t border-[var(--t-border-subtle)] bg-[var(--t-surface)]/65 px-6 py-4">
          <p className="text-[13px] text-pretty text-[var(--t-ink-muted)]">{comparison.footnote}</p>
        </div>
      </div>
    </Section>
  )
}
