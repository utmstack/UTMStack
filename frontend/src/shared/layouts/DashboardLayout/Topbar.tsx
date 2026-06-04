import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Bell,
  Check,
  ChevronDown,
  ChevronsUpDown,
  LogOut,
  Moon,
  Search,
  Settings,
  Star,
  Sun,
  UserCircle,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers'
import { useAuth } from '@/features/auth'
import { useWorkspaces, type Workspace } from '@/features/workspace'
import { APP_INFO, type LicenseStatus } from '@/shared/config/app'

/* ─── Workspace visual helpers ─────────────────────────────────────────── */

// Stable gradient palette — picked deterministically from the workspace slug
// so each workspace keeps the same tile color across reloads.
const WORKSPACE_TONES = [
  'from-fuchsia-500 to-violet-600',
  'from-sky-500 to-cyan-500',
  'from-amber-500 to-orange-500',
  'from-emerald-500 to-teal-600',
  'from-rose-500 to-pink-600',
  'from-indigo-500 to-blue-600',
  'from-lime-500 to-green-600',
  'from-purple-500 to-fuchsia-600',
]

function workspaceTone(slug: string): string {
  let hash = 0
  for (let i = 0; i < slug.length; i++) {
    hash = (hash * 31 + slug.charCodeAt(i)) | 0
  }
  return WORKSPACE_TONES[Math.abs(hash) % WORKSPACE_TONES.length]
}

function workspaceInitials(name: string) {
  const parts = name.split(' ').filter(Boolean)
  if (parts.length === 0) return '··'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[1][0]).toUpperCase()
}

const STATUS_COLOR: Record<LicenseStatus, string> = {
  active: 'bg-emerald-500',
  trial: 'bg-sky-500',
  expiring: 'bg-amber-500',
  expired: 'bg-destructive',
  none: 'bg-muted-foreground/40',
}

const STATUS_LABEL: Record<LicenseStatus, string> = {
  active: 'Active',
  trial: 'Trial',
  expiring: 'Expiring soon',
  expired: 'Expired',
  none: 'No license',
}

function editionLabel(): string {
  if (APP_INFO.edition === 'enterprise') {
    return APP_INFO.tier ? `Enterprise · ${APP_INFO.tier}` : 'Enterprise'
  }
  return 'Community'
}

