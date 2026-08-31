import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import {
  Activity,
  AlertTriangle,
  ArrowLeft,
  BadgeCheck,
  Bell,
  Bot,
  Building2,
  CalendarClock,
  ChevronDown,
  Database,
  Flame,
  HardDriveDownload,
  History,
  Home,
  Info,
  KeyRound,
  KeySquare,
  Languages,
  LayoutDashboard,
  LifeBuoy,
  Mail,
  Palette,
  PanelLeftClose,
  PanelLeftOpen,
  Plug,
  Radar,
  ScrollText,
  Settings,
  Shield,
  ShieldCheck,
  SlidersHorizontal,
  Filter,
  Tags,
  Terminal,
  UserCheck,
  UserX,
  Users,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip'
import { useAuth } from '@/features/auth'
import { useBilling } from '@/features/billing'
import { useSupportTenant } from '@/shared/lib/current-tenant'
import { IS_FEDERATION } from '@/shared/config/mode'

type LeafItem = {
  to: string
  label: string
  icon: LucideIcon
  /** Light up for any path under `to` (e.g. Dashboards spans /dashboards/*). */
  matchPrefix?: boolean
  /** Only for whoever runs the instance, and only where there are tenants to
   * run: a customer's own administrator holds ROLE_ADMIN too, and without MSSP
   * the single tenant's administrator *is* the platform identity. */
  platformOnly?: boolean
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
      { to: '/home', label: 'nav.overview', icon: Home },
      { to: '/tenants', label: 'nav.tenants', icon: Building2, platformOnly: true },
      { to: '/dashboards', label: 'nav.dashboards', icon: LayoutDashboard, matchPrefix: true },
      {
        id: 'threat-management',
        label: 'nav.threatManagement',
        icon: Shield,
        basePath: '/threat-management',
        children: [
          { to: '/threat-management/alerts', label: 'nav.alerts', icon: AlertTriangle },
          { to: '/threat-management/alerts/tagging-rules', label: 'nav.taggingRules', icon: Tags },
          { to: '/threat-management/incidents', label: 'nav.incidents', icon: Flame },
          { to: '/threat-management/adversaries', label: 'nav.adversaryView', icon: UserX },
        ],
      },
      {
        id: 'soar',
        label: 'nav.soar',
        icon: Workflow,
        basePath: '/soar',
        children: [
          { to: '/soar/flows', label: 'nav.flows', icon: Workflow },
          { to: '/soar/execution-history', label: 'nav.executionHistory', icon: History },
          { to: '/soar/interactive-console', label: 'nav.interactiveConsole', icon: Terminal },
        ],
      },
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
      { to: '/pipelines', label: 'nav.parsingFilters', icon: Filter },
      { to: '/data-processing', label: 'nav.dataProcessing', icon: Activity },
      { to: '/team', label: 'nav.team', icon: Users },
      { to: '/settings', label: 'nav.settings', icon: Settings },
    ],
  },
]

// Settings drill-down — when pathname is under /settings the sidebar swaps
// its content for this list + a Back button. `label` values are i18n keys.
const settingsItems: {
  to: string
  label: string
  icon: LucideIcon
  /** For customer tenants only: the platform's own tenant has nobody to grant
   * support access to. */
  tenantOnly?: boolean
  /** Properties of the instance rather than of a customer on it — the licence,
   * the retention policy, the build. Only the default tenant sees these. */
  platformOnly?: boolean
}[] = [
  { to: '/settings/license', label: 'settings.license', icon: BadgeCheck, platformOnly: true },
  { to: '/settings/notifications', label: 'settings.notifications', icon: Bell },
  { to: '/settings/connection-key', label: 'settings.connectionKey', icon: KeyRound },
  {
    to: '/settings/data-retention',
    label: 'settings.dataRetention',
    icon: HardDriveDownload,
    platformOnly: true,
  },
  { to: '/settings/branding', label: 'settings.branding', icon: Palette },
  ...(!IS_FEDERATION
    ? [{ to: '/settings/api-keys', label: 'settings.apiKeys', icon: KeySquare }]
    : []),
  { to: '/settings/identity-providers', label: 'settings.identityProviders', icon: ShieldCheck },
  {
    to: '/settings/support-access',
    label: 'settings.supportAccess',
    icon: LifeBuoy,
    tenantOnly: true,
  },
  { to: '/settings/email', label: 'settings.email', icon: Mail },
  { to: '/settings/soc-ai', label: 'settings.socAi', icon: Bot },
  { to: '/settings/date-format', label: 'settings.dateFormat', icon: CalendarClock },
  { to: '/settings/language', label: 'settings.language', icon: Languages },
  { to: '/settings/audit-logs', label: 'settings.auditLogs', icon: History },
  { to: '/settings/about', label: 'settings.about', icon: Info, platformOnly: true },
]

