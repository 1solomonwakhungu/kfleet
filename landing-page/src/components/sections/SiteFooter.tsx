import { BrandMark } from '@/components/brand-mark'
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
    <footer className="border-t border-border/60">
      <div className="page-container py-14 sm:py-16">
        <div className="grid grid-cols-2 gap-x-8 gap-y-12 lg:grid-cols-[minmax(0,1.5fr)_repeat(3,minmax(0,1fr))]">
          <div className="col-span-2 flex max-w-sm flex-col gap-4 lg:col-span-1">
            <a
              href="#top"
              aria-label={`${meta.name} home`}
              className="flex w-fit items-center gap-2.5 rounded-sm text-[17px] font-semibold tracking-tight"
            >
              <BrandMark />
              {meta.name}
            </a>
            <p className="text-sm leading-relaxed text-pretty text-muted-foreground">{footer.tagline}</p>
          </div>

          {footer.columns.map((column) => (
            <nav key={column.title} aria-label={column.title}>
              <h2 className="text-sm font-medium text-foreground">{column.title}</h2>
              <ul className="mt-4 flex flex-col gap-3">
                {column.links.map((link) => {
                  const event = footerEvent(link)
                  return (
                    <li key={link.label}>
                      <a
                        href={link.href}
                        {...(link.external ? { target: '_blank', rel: 'noreferrer' } : {})}
                        onClick={event ? () => capture(event) : undefined}
                        className="rounded-sm text-sm text-muted-foreground transition-colors hover:text-foreground"
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

        <div className="mt-14 flex flex-col gap-2 border-t border-border/60 pt-6 text-[13px] text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <p>{footer.copyright}</p>
          <p className="font-mono text-xs">{meta.version}</p>
        </div>
      </div>
    </footer>
  )
}