export function Topbar() {
  const [wsOpen, setWsOpen] = useState(false)
  const [notifOpen, setNotifOpen] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const { theme, toggleTheme } = useThemeContext()
  const { user, logout } = useAuth()
  const { workspaces, current: currentWs, setCurrent } = useWorkspaces()
  const navigate = useNavigate()
  const wsRef = useRef<HTMLDivElement>(null)
  const notifRef = useRef<HTMLDivElement>(null)
  const profileRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const t = e.target as Node
      if (wsRef.current && !wsRef.current.contains(t)) setWsOpen(false)
      if (notifRef.current && !notifRef.current.contains(t)) setNotifOpen(false)
      if (profileRef.current && !profileRef.current.contains(t)) setProfileOpen(false)
    }
    window.addEventListener('mousedown', handler)
    return () => window.removeEventListener('mousedown', handler)
  }, [])

  const handleLogout = async () => {
    await logout()
    navigate('/auth/login', { replace: true })
  }

  const userName =
    [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.login || 'User'
  const userInitials = userName
    .split(' ')
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase()
  const dot = STATUS_COLOR[APP_INFO.licenseStatus]

  return (
    <header className="relative z-40 flex h-14 shrink-0 items-center justify-between gap-3 border-b border-border bg-card px-3">
      {/* Left cluster */}
      <div className="flex min-w-0 items-center gap-2">
        <Link to="/home" className="flex items-center gap-2 px-1">
          <img src="/logo.svg" alt="UTMStack" className="h-7 w-7" />
          <span className="text-sm font-semibold">UTMStack</span>
        </Link>

        {/* Workspace switcher — always rendered when at least one workspace
            is loaded. Even with one workspace, showing the picker exposes
            the path to create more (an Enterprise feature). */}
        {currentWs && (
          <>
            <span className="text-border" aria-hidden>/</span>
            <div className="relative" ref={wsRef}>
              <button
                onClick={() => {
                  setWsOpen((v) => !v)
                  setNotifOpen(false)
                  setProfileOpen(false)
                }}
                className={cn(
                  'flex h-9 items-center gap-2 rounded-md pl-1 pr-2 text-sm transition-colors',
                  wsOpen
                    ? 'bg-muted text-foreground'
                    : 'text-foreground hover:bg-muted'
                )}
              >
                <span
                  className={cn(
                    'flex h-6 w-6 items-center justify-center rounded-md bg-gradient-to-br text-[10px] font-semibold tracking-tight text-white',
                    workspaceTone(currentWs.slug)
                  )}
                >
                  {workspaceInitials(currentWs.name)}
                </span>
                <span className="max-w-[180px] truncate font-medium">{currentWs.name}</span>
                <ChevronsUpDown size={12} className="text-muted-foreground" />
              </button>
              {wsOpen && (
                <WorkspacePicker
                  current={currentWs}
                  workspaces={workspaces}
                  onPick={(id) => {
                    setCurrent(id)
                    setWsOpen(false)
                  }}
                  onClose={() => setWsOpen(false)}
                  onManage={() => {
                    setWsOpen(false)
                    navigate('/settings/workspaces')
                  }}
                />
              )}
            </div>
          </>
        )}

      </div>

      {/* Right cluster */}
      <div className="flex items-center gap-1">
        <div className="relative" ref={notifRef}>
          <IconButton
            label="Notifications"
            icon={Bell}
            badge={3}
            active={notifOpen}
            onClick={() => {
              setNotifOpen((v) => !v)
              setProfileOpen(false)
            }}
          />
          {notifOpen && (
            <DropdownPanel align="right">
              <div className="border-b border-border px-3 py-2 text-sm font-medium">
                Notifications
              </div>
              <div className="max-h-72 divide-y divide-border overflow-y-auto">
                <NotificationItem
                  title="High severity alert"
                  body="Multiple failed logins from 192.0.2.41 on host db-01."
                  time="2m ago"
                />
                <NotificationItem
                  title="Pipeline degraded"
                  body="Event processor lag is above threshold."
                  time="14m ago"
                />
                <NotificationItem
                  title="New integration available"
                  body="CrowdStrike connector v2 is ready to install."
                  time="1h ago"
                />
              </div>
              <button className="w-full px-3 py-2 text-center text-xs text-muted-foreground hover:bg-muted hover:text-foreground">
                View all
              </button>
            </DropdownPanel>
          )}
        </div>

        <IconButton
          label={theme === 'dark' ? 'Light mode' : 'Dark mode'}
          icon={theme === 'dark' ? Sun : Moon}
          onClick={toggleTheme}
        />

        <div className="relative" ref={profileRef}>
          <button
            onClick={() => {
              setProfileOpen((v) => !v)
              setNotifOpen(false)
            }}
            className={cn(
              'flex h-9 items-center gap-2 rounded-md pl-1 pr-2 text-sm transition-colors',
              profileOpen
                ? 'bg-muted text-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground'
            )}
          >
            {user?.image_url ? (
              <img
                src={user.image_url}
                alt={userName}
                className="h-7 w-7 rounded-full object-cover"
              />
            ) : (
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">
                {userInitials || 'U'}
              </span>
            )}
            <ChevronDown
              size={12}
              className={cn(
                'transition-transform duration-150',
                profileOpen ? 'rotate-180' : 'rotate-0'
              )}
            />
          </button>
          {profileOpen && (
            <DropdownPanel align="right">
              <div className="border-b border-border px-3 py-2">
                <div className="truncate text-sm font-medium">{userName}</div>
                {user?.email && (
                  <div className="truncate text-xs text-muted-foreground">{user.email}</div>
                )}
              </div>
              <Link
                to="/profile"
                onClick={() => setProfileOpen(false)}
                className="flex items-center gap-2 px-3 py-2 text-sm text-popover-foreground/70 hover:bg-muted hover:text-popover-foreground"
              >
                <UserCircle size={16} strokeWidth={1.75} /> Profile
              </Link>
              <button
                onClick={handleLogout}
                className="flex w-full items-center gap-2 px-3 py-2 text-sm text-destructive/80 hover:bg-destructive/10 hover:text-destructive"
              >
                <LogOut size={16} strokeWidth={1.75} /> Log out
              </button>

              {/* License footer */}
              <div className="border-t border-border bg-muted/30 px-3 py-2">
                <div className="flex items-center justify-between gap-2 text-[11px]">
                  <span className="inline-flex items-center gap-1.5">
                    <span className={cn('h-1.5 w-1.5 rounded-full', dot)} />
                    <span
                      className={cn(
                        'font-medium tracking-wide',
                        APP_INFO.edition === 'enterprise' ? 'text-primary' : 'text-foreground'
                      )}
                    >
                      {editionLabel()}
                    </span>
                    <span className="text-muted-foreground">
                      · {STATUS_LABEL[APP_INFO.licenseStatus]}
                      {APP_INFO.trialDaysLeft != null && ` (${APP_INFO.trialDaysLeft}d left)`}
                    </span>
                  </span>
                  <span className="font-mono text-muted-foreground">v{APP_INFO.version}</span>
                </div>
                {(APP_INFO.edition === 'community' ||
                  APP_INFO.licenseStatus === 'expired' ||
                  APP_INFO.licenseStatus === 'expiring') && (
                  <Link
                    to="/settings"
                    onClick={() => setProfileOpen(false)}
                    className="mt-1.5 block w-full rounded-md bg-primary/10 px-2 py-1.5 text-center text-xs font-medium text-primary hover:bg-primary/15"
                  >
                    {APP_INFO.edition === 'community'
                      ? 'Upgrade to Enterprise'
                      : APP_INFO.licenseStatus === 'expired'
                        ? 'Renew license'
                        : 'Renew'}
                  </Link>
                )}
              </div>
            </DropdownPanel>
          )}
        </div>
      </div>
    </header>
  )
}

