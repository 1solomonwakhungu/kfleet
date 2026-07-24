import { useEffect, useMemo, useRef, useState } from 'react'
import { Activity, Boxes, Eye, LogOut, Menu, ShieldCheck, X } from 'lucide-react'
import { Link, Outlet, useLocation } from 'react-router-dom'

import { useAuth } from '../../auth/AuthContext'
import { Button } from '../ui/button'
import { cn } from '../../lib/utils'
import { api, type RuntimeInfo } from '../../lib/api'
import {
  PrimaryNavigation,
  primaryNavigationItems,
  readOnlyNavigationItems,
} from '../navigation/PrimaryNavigation'

export function ApplicationShell() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
  const [runtime, setRuntime] = useState<RuntimeInfo | null>(null)
  const menuButtonRef = useRef<HTMLButtonElement>(null)
  const navigationItems = useMemo(
    () => (runtime?.readOnly ? readOnlyNavigationItems : primaryNavigationItems),
    [runtime?.readOnly],
  )

  useEffect(() => {
    const controller = new AbortController()
    void api.getRuntimeInfo(controller.signal).then(setRuntime).catch(() => {
      // Older hubs do not expose runtime metadata. Preserve the normal UI.
    })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    setMobileNavigationOpen(false)
  }, [location.pathname])

  useEffect(() => {
    if (!mobileNavigationOpen) return

    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return

      setMobileNavigationOpen(false)
      menuButtonRef.current?.focus()
    }

    document.addEventListener('keydown', closeOnEscape)
    return () => document.removeEventListener('keydown', closeOnEscape)
  }, [mobileNavigationOpen])

  return (
    <div className="min-h-dvh bg-background text-foreground lg:grid lg:grid-cols-[16rem_minmax(0,1fr)]">
      <a
        href="#main-content"
        className="fixed left-4 top-4 z-50 -translate-y-24 rounded-md bg-accent px-4 py-3 text-sm font-semibold text-accent-foreground transition-transform duration-150 ease-out focus:translate-y-0"
      >
        Skip to main content
      </a>

      <aside className="sticky top-0 hidden h-dvh flex-col border-r border-border bg-surface lg:flex">
        <div className="border-b border-border px-5 py-5">
          <BrandLink />
        </div>

        <div className="flex min-h-0 flex-1 flex-col px-3 py-5">
          <p className="px-3 font-mono text-xs font-semibold uppercase tracking-[0.14em] text-muted">Workspace</p>
          <PrimaryNavigation className="mt-2" items={navigationItems} />
          <div className="mt-auto space-y-3">
            {!runtime?.readOnly && (
              <AccountSummary
                username={user?.username ?? ''}
                role={user?.role ?? 'read_only'}
                onLogout={() => void logout()}
              />
            )}
            <EnvironmentStatus runtime={runtime} />
          </div>
        </div>
      </aside>

      <div className="min-w-0">
        <header className="sticky top-0 z-40 border-b border-border bg-surface lg:hidden">
          <div className="flex min-h-16 items-center justify-between gap-4 px-4 sm:px-6">
            <BrandLink />
            <Button
              ref={menuButtonRef}
              variant="outline"
              className="min-w-11 px-0"
              aria-controls="mobile-navigation"
              aria-expanded={mobileNavigationOpen}
              aria-label={mobileNavigationOpen ? 'Close navigation menu' : 'Open navigation menu'}
              onClick={() => setMobileNavigationOpen((open) => !open)}
            >
              {mobileNavigationOpen ? (
                <X className="h-5 w-5" aria-hidden="true" />
              ) : (
                <Menu className="h-5 w-5" aria-hidden="true" />
              )}
            </Button>
          </div>

          {mobileNavigationOpen && (
            <div id="mobile-navigation" className="border-t border-border px-4 pb-4 pt-3 sm:px-6">
              <PrimaryNavigation items={navigationItems} onNavigate={() => setMobileNavigationOpen(false)} />
              {!runtime?.readOnly && (
                <AccountSummary
                  className="mt-3"
                  username={user?.username ?? ''}
                  role={user?.role ?? 'read_only'}
                  onLogout={() => void logout()}
                />
              )}
              <EnvironmentStatus className="mt-3" compact runtime={runtime} />
            </div>
          )}
        </header>

        {runtime?.demoMode && <DemoNotice policy={runtime.dataPolicy} />}

        <div id="main-content" className="min-w-0 flex-1" tabIndex={-1}>
          <Outlet />
        </div>
      </div>
    </div>
  )
}

