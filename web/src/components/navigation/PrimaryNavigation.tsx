import { NavList } from '@primer/react'
import {
  BellIcon,
  LogIcon,
  PeopleIcon,
  PersonIcon,
  ShieldCheckIcon,
  StackIcon,
  type Icon,
} from '@primer/octicons-react'
import { Link, useLocation } from 'react-router-dom'

export interface NavigationItem {
  label: string
  description: string
  to: string
  icon: Icon
  end?: boolean
  activePathPrefixes?: readonly string[]
}

export const primaryNavigationItems: readonly NavigationItem[] = [
  {
    label: 'Policy',
    description: 'Drift and compliance',
    to: '/policies',
    icon: ShieldCheckIcon,
    end: true,
  },
  {
    label: 'Fleet',
    description: 'Cluster overview',
    to: '/',
    icon: StackIcon,
    end: true,
    activePathPrefixes: ['/clusters/'],
  },
  {
    label: 'Alerts',
    description: 'Fleet health history',
    to: '/alerts',
    icon: BellIcon,
    end: true,
  },
  {
    label: 'Agents',
    description: 'Pending approvals',
    to: '/agents',
    icon: PeopleIcon,
    end: true,
  },
]

export const readOnlyNavigationItems: readonly NavigationItem[] = primaryNavigationItems.filter(
  (item) => item.to !== '/agents',
)

/**
 * Admin-only destinations. The backing endpoints are registered behind
 * requireRole(types.RoleAdmin), so these are hidden for operators and
 * read-only users.
 */
export const adminNavigationItems: readonly NavigationItem[] = [
  {
    label: 'Users',
    description: 'Accounts and roles',
    to: '/admin/users',
    icon: PersonIcon,
    end: true,
  },
  {
    label: 'Audit log',
    description: 'Security history',
    to: '/admin/audit',
    icon: LogIcon,
    end: true,
  },
]

function isItemActive(item: NavigationItem, pathname: string): boolean {
  if (item.activePathPrefixes?.some((prefix) => pathname.startsWith(prefix))) return true
  if (item.end) return pathname === item.to
  return pathname === item.to || pathname.startsWith(`${item.to}/`)
}

interface PrimaryNavigationProps {
  items?: readonly NavigationItem[]
  onNavigate?: () => void
}

export function PrimaryNavigation({ items = primaryNavigationItems, onNavigate }: PrimaryNavigationProps) {
  const { pathname } = useLocation()

  return (
    <NavList aria-label="Primary navigation">
      {items.map((item) => {
        const ItemIcon = item.icon

        return (
          <NavList.Item
            key={item.to}
            as={Link}
            to={item.to}
            onClick={onNavigate}
            aria-current={isItemActive(item, pathname) ? 'page' : undefined}
          >
            <NavList.LeadingVisual>
              <ItemIcon />
            </NavList.LeadingVisual>
            {item.label}
            <NavList.Description>{item.description}</NavList.Description>
          </NavList.Item>
        )
      })}
    </NavList>
  )
}
