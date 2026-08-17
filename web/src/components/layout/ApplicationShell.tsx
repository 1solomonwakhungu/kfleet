import { useEffect, useMemo, useRef, useState } from 'react'
import { IconButton, Label, Text, Truncate } from '@primer/react'
import {
  EyeIcon,
  MoonIcon,
  PulseIcon,
  ShieldCheckIcon,
  SignOutIcon,
  StackIcon,
  SunIcon,
  ThreeBarsIcon,
  XIcon,
} from '@primer/octicons-react'
import { Link, Outlet, useLocation } from 'react-router-dom'

import { useAuth } from '../../auth/AuthContext'
import { api, type RuntimeInfo } from '../../lib/api'
import { useColorMode } from '../../theme/ColorModeProvider'
import {
  PrimaryNavigation,
  adminNavigationItems,
  primaryNavigationItems,
  readOnlyNavigationItems,
} from '../navigation/PrimaryNavigation'
import styles from './ApplicationShell.module.css'

export function ApplicationShell() {
  const { user, logout } = useAuth()
  const location = useLocation()
  const [mobileNavigationOpen, setMobileNavigationOpen] = useState(false)
  const [runtime, setRuntime] = useState<RuntimeInfo | null>(null)
  const menuButtonRef = useRef<HTMLButtonElement>(null)
  const navigationItems = useMemo(() => {
    if (runtime?.readOnly) return readOnlyNavigationItems
    if (user?.role === 'admin') return [...primaryNavigationItems, ...adminNavigationItems]
    return primaryNavigationItems
  }, [runtime?.readOnly, user?.role])

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
    <div className={styles.shell}>
      <a href="#main-content" className={styles.skipLink}>
        Skip to main content
      </a>

      <aside className={styles.sidebar}>
        <BrandLink />

        <div className={styles.sidebarScroll}>
          <PrimaryNavigation items={navigationItems} />
        </div>

        <div className={styles.sidebarFooter}>
          {!runtime?.readOnly && (
            <AccountSummary username={user?.username ?? ''} role={user?.role ?? 'read_only'} onLogout={() => void logout()} />
          )}
          <EnvironmentStatus runtime={runtime} />
        </div>
      </aside>

      <div className={styles.main}>
        <header className={styles.topbar}>
          <IconButton
            ref={menuButtonRef}
            className={styles.mobileOnly}
            icon={mobileNavigationOpen ? XIcon : ThreeBarsIcon}
            variant="invisible"
            aria-controls="mobile-navigation"
            aria-expanded={mobileNavigationOpen}
            aria-label={mobileNavigationOpen ? 'Close navigation menu' : 'Open navigation menu'}
            onClick={() => setMobileNavigationOpen((open) => !open)}
          />
          <Text className={styles.mobileOnly} weight="semibold">
            kfleet
          </Text>
          <div className={styles.topbarActions}>
            <ColorModeToggle />
          </div>
        </header>

        {mobileNavigationOpen && (
          <div id="mobile-navigation" className={styles.mobileNav}>
            <PrimaryNavigation items={navigationItems} onNavigate={() => setMobileNavigationOpen(false)} />
            {!runtime?.readOnly && (
              <AccountSummary
                username={user?.username ?? ''}
                role={user?.role ?? 'read_only'}
                onLogout={() => void logout()}
              />
            )}
          </div>
        )}

        {runtime?.demoMode && <DemoNotice policy={runtime.dataPolicy} />}

        <div id="main-content" tabIndex={-1}>
          <Outlet />
        </div>
      </div>
    </div>
  )
}

function ColorModeToggle() {
  const { resolvedMode, toggleColorMode } = useColorMode()

  return (
    <IconButton
      icon={resolvedMode === 'dark' ? SunIcon : MoonIcon}
      variant="invisible"
      aria-label={resolvedMode === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
      onClick={toggleColorMode}
    />
  )
}

interface AccountSummaryProps {
  username: string
  role: string
  onLogout: () => void
}

function AccountSummary({ username, role, onLogout }: AccountSummaryProps) {
  const initials = username.slice(0, 2).toUpperCase() || '··'

  return (
    <section className={styles.footerCard} aria-label="Signed in user">
      <span className={styles.avatar} aria-hidden="true">
        {initials}
      </span>
      <div className={styles.identity}>
        <Truncate className={styles.identityName} title={username}>
          {username}
        </Truncate>
        <span className={styles.footerLabel}>{role.replace('_', ' ')}</span>
      </div>
      <IconButton icon={SignOutIcon} variant="invisible" size="small" aria-label="Sign out" onClick={onLogout} />
    </section>
  )
}

function BrandLink() {
  return (
    <Link to="/" className={styles.brand}>
      <span className={styles.brandMark}>
        <StackIcon size={16} />
      </span>
      <span>
        <Text weight="semibold">kfleet</Text>
        <span className={styles.footerLabel}>Control plane</span>
      </span>
    </Link>
  )
}

function EnvironmentStatus({ runtime }: { runtime: RuntimeInfo | null }) {
  return (
    <section className={styles.footerCard} aria-label="Environment and control plane status">
      <div className={styles.identity}>
        <span className={styles.footerLabel}>{runtime?.demoMode ? 'Public demo' : 'Fleet'}</span>
      </div>
      <Label variant={runtime?.readOnly ? 'success' : 'secondary'}>
        {runtime?.readOnly ? <ShieldCheckIcon size={12} /> : <PulseIcon size={12} />}
        {runtime?.readOnly ? ' Read-only' : ' Hub'}
      </Label>
    </section>
  )
}

function DemoNotice({ policy }: { policy: string }) {
  return (
    <aside className={styles.demoNotice} aria-label="Public demo safety notice">
      <div className={styles.demoNoticeInner}>
        <EyeIcon size={16} />
        <Text>
          <Text weight="semibold">Read-only synthetic demo.</Text>{' '}
          <Text>{policy} Mutating API requests are disabled.</Text>
        </Text>
      </div>
    </aside>
  )
}