interface AccountSummaryProps {
  className?: string
  username: string
  role: string
  onLogout: () => void
}

function AccountSummary({ className, username, role, onLogout }: AccountSummaryProps) {
  return (
    <section className={cn('rounded-md border border-border bg-background p-3', className)} aria-label="Signed in user">
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold">{username}</p>
          <p className="mt-1 font-mono text-[0.6875rem] uppercase tracking-[0.1em] text-muted">
            {role.replace('_', ' ')}
          </p>
        </div>
        <Button variant="ghost" size="sm" className="px-2" aria-label="Sign out" onClick={onLogout}>
          <LogOut className="size-4" aria-hidden="true" />
        </Button>
      </div>
    </section>
  )
}

function BrandLink() {
  return (
    <Link to="/" className="inline-flex min-h-11 items-center gap-3 whitespace-nowrap rounded-md">
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-blue-600 text-white shadow-sm shadow-blue-950/30">
        <Boxes className="h-5 w-5" aria-hidden="true" />
      </span>
      <span className="min-w-0 leading-none">
        <span className="block font-display text-sm font-bold tracking-[0.08em]">kFLEET</span>
        <span className="mt-1 block text-xs text-muted">Control plane</span>
      </span>
    </Link>
  )
}

interface EnvironmentStatusProps {
  className?: string
  compact?: boolean
  runtime: RuntimeInfo | null
}

function EnvironmentStatus({ className, compact = false, runtime }: EnvironmentStatusProps) {
  return (
    <section
      className={cn(
        'rounded-md border border-border bg-background p-3',
        compact && 'grid grid-cols-2 gap-3',
        className,
      )}
      aria-label="Environment and control plane status"
    >
      <div className="min-w-0">
        <p className="font-mono text-[0.6875rem] uppercase tracking-[0.12em] text-muted">Environment</p>
        <p className="mt-1 truncate whitespace-nowrap text-sm font-semibold">
          {runtime?.demoMode ? 'Public demo' : 'Fleet'}
        </p>
      </div>
      <div className={cn('mt-3 border-t border-border pt-3', compact && 'mt-0 border-l border-t-0 pl-3 pt-0')}>
        <p className="font-mono text-[0.6875rem] uppercase tracking-[0.12em] text-muted">Hub status</p>
        <p className="mt-1 flex items-center gap-2 whitespace-nowrap text-sm text-muted">
          {runtime?.readOnly ? (
            <ShieldCheck className="h-3.5 w-3.5 shrink-0 text-healthy" aria-hidden="true" />
          ) : (
            <Activity className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          )}
          {runtime?.readOnly ? 'Read-only' : 'Not reported'}
        </p>
      </div>
    </section>
  )
}

function DemoNotice({ policy }: { policy: string }) {
  return (
    <aside
      className="border-b border-blue-800/70 bg-blue-950/80 px-4 py-3 text-blue-100 sm:px-6 lg:px-8"
      aria-label="Public demo safety notice"
    >
      <div className="mx-auto flex max-w-[100rem] items-start gap-3">
        <Eye className="mt-0.5 h-4 w-4 shrink-0 text-blue-300" aria-hidden="true" />
        <p className="text-sm leading-5">
          <strong className="font-semibold">Read-only synthetic demo.</strong>{' '}
          <span className="text-blue-200">{policy} Mutating API requests are disabled.</span>
        </p>
      </div>
    </aside>
  )
}