const SECTIONS_KEY = 'utmstack-sidebar-sections-closed'
const GROUPS_KEY = 'utmstack-sidebar-groups-closed'
const COLLAPSED_KEY = 'utmstack-sidebar-collapsed'

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

function loadCollapsed(): boolean {
  if (typeof window === 'undefined') return false
  return localStorage.getItem(COLLAPSED_KEY) === '1'
}

function useCollapsed(): [boolean, () => void] {
  const [collapsed, setCollapsed] = useState<boolean>(loadCollapsed)
  useEffect(() => {
    localStorage.setItem(COLLAPSED_KEY, collapsed ? '1' : '0')
  }, [collapsed])
  return [collapsed, () => setCollapsed((v) => !v)]
}

function CollapseToggle({ collapsed, onToggle }: { collapsed: boolean; onToggle: () => void }) {
  const Icon = collapsed ? PanelLeftOpen : PanelLeftClose
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
      className={cn(
        'flex items-center rounded-lg p-2 text-sidebar-foreground/60 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground',
        collapsed ? 'justify-center' : 'ml-auto'
      )}
    >
      <Icon size={16} strokeWidth={1.75} />
    </button>
  )
}

export function Sidebar() {
  const { t } = useTranslation()
  const { isAdmin, isPlatformAdmin } = useAuth()
  const { license } = useBilling()
  const isMSSP = license?.mssp === true
  const supportTenant = useSupportTenant()
  // Inside a support session the operator *is* that tenant as far as the API is
  // concerned, so the platform-plane entries would only lead to a 403.
  const actingAsPlatform = isPlatformAdmin && supportTenant === null
  const [closedSections, setClosedSections] = useState<Set<string>>(() => loadSet(SECTIONS_KEY))
  const [closedGroups, setClosedGroups] = useState<Set<string>>(() => loadSet(GROUPS_KEY))
  const [collapsed, toggleCollapsed] = useCollapsed()
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

  const isPathActive = (to: string) => pathname === to
  const isBaseActive = (basePath: string) =>
    pathname === basePath || pathname.startsWith(basePath + '/')

  // Drill-down: when navigating under /settings the sidebar swaps to a sub-nav.
  if (pathname.startsWith('/settings')) {
    return <SettingsSidebar isPathActive={isPathActive} collapsed={collapsed} onToggleCollapsed={toggleCollapsed} />
  }

  return (
    <aside
      className={cn(
        'flex h-full shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-150',
        collapsed ? 'w-14' : 'w-60'
      )}
    >
      <div className="flex p-2">
        <CollapseToggle collapsed={collapsed} onToggle={toggleCollapsed} />
      </div>
      <nav className="flex-1 overflow-y-auto overflow-x-hidden p-2 pt-0">
        {sections
          .filter((section) => !section.adminOnly || isAdmin)
          .map((section, idx) => {
          const isClosed = closedSections.has(section.id)
          return (
            <div
              key={section.id}
              className={cn(idx > 0 && 'mt-2 pt-2 border-t border-sidebar-border/60')}
            >
              {section.label && !collapsed && (
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
              {(collapsed || !isClosed) && (
                <div className="space-y-0.5">
                  {section.items
                    .filter(
                      (item) =>
                        isGroup(item) || !item.platformOnly || (actingAsPlatform && isMSSP)
                    )
                    .map((item) =>
                    isGroup(item) ? (
                      <SidebarGroup
                        key={item.id}
                        item={item}
                        open={!closedGroups.has(item.id)}
                        onToggle={() => toggleGroup(item.id)}
                        isPathActive={isPathActive}
                        isBaseActive={isBaseActive}
                        collapsed={collapsed}
                        onExpandRequest={toggleCollapsed}
                      />
                    ) : (
                      <SidebarLeaf
                        key={item.to}
                        item={item}
                        active={item.matchPrefix ? isBaseActive(item.to) : isPathActive(item.to)}
                        collapsed={collapsed}
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
  collapsed,
}: {
  item: { to: string; label: string; icon: LucideIcon }
  active: boolean
  nested?: boolean
  collapsed?: boolean
}) {
  const { t } = useTranslation()
  const Icon = item.icon
  const label = t(item.label)
  const link = (
    <Link
      to={item.to}
      className={cn(
        'flex items-center rounded-lg text-sm transition-colors',
        collapsed ? 'justify-center px-2 py-2' : 'gap-3 px-3 py-2',
        active
          ? 'bg-primary/15 text-primary'
          : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground',
        !collapsed && nested && 'pl-7'
      )}
      aria-label={collapsed ? label : undefined}
    >
      <Icon size={16} strokeWidth={1.75} className="shrink-0" />
      {!collapsed && <span className="whitespace-nowrap">{label}</span>}
    </Link>
  )
  if (!collapsed) return link
  return (
    <Tooltip>
      <TooltipTrigger asChild>{link}</TooltipTrigger>
      <TooltipContent side="right">{label}</TooltipContent>
    </Tooltip>
  )
}

function SidebarGroup({
  item,
  open,
  onToggle,
  isPathActive,
  isBaseActive,
  collapsed,
  onExpandRequest,
}: {
  item: GroupItem
  open: boolean
  onToggle: () => void
  isPathActive: (to: string) => boolean
  isBaseActive: (base: string) => boolean
  collapsed?: boolean
  onExpandRequest?: () => void
}) {
  const { t } = useTranslation()
  const Icon = item.icon
  const baseActive = isBaseActive(item.basePath)
  const label = t(item.label)

  // Collapsed: no sub-menu room, so clicking a group just re-opens the sidebar.
  if (collapsed) {
    const btn = (
      <button
        onClick={onExpandRequest}
        aria-label={label}
        className={cn(
          'flex w-full items-center justify-center rounded-lg px-2 py-2 text-sm transition-colors',
          baseActive
            ? 'bg-primary/10 text-primary'
            : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground'
        )}
      >
        <Icon size={16} strokeWidth={1.75} className="shrink-0" />
      </button>
    )
    return (
      <Tooltip>
        <TooltipTrigger asChild>{btn}</TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    )
  }

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
        <span className="flex-1 whitespace-nowrap text-left">{label}</span>
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

function SettingsSidebar({
  isPathActive,
  collapsed,
  onToggleCollapsed,
}: {
  isPathActive: (to: string) => boolean
  collapsed: boolean
  onToggleCollapsed: () => void
}) {
  const { t } = useTranslation()
  const { isAdmin, isPlatformAdmin } = useAuth()
  const { license } = useBilling()
  const supportTenant = useSupportTenant()
  const items = settingsItems.filter((item) => {
    if (item.platformOnly) return isPlatformAdmin && supportTenant === null
    if (item.tenantOnly) return isAdmin && !isPlatformAdmin && license?.mssp === true
    return true
  })
  const backLabel = t('nav.back')
  return (
    <aside
      className={cn(
        'flex h-full shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-150',
        collapsed ? 'w-14' : 'w-60'
      )}
    >
      <div className="flex border-b border-sidebar-border p-2">
        {collapsed ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to="/home"
                aria-label={backLabel}
                className="flex flex-1 items-center justify-center rounded-lg px-2 py-2 text-sm text-sidebar-foreground/70 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
              >
                <ArrowLeft size={16} strokeWidth={1.75} className="shrink-0" />
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right">{backLabel}</TooltipContent>
          </Tooltip>
        ) : (
          <Link
            to="/home"
            className="flex flex-1 items-center gap-2 rounded-lg px-3 py-2 text-sm text-sidebar-foreground/70 transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground"
          >
            <ArrowLeft size={16} strokeWidth={1.75} className="shrink-0" />
            <span>{backLabel}</span>
          </Link>
        )}
      </div>
      <div className="flex p-2 pb-0">
        <CollapseToggle collapsed={collapsed} onToggle={onToggleCollapsed} />
      </div>
      {!collapsed && (
        <div className="px-3 pt-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-sidebar-foreground/50">
          {t('settings.title')}
        </div>
      )}
      <nav className="flex-1 space-y-0.5 overflow-y-auto overflow-x-hidden p-2 pt-1">
        {items.map((item) => (
          <SidebarLeaf
            key={item.to}
            item={{ to: item.to, label: item.label, icon: item.icon }}
            active={isPathActive(item.to)}
            collapsed={collapsed}
          />
        ))}
      </nav>
    </aside>
  )
}
