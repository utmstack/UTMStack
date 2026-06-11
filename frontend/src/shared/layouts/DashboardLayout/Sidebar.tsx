import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  Activity,
  AlertTriangle,
  ArrowLeft,
  BadgeCheck,
  Bell,
  Bot,
  Boxes,
  CalendarClock,
  ChevronDown,
  Database,
  Flame,
  HardDriveDownload,
  History,
  Home,
  Info,
  KeyRound,
  Languages,
  Mail,
  Palette,
  Plug,
  Radar,
  ScrollText,
  Settings,
  Shield,
  ShieldCheck,
  SlidersHorizontal,
  Filter,
  Regex,
  UserCheck,
  UserX,
  Users,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { useAuth } from '@/features/auth'

type LeafItem = {
  to: string
  label: string
  icon: LucideIcon
}

type GroupItem = {
  id: string
  label: string
  icon: LucideIcon
  basePath: string
  children: { to: string; label: string; icon: LucideIcon }[]
}

type Item = LeafItem | GroupItem

function isGroup(item: Item): item is GroupItem {
  return 'children' in item
}

type Section = {
  id: string
  label: string | null
  items: Item[]
  // Hidden from non-admins (ROLE_USER). System configuration is admin-only.
  adminOnly?: boolean
}

// `label` values are i18n keys, resolved with t() at render time.
const sections: Section[] = [
  {
    id: 'main',
    label: null,
    items: [
      { to: '/home', label: 'nav.home', icon: Home },
      {
        id: 'threat-management',
        label: 'nav.threatManagement',
        icon: Shield,
        basePath: '/threat-management',
        children: [
          { to: '/threat-management/alerts', label: 'nav.alerts', icon: AlertTriangle },
          { to: '/threat-management/incidents', label: 'nav.incidents', icon: Flame },
          { to: '/threat-management/adversaries', label: 'nav.adversaryView', icon: UserX },
        ],
      },
      { to: '/soar', label: 'nav.soar', icon: Workflow },
      { to: '/log-explorer', label: 'nav.logExplorer', icon: ScrollText },
      { to: '/user-auditor', label: 'nav.userAuditor', icon: UserCheck },
      { to: '/threat-intelligence', label: 'nav.threatIntelligence', icon: Radar },
      { to: '/compliance', label: 'nav.compliance', icon: BadgeCheck },
    ],
  },
  {
    id: 'configure',
    label: 'nav.configure',
    adminOnly: true,
    items: [
      { to: '/datasources', label: 'nav.dataSources', icon: Database },
      { to: '/integrations', label: 'nav.integrations', icon: Plug },
      { to: '/alerting-rules', label: 'nav.alertingRules', icon: SlidersHorizontal },
      { to: '/parsing-filters', label: 'nav.parsingFilters', icon: Filter },
      { to: '/regex-patterns', label: 'nav.regexPatterns', icon: Regex },
      { to: '/data-processing', label: 'nav.dataProcessing', icon: Activity },
      { to: '/team', label: 'nav.team', icon: Users },
      { to: '/settings', label: 'nav.settings', icon: Settings },
    ],
  },
]

// Settings drill-down — when pathname is under /settings the sidebar swaps
// its content for this list + a Back button. `label` values are i18n keys.
const settingsItems: { to: string; label: string; icon: LucideIcon }[] = [
  { to: '/settings/license', label: 'settings.license', icon: BadgeCheck },
  { to: '/settings/notifications', label: 'settings.notifications', icon: Bell },
  { to: '/settings/connection-key', label: 'settings.connectionKey', icon: KeyRound },
  { to: '/settings/data-retention', label: 'settings.dataRetention', icon: HardDriveDownload },
  { to: '/settings/indices', label: 'settings.indices', icon: Boxes },
  { to: '/settings/branding', label: 'settings.branding', icon: Palette },
  { to: '/settings/identity-providers', label: 'settings.identityProviders', icon: ShieldCheck },
  { to: '/settings/email', label: 'settings.email', icon: Mail },
  { to: '/settings/soc-ai', label: 'settings.socAi', icon: Bot },
  { to: '/settings/date-format', label: 'settings.dateFormat', icon: CalendarClock },
  { to: '/settings/language', label: 'settings.language', icon: Languages },
  { to: '/settings/audit-logs', label: 'settings.auditLogs', icon: History },
  { to: '/settings/about', label: 'settings.about', icon: Info },
]

const SECTIONS_KEY = 'utmstack-sidebar-sections-closed'
const GROUPS_KEY = 'utmstack-sidebar-groups-closed'

function loadSet(key: string): Set<string> {
  if (typeof window === 'undefined') return new Set()
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return new Set()
    const arr = JSON.parse(raw)
    return Array.isArray(arr) ? new Set(arr) : new Set()
  } catch {
    return new Set()
  }
}