function IconButton({
  label,
  icon: Icon,
  onClick,
  active,
  badge,
}: {
  label: string
  icon: LucideIcon
  onClick: () => void
  active?: boolean
  badge?: number
}) {
  return (
    <button
      onClick={onClick}
      aria-label={label}
      title={label}
      className={cn(
        'relative flex h-9 w-9 items-center justify-center rounded-md transition-colors',
        active
          ? 'bg-muted text-foreground'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground'
      )}
    >
      <Icon size={18} strokeWidth={1.75} />
      {badge != null && badge > 0 && (
        <span className="absolute right-1 top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-semibold leading-none text-destructive-foreground">
          {badge > 9 ? '9+' : badge}
        </span>
      )}
    </button>
  )
}

function DropdownPanel({
  align,
  children,
}: {
  align: 'left' | 'right'
  children: React.ReactNode
}) {
  return (
    <div
      className={cn(
        'absolute top-full z-50 mt-1 w-72 overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-lg',
        align === 'left' ? 'left-0' : 'right-0'
      )}
    >
      {children}
    </div>
  )
}

function WorkspacePicker({
  current,
  workspaces,
  onPick,
  onManage,
}: {
  current: Workspace
  workspaces: Workspace[]
  onPick: (id: number) => void
  onClose: () => void
  onManage: () => void
}) {
  const [q, setQ] = useState('')
  const filtered = useMemo(
    () =>
      workspaces.filter((w) => {
        const needle = q.toLowerCase()
        return (
          w.name.toLowerCase().includes(needle) ||
          w.slug.toLowerCase().includes(needle)
        )
      }),
    [q, workspaces],
  )
  // Default workspace floats to the top, then the rest by name.
  const sorted = useMemo(
    () =>
      [...filtered].sort((a, b) => {
        if (a.is_default !== b.is_default) return a.is_default ? -1 : 1
        return a.name.localeCompare(b.name)
      }),
    [filtered],
  )
  return (
    <div className="absolute left-0 top-full z-50 mt-1 w-80 overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-lg">
      {/* Search — only useful when there's more than a handful */}
      {workspaces.length > 5 && (
        <div className="border-b border-border p-2">
          <div className="relative">
            <Search size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder="Find a workspace…"
              autoFocus
              className="h-8 w-full rounded-md border-0 bg-muted/40 pl-7 pr-2 text-xs outline-none placeholder:text-muted-foreground/70 focus:bg-card focus:ring-1 focus:ring-border"
            />
          </div>
        </div>
      )}

      {/* Workspaces list */}
      <div className="max-h-72 overflow-y-auto p-1">
        <div className="px-2 py-1 text-[10px] uppercase tracking-wider text-muted-foreground">
          Your workspaces
        </div>
        {sorted.map((w) => {
          const active = w.id === current.id
          return (
            <button
              key={w.id}
              onClick={() => onPick(w.id)}
              className={cn(
                'flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left transition-colors',
                active ? 'bg-muted/60' : 'hover:bg-muted/60',
              )}
            >
              <span
                className={cn(
                  'flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-gradient-to-br text-[10px] font-semibold text-white',
                  workspaceTone(w.slug),
                )}
              >
                {workspaceInitials(w.name)}
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-1.5">
                  <span className="truncate text-sm font-medium">{w.name}</span>
                  {w.is_default && (
                    <Star
                      size={10}
                      strokeWidth={2.5}
                      className="shrink-0 text-amber-500"
                      aria-label="Default workspace"
                    />
                  )}
                </div>
                <div className="truncate font-mono text-[11px] text-muted-foreground">
                  {w.slug}
                </div>
              </div>
              {active && <Check size={14} className="shrink-0 text-primary" />}
            </button>
          )
        })}
        {sorted.length === 0 && (
          <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">
            No workspaces match.
          </div>
        )}
      </div>

      {/* Footer action — Manage takes the user to /settings/workspaces, where
          they can create, edit, and delete. We intentionally don't surface a
          create CTA here to keep the picker focused on switching context. */}
      <div className="border-t border-border">
        <button
          onClick={onManage}
          className="flex w-full items-center gap-2 px-3 py-2 text-xs text-popover-foreground/80 hover:bg-muted hover:text-popover-foreground"
        >
          <Settings size={13} strokeWidth={1.75} />
          Manage workspaces
        </button>
      </div>
    </div>
  )
}

function NotificationItem({ title, body, time }: { title: string; body: string; time: string }) {
  return (
    <div className="group cursor-pointer px-3 py-2 hover:bg-muted">
      <div className="flex items-start justify-between gap-2">
        <div className="text-sm font-medium text-popover-foreground/80 group-hover:text-popover-foreground">
          {title}
        </div>
        <div className="shrink-0 text-[10px] text-muted-foreground">{time}</div>
      </div>
      <div className="mt-0.5 line-clamp-2 text-xs text-muted-foreground group-hover:text-popover-foreground/70">
        {body}
      </div>
    </div>
  )
}
