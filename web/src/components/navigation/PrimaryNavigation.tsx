import { LayoutDashboard, type LucideIcon } from 'lucide-react'
import { NavLink, useLocation } from 'react-router-dom'

import { cn } from '../../lib/utils'

export interface NavigationItem {
  label: string
  description: string
  to: string
  icon: LucideIcon
  end?: boolean
  activePathPrefixes?: readonly string[]
}

export const primaryNavigationItems: readonly NavigationItem[] = [
  {
    label: 'Fleet',
    description: 'Cluster overview',
    to: '/',
    icon: LayoutDashboard,
    end: true,
    activePathPrefixes: ['/clusters/'],
  },
]

interface PrimaryNavigationProps {
  className?: string
  items?: readonly NavigationItem[]
  onNavigate?: () => void
}

export function PrimaryNavigation({
  className,
  items = primaryNavigationItems,
  onNavigate,
}: PrimaryNavigationProps) {
  const location = useLocation()

  return (
    <nav className={className} aria-label="Primary navigation">
      <ul className="space-y-1">
        {items.map((item) => {
          const Icon = item.icon
          const isRelatedRoute = item.activePathPrefixes?.some((prefix) => location.pathname.startsWith(prefix)) ?? false

          return (
            <li key={item.to}>
              <NavLink
                to={item.to}
                end={item.end}
                onClick={onNavigate}
                className={({ isActive }) =>
                  cn(
                    'group flex min-h-11 items-center gap-3 rounded-md px-3 py-2 text-sm font-semibold transition-[background-color,color,transform] duration-150 ease-out active:translate-y-px',
                    isActive || isRelatedRoute
                      ? 'bg-elevated text-foreground'
                      : 'text-muted hover:bg-elevated hover:text-foreground',
                  )
                }
              >
                {({ isActive }) => (
                  <>
                    <Icon
                      className={cn(
                        'h-5 w-5 shrink-0 transition-colors duration-150 ease-out',
                        isActive || isRelatedRoute ? 'text-accent' : 'text-muted group-hover:text-foreground',
                      )}
                      aria-hidden="true"
                    />
                    <span className="min-w-0">
                      <span className="block whitespace-nowrap">{item.label}</span>
                      <span className="block truncate whitespace-nowrap text-xs font-normal text-muted">
                        {item.description}
                      </span>
                    </span>
                  </>
                )}
              </NavLink>
            </li>
          )
        })}
      </ul>
    </nav>
  )
}
