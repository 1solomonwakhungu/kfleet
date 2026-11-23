import { useRef, useState, type MouseEvent } from 'react'
import { ArrowUpRight, Menu } from 'lucide-react'
import { BrandMark } from '@/components/brand-mark'
import { BrandGithubIcon } from '@/components/icons/brand-github'
import { LinkButton } from '@/components/link-button'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { buttonVariants } from '@/components/ui/button-variants'
import { siteCopy } from '@/content/siteCopy'
import { cn } from '@/lib/utils'

export function Navbar() {
  const { nav, meta } = siteCopy
  const [menuOpen, setMenuOpen] = useState(false)
  const pendingAnchor = useRef<string | null>(null)

  // The sheet locks page scroll while it is open, so an in-page link cannot
  // scroll yet. Remember the target, close, and jump once the close finishes.
  const followAnchorAfterClose = (event: MouseEvent<HTMLAnchorElement>, href: string) => {
    event.preventDefault()
    pendingAnchor.current = href
    setMenuOpen(false)
  }

  const handleOpenChangeComplete = (open: boolean) => {
    const href = pendingAnchor.current
    if (open || !href) {
      return
    }
    pendingAnchor.current = null
    document.getElementById(href.slice(1))?.scrollIntoView()
    window.history.pushState(null, '', href)
  }

  return (
    <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-md">
      <div className="page-container flex h-14 items-center gap-6">
        <a
          href="#top"
          aria-label={`${meta.name} home`}
          className="flex items-center gap-2.5 rounded-sm text-[17px] font-semibold tracking-tight"
        >
          <BrandMark />
          {meta.name}
        </a>

        <nav aria-label="Primary" className="hidden lg:block">
          <ul className="flex items-center gap-0.5">
            {nav.items.map((item) => (
              <li key={item.href}>
                <a
                  href={item.href}
                  className="rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
                >
                  {item.label}
                </a>
              </li>
            ))}
          </ul>
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <LinkButton
            link={nav.secondaryCta}
            event="header_github_clicked"
            variant="outline"
            size="sm"
            className="hidden sm:inline-flex"
          >
            <BrandGithubIcon aria-hidden="true" />
            {nav.secondaryCta.label}
          </LinkButton>
          <LinkButton link={nav.primaryCta} event="header_live_demo_clicked" size="sm">
            {nav.primaryCta.label}
            <ArrowUpRight aria-hidden="true" />
          </LinkButton>

          <Sheet
            open={menuOpen}
            onOpenChange={setMenuOpen}
            onOpenChangeComplete={handleOpenChangeComplete}
          >
            <SheetTrigger
              aria-label="Open menu"
              className={cn(buttonVariants({ variant: 'ghost', size: 'icon-sm' }), 'lg:hidden')}
            >
              <Menu aria-hidden="true" />
            </SheetTrigger>
            <SheetContent closeLabel="Close menu">
              <SheetHeader>
                <SheetTitle className="flex items-center gap-2.5">
                  <BrandMark />
                  {meta.name}
                </SheetTitle>
                <SheetDescription className="sr-only">Page sections and project links</SheetDescription>
              </SheetHeader>
              <Separator />
              <nav aria-label="Mobile" className="px-3 py-3">
                <ul className="flex flex-col">
                  {nav.items.map((item) => (
                    <li key={item.href}>
                      <a
                        href={item.href}
                        onClick={(event) => followAnchorAfterClose(event, item.href)}
                        className="flex h-11 items-center rounded-md px-2 text-[15px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                      >
                        {item.label}
                      </a>
                    </li>
                  ))}
                </ul>
              </nav>
              <Separator />
              <div className="flex flex-col gap-2 p-5">
                <LinkButton link={nav.primaryCta} event="header_live_demo_clicked" size="lg">
                  {nav.primaryCta.label}
                  <ArrowUpRight aria-hidden="true" />
                </LinkButton>
                <LinkButton
                  link={nav.secondaryCta}
                  event="header_github_clicked"
                  variant="outline"
                  size="lg"
                >
                  <BrandGithubIcon aria-hidden="true" />
                  {nav.secondaryCta.label}
                </LinkButton>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  )
}
