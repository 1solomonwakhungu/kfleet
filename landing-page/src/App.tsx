import * as React from 'react'
import { ArchitectureSection } from '@/components/sections/ArchitectureSection'
import { ComparisonSection } from '@/components/sections/ComparisonSection'
import { CtaSection } from '@/components/sections/CtaSection'
import { FeaturesBento } from '@/components/sections/FeaturesBento'
import { HeroSection } from '@/components/sections/HeroSection'
import { McpSection } from '@/components/sections/McpSection'
import { MetricsStrip } from '@/components/sections/MetricsStrip'
import { Navbar } from '@/components/sections/Navbar'
import { QuickstartSection } from '@/components/sections/QuickstartSection'
import { SiteFooter } from '@/components/sections/SiteFooter'
import { TradeoffsAndFaqSection } from '@/components/sections/TradeoffsAndFaqSection'

type TonalTheme = 'light' | 'dark'

function readInitialTheme(): TonalTheme {
  if (typeof document === 'undefined') {
    return 'light'
  }
  return document.documentElement.dataset.tonalTheme === 'dark' ? 'dark' : 'light'
}

function App() {
  const [theme, setTheme] = React.useState<TonalTheme>(readInitialTheme)

  React.useEffect(() => {
    document.documentElement.dataset.tonalTheme = theme
  }, [theme])

  const toggleTheme = React.useCallback(() => {
    setTheme((prev) => {
      const next: TonalTheme = prev === 'dark' ? 'light' : 'dark'
      try {
        window.localStorage.setItem('kfleet-tonal-theme', next)
      } catch {
        // Ignore storage access errors in restricted browsers.
      }
      return next
    })
  }, [])

  return (
    <div
      id="top"
      data-tonal-theme={theme}
      className="flex min-h-svh flex-col bg-[var(--t-bg)] text-[var(--t-ink)] transition-colors duration-200"
    >
      <a
        href="#main-content"
        className="sr-only rounded-full bg-[var(--t-cta-bg)] px-4 py-2 text-sm font-medium text-[var(--t-cta-fg)] focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:z-50"
      >
        Skip to content
      </a>
      <Navbar theme={theme} onToggleTheme={toggleTheme} />
      <main
        id="main-content"
        tabIndex={-1}
        className="flex flex-1 flex-col gap-20 pb-20 outline-none sm:gap-28 sm:pb-24"
      >
        <HeroSection />
        <MetricsStrip />
        <FeaturesBento />
        <McpSection />
        <ArchitectureSection />
        <ComparisonSection />
        <QuickstartSection />
        <TradeoffsAndFaqSection />
        <CtaSection />
      </main>
      <SiteFooter />
    </div>
  )
}

export default App
