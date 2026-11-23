import { BrandMark } from '@/components/brand-mark'
import { TonalChip } from '@/components/tonal/chip'
import { siteCopy, type FooterLink } from '@/content/siteCopy'
import { capture } from '@/lib/posthog'

const DOCS_PREFIX = `${siteCopy.links.repoUrl}/blob/main/docs/`

// Keeps the event names the previous footer used, so existing PostHog
// insights continue to receive data.
function footerEvent(link: FooterLink): string | undefined {
  const { links } = siteCopy

  if (!link.external) {
    return undefined
  }
  if (link.href === links.liveDemoUrl) {
    return 'footer_live_demo_clicked'
  }
  if (link.href === links.releaseUrl) {
    return 'release_notes_opened'
  }
  if (link.href === links.contributingUrl) {
    return 'contribution_guide_opened'
  }
  if (link.href === links.repoUrl) {
    return 'repository_opened'
  }
  if (link.href.startsWith(DOCS_PREFIX)) {
    return 'docs_opened'
  }
  return 'footer_link_clicked'
}

export function SiteFooter() {
  const { footer, meta } = siteCopy

  return (
    <footer className="border-t border-[var(--t-border-subtle)] bg-[var(--t-bg)]">
      <div className="page-container py-14 sm:py-16">
        <div className="grid grid-cols-2 gap-x-8 gap-y-12 lg:grid-cols-[minmax(0,1.5fr)_repeat(3,minmax(0,1fr))]">
          <div className="col-span-2 flex max-w-sm flex-col gap-4 lg:col-span-1">
            <a
              href="#top"
              aria-label={`${meta.name} home`}
              className="flex w-fit items-center gap-2.5 rounded-full text-[17px] font-medium tracking-[-0.02em] text-[var(--t-ink)]"
            >
              <BrandMark />
              <span>{meta.name}</span>
            </a>
            <p className="t-body text-[14px] text-pretty text-[var(--t-ink-muted)]">
              {footer.tagline}
            </p>
          </div>

          {footer.columns.map((column) => (
            <nav key={column.title} aria-label={column.title}>
              <h2 className="text-[13.5px] font-medium text-[var(--t-ink)]">{column.title}</h2>
              <ul className="mt-4 flex flex-col gap-3">
                {column.links.map((link) => {
                  const event = footerEvent(link)
                  return (
                    <li key={link.label}>
                      <a
                        href={link.href}
                        {...(link.external ? { target: '_blank', rel: 'noreferrer' } : {})}
                        onClick={event ? () => capture(event) : undefined}
                        className="rounded-sm text-[13.5px] text-[var(--t-ink-muted)] transition-colors duration-150 hover:text-[var(--t-ink)]"
                      >
                        {link.label}
                      </a>
                    </li>
                  )
                })}
              </ul>
            </nav>
          ))}
        </div>

        <div className="mt-14 flex flex-col gap-3 border-t border-[var(--t-border-subtle)] pt-6 text-[13px] text-[var(--t-ink-muted)] sm:flex-row sm:items-center sm:justify-between">
          <p>{footer.copyright}</p>
          <TonalChip tone="neutral" className="w-fit font-mono text-[11.5px]">
            {meta.version}
          </TonalChip>
        </div>
      </div>
    </footer>
  )
}
