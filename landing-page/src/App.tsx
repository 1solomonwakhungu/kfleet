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

function App() {
  return (
    <div id="top" className="flex min-h-svh flex-col">
      <a
        href="#main-content"
        className="sr-only rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:z-50"
      >
        Skip to content
      </a>
      <Navbar />
      <main id="main-content" tabIndex={-1} className="flex-1 outline-none">
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
