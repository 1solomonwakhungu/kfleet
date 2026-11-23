import { ArrowUpRight, Moon, Sun } from 'lucide-react'
import { BrandMark } from '@/components/brand-mark'
import { BrandGithubIcon } from '@/components/icons/brand-github'
import { TonalLinkButton } from '@/components/tonal/button'
import { siteCopy } from '@/content/siteCopy'

interface NavbarProps {
  readonly theme: 'light' | 'dark'
  readonly onToggleTheme: () => void
}

export function Navbar({ theme, onToggleTheme }: NavbarProps) {
  const { nav, meta } = siteCopy

  return (
    <header className="sticky top-0 z-40 border-b border-[var(--t-border-subtle)] bg-[var(--t-bg)]/90 backdrop-blur-md transition-colors duration-200">
      <div className="page-container flex h-16 items-center justify-between gap-4">
        <div className="flex items-center gap-6">
          <a
            href="#top"
            aria-label={`${meta.name} home`}
            className="flex items-center gap-2.5 rounded-full text-[17px] font-medium tracking-[-0.02em] text-[var(--t-ink)]"
          >
            <BrandMark />
            <span>{meta.name}</span>
          </a>

          <nav aria-label="Primary" className="hidden lg:block">
            <ul className="flex items-center gap-1">
              {nav.items.map((item) => (
                <li key={item.href}>
                  <a
                    href={item.href}
                    className="rounded-full px-3.5 py-1.5 text-[13.5px] font-medium text-[var(--t-ink-muted)] transition-colors duration-200 hover:bg-[var(--t-surface)] hover:text-[var(--t-ink)]"
                  >
                    {item.label}
                  </a>
                </li>
              ))}
            </ul>
          </nav>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={onToggleTheme}
            aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            className="inline-flex h-9 w-9 shrink-0 cursor-pointer items-center justify-center rounded-full border border-[var(--t-border-subtle)] bg-[var(--t-surface)] text-[var(--t-ink-secondary)] transition-colors duration-200 hover:bg-[var(--t-surface-muted)] hover:text-[var(--t-ink)]"
          >
            {theme === 'dark' ? (
              <Sun aria-hidden="true" className="h-4 w-4" />
            ) : (
              <Moon aria-hidden="true" className="h-4 w-4" />
            )}
          </button>

          <TonalLinkButton
            link={nav.secondaryCta}
            event="header_github_clicked"
            variant="outline"
            size="sm"
            className="hidden sm:inline-flex"
          >
            <BrandGithubIcon aria-hidden="true" />
            {nav.secondaryCta.label}
          </TonalLinkButton>

          <TonalLinkButton link={nav.primaryCta} event="header_live_demo_clicked" size="sm">
            {nav.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </TonalLinkButton>
        </div>
      </div>

      {/* Mobile horizontal pill section bar */}
      <nav
        aria-label="Mobile sections"
        className="scrollbar-none flex items-center gap-1.5 overflow-x-auto px-5 pb-2.5 lg:hidden"
      >
        {nav.items.map((item) => (
          <a
            key={item.href}
            href={item.href}
            className="shrink-0 rounded-full bg-[var(--t-surface)] px-3.5 py-1 text-[12.5px] font-medium text-[var(--t-ink-muted)] transition-colors hover:bg-[var(--t-blue-tint)] hover:text-[var(--t-blue-hover)]"
          >
            {item.label}
          </a>
        ))}
      </nav>
    </header>
  )
}
