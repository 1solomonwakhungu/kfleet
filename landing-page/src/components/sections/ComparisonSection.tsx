import { Check, Minus } from 'lucide-react'
import { Section, SectionHeader } from '@/components/section'
import { Card, CardFooter } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { siteCopy } from '@/content/siteCopy'
import { cn } from '@/lib/utils'

const KFLEET_COLUMN_CLASS = 'bg-primary/[0.06] text-foreground font-medium'

// In the kfleet column "None" and "No" are the selling points (no stored
// kubeconfigs, no mutations), so they get a check. Elsewhere "None" only
// means the tool has nothing in that row, shown as a neutral dash.
function ComparisonValue({ value, isKfleet }: { readonly value: string; readonly isKfleet: boolean }) {
  const isEmptyValue = value === 'None' || value === 'No'

  if (isKfleet && isEmptyValue) {
    return (
      <span className="inline-flex items-center gap-1.5">
        <Check aria-hidden="true" className="size-4 shrink-0 text-success" />
        {value}
      </span>
    )
  }

  if (value === 'None') {
    return (
      <span className="inline-flex items-center gap-1.5 text-muted-foreground/80">
        <Minus aria-hidden="true" className="size-3.5 shrink-0" />
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

      <Card className="mt-12 overflow-hidden">
        <div
          role="region"
          aria-labelledby="compare-heading"
          tabIndex={0}
          className="scrollbar-subtle overflow-x-auto focus-visible:-outline-offset-2"
        >
          <Table className="min-w-[62rem]">
            <TableHeader>
              <TableRow>
                <TableHead scope="col" className="w-[15%] pl-6">
                  {capabilityColumn}
                </TableHead>
                {productColumns.map((column, index) => (
                  <TableHead
                    key={column}
                    scope="col"
                    className={cn(
                      index === 0 &&
                        'bg-primary/[0.06] text-sm font-semibold text-brand shadow-[inset_0_2px_0_0_var(--primary)]',
                    )}
                  >
                    {column}
                  </TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {comparison.rows.map((row) => (
                <TableRow key={row.capability}>
                  <TableHead
                    scope="row"
                    className="h-auto py-3.5 pl-6 text-sm font-medium whitespace-normal text-foreground"
                  >
                    {row.capability}
                  </TableHead>
                  {row.values.map((value, index) => (
                    <TableCell
                      key={productColumns[index]}
                      className={cn(
                        'py-3.5 text-[13px] leading-snug text-muted-foreground',
                        index === 0 && KFLEET_COLUMN_CLASS,
                      )}
                    >
                      <ComparisonValue value={value} isKfleet={index === 0} />
                    </TableCell>
                  ))}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
        <CardFooter>
          <p className="text-[13px] text-pretty text-muted-foreground">{comparison.footnote}</p>
        </CardFooter>
      </Card>
    </Section>
  )
}