export function Sidebar() {
  const { t } = useTranslation()
  const { isAdmin } = useAuth()
  const [closedSections, setClosedSections] = useState<Set<string>>(() => loadSet(SECTIONS_KEY))
  const [closedGroups, setClosedGroups] = useState<Set<string>>(() => loadSet(GROUPS_KEY))
  const { pathname } = useLocation()

  useEffect(() => {
    localStorage.setItem(SECTIONS_KEY, JSON.stringify([...closedSections]))
  }, [closedSections])

  useEffect(() => {
    localStorage.setItem(GROUPS_KEY, JSON.stringify([...closedGroups]))
  }, [closedGroups])

  const toggleSection = (id: string) => {
    setClosedSections((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const toggleGroup = (id: string) => {
    setClosedGroups((prev) => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const isPathActive = (to: string) => pathname === to || pathname.startsWith(to + '/')
  const isBaseActive = (basePath: string) =>
    pathname === basePath || pathname.startsWith(basePath + '/')

  // Drill-down: when navigating under /settings the sidebar swaps to a sub-nav.
  if (pathname.startsWith('/settings')) {
    return <SettingsSidebar isPathActive={isPathActive} />
  }

  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
      <nav className="flex-1 overflow-y-auto overflow-x-hidden p-2">
        {sections
          .filter((section) => !section.adminOnly || isAdmin)
          .map((section, idx) => {
          const isClosed = closedSections.has(section.id)
          return (
            <div
              key={section.id}
              className={cn(idx > 0 && 'mt-2 pt-2 border-t border-sidebar-border/60')}
            >
              {section.label && (
                <button
                  onClick={() => toggleSection(section.id)}
                  className="mb-1 flex w-full items-center justify-between px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/50 hover:text-sidebar-foreground"
                >
                  <span>{t(section.label)}</span>
                  <ChevronDown
                    size={12}
                    className={cn('transition-transform duration-150', isClosed && '-rotate-90')}
                  />
                </button>
              )}
              {!isClosed && (
                <div className="space-y-0.5">
                  {section.items.map((item) =>
                    isGroup(item) ? (
                      <SidebarGroup
                        key={item.id}
                        item={item}
                        open={!closedGroups.has(item.id)}
                        onToggle={() => toggleGroup(item.id)}
                        isPathActive={isPathActive}
                        isBaseActive={isBaseActive}
                      />
                    ) : (
                      <SidebarLeaf
                        key={item.to}
                        item={item}
                        active={isPathActive(item.to)}
                      />
                    )
                  )}
                </div>
              )}
            </div>
          )
        })}
      </nav>
    </aside>
  )
}

function SidebarLeaf({
  item,
  active,
  nested,
}: {
  item: { to: string; label: string; icon: LucideIcon }
  active: boolean
  nested?: boolean
}) {
  const { t } = useTranslation()
  const Icon = item.icon
  return (
    <Link
      to={item.to}
      className={cn(
        'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
        active
          ? 'bg-primary/15 text-primary'
          : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground',
        nested && 'pl-7'
      )}
    >
      <Icon size={16} strokeWidth={1.75} className="shrink-0" />
      <span className="whitespace-nowrap">{t(item.label)}</span>
    </Link>
  )
}

function SidebarGroup({
  item,
  open,
  onToggle,
  isPathActive,
  isBaseActive,
}: {
  item: GroupItem
  open: boolean
  onToggle: () => void
  isPathActive: (to: string) => boolean
  isBaseActive: (base: string) => boolean
}) {
  const { t } = useTranslation()
  const Icon = item.icon
  const baseActive = isBaseActive(item.basePath)

  return (
    <div>
      <button
        onClick={onToggle}
        className={cn(
          'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
          baseActive
            ? 'bg-primary/10 text-primary'
            : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground'
        )}
      >
        <Icon size={16} strokeWidth={1.75} className="shrink-0" />
        <span className="flex-1 whitespace-nowrap text-left">{t(item.label)}</span>
        <ChevronDown
          size={14}
          className={cn(
            'shrink-0 transition-transform duration-150',
            open ? 'rotate-0' : '-rotate-90'
          )}
        />
      </button>
      {open && (
        <div className="mt-0.5 space-y-0.5">
          {item.children.map((child) => (
            <SidebarLeaf
              key={child.to}
              item={{ to: child.to, label: child.label, icon: child.icon }}
              active={isPathActive(child.to)}
              nested
            />
          ))}
        </div>
      )}
    </div>
  )
}

function SettingsSidebar({ isPathActive }: { isPathActive: (to: string) => boolean }) {
  const { t } = useTranslation()
  return (
    <aside className="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
      <div className="border-b border-sidebar-border p-2">
        <Link
          to="/home"
          className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-sidebar-foreground/70 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
        >
          <ArrowLeft size={16} strokeWidth={1.75} className="shrink-0" />
          <span>{t('nav.back')}</span>
        </Link>
      </div>
      <div className="px-3 pt-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/50">
        {t('settings.title')}
      </div>
      <nav className="flex-1 space-y-0.5 overflow-y-auto overflow-x-hidden p-2 pt-1">
        {settingsItems.map((item) => {
          const Icon = item.icon
          return (
            <Link
              key={item.to}
              to={item.to}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors',
                isPathActive(item.to)
                  ? 'bg-primary/15 text-primary'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground'
              )}
            >
              <Icon size={16} strokeWidth={1.75} className="shrink-0" />
              <span className="whitespace-nowrap">{t(item.label)}</span>
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
