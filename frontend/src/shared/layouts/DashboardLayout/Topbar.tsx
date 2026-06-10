import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Bell,
  Check,
  CheckCheck,
  ChevronDown,
  ChevronLeft,
  KeyRound,
  Languages,
  Loader2,
  LogOut,
  Moon,
  Sun,
  UserCircle,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers'
import { useAuth } from '@/features/auth'
import { useBilling } from '@/features/billing'
import {
  NotificationRow,
  useNotificationFeed,
  useNotifications,
  type Notification,
} from '@/features/notifications'
import { SUPPORTED_LANGUAGES } from '@/shared/i18n'

export function Topbar() {
  const [notifOpen, setNotifOpen] = useState(false)
  const [profileOpen, setProfileOpen] = useState(false)
  const [langOpen, setLangOpen] = useState(false)
  const { theme, toggleTheme } = useThemeContext()
  const { user, logout, updateMe } = useAuth()
  const { license, version } = useBilling()
  const { unreadCount, markRead, remove, markAllRead, refreshUnread } = useNotifications()
  const {
    items: notifItems,
    setItems: setNotifItems,
    loading: notifLoading,
    hasMore: notifHasMore,
    loadMore: loadNotifs,
    reset: resetNotifs,
  } = useNotificationFeed(5)
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const currentLang =
    SUPPORTED_LANGUAGES.find((l) => l.code === i18n.language)?.code ??
    SUPPORTED_LANGUAGES.find((l) => i18n.language?.startsWith(l.code))?.code ??
    'en'

  const handleLanguageChange = (code: string) => {
    void i18n.changeLanguage(code)
    // Persist to the user's saved preference (best-effort; UI already switched).
    void updateMe({ lang_key: code }).catch(() => {})
  }

  // Load the notification feed each time the bell opens (fresh, page 0).
  useEffect(() => {
    if (notifOpen) {
      resetNotifs()
      void loadNotifs()
    }
  }, [notifOpen, resetNotifs, loadNotifs])

  const onNotifScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    if (el.scrollHeight - el.scrollTop - el.clientHeight < 60 && notifHasMore && !notifLoading) {
      void loadNotifs()
    }
  }

  const toggleNotifRead = async (n: Notification) => {
    setNotifItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: !n.read } : x)))
    try {
      await markRead(n.id, !n.read)
    } catch {
      setNotifItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: n.read } : x)))
    }
  }

  const deleteNotif = async (n: Notification) => {
    setNotifItems((prev) => prev.filter((x) => x.id !== n.id))
    try {
      await remove(n.id)
    } catch {
      void refreshUnread()
    }
  }

  const markAllNotifsRead = async () => {
    setNotifItems((prev) => prev.map((x) => ({ ...x, read: true })))
    try {
      await markAllRead()
    } catch {
      void refreshUnread()
    }
  }

  const notifRef = useRef<HTMLDivElement>(null)
  const profileRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      const target = e.target as Node
      if (notifRef.current && !notifRef.current.contains(target)) setNotifOpen(false)
      if (profileRef.current && !profileRef.current.contains(target)) setProfileOpen(false)
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

  return (
    <header className="relative z-40 flex h-14 shrink-0 items-center justify-between gap-3 border-b border-border bg-card px-3">
      {/* Left cluster */}
      <div className="flex min-w-0 items-center gap-2">
        <Link to="/home" className="flex items-center gap-2 px-1">
          <img src="/logo.svg" alt="UTMStack" className="h-7 w-7" />
          <span className="text-lg font-normal tracking-tight">UTMStack</span>
        </Link>

        {/* Version + edition badge. Community links to the License page (upgrade path);
            Enterprise is a static label. */}
        {license && version &&
          (license.edition === 'enterprise' ? (
            <span className="hidden items-center gap-1.5 rounded-full border border-border bg-muted/40 px-2.5 py-1 text-[11px] sm:inline-flex">
              <span className="font-mono text-muted-foreground">{version.version}</span>
              <span className="h-3 w-px bg-border" />
              <span className="font-medium text-primary">Enterprise</span>
            </span>
          ) : (
            <Link
              to="/settings/license"
              title={t('topbar.manageLicense')}
              className="hidden items-center gap-1.5 rounded-full border border-border bg-muted/40 px-2.5 py-1 text-[11px] transition-colors hover:border-primary/40 hover:bg-primary/5 sm:inline-flex"
            >
              <span className="font-mono text-muted-foreground">{version.version}</span>
              <span className="h-3 w-px bg-border" />
              <span className="font-medium">Community</span>
            </Link>
          ))}
      </div>

      {/* Right cluster */}
      <div className="flex items-center gap-1">
        <div className="relative" ref={notifRef}>
          <IconButton
            label={t('notifications.title')}
            icon={Bell}
            badge={unreadCount}
            active={notifOpen}
            onClick={() => {
              setNotifOpen((v) => !v)
              setProfileOpen(false)
            }}
          />
          {notifOpen && (
            <DropdownPanel align="right">
              <div className="flex items-center justify-between border-b border-border px-3 py-2">
                <span className="text-sm font-medium">{t('notifications.title')}</span>
                <button
                  type="button"
                  onClick={() => void markAllNotifsRead()}
                  disabled={unreadCount === 0}
                  className="flex items-center gap-1 text-[11px] text-muted-foreground transition-colors hover:text-foreground disabled:opacity-40"
                >
                  <CheckCheck size={12} />
                  {t('notifications.markAllRead')}
                </button>
              </div>
              <div
                onScroll={onNotifScroll}
                className="max-h-80 divide-y divide-border overflow-y-auto"
              >
                {notifItems.map((n) => (
                  <NotificationRow
                    key={n.id}
                    notification={n}
                    onToggleRead={(x) => void toggleNotifRead(x)}
                    onDelete={(x) => void deleteNotif(x)}
                  />
                ))}
                {notifItems.length === 0 && !notifLoading && (
                  <div className="px-3 py-10 text-center text-xs text-muted-foreground">
                    {t('notifications.empty')}
                  </div>
                )}
                {notifLoading && (
                  <div className="flex items-center justify-center py-3 text-muted-foreground">
                    <Loader2 className="h-4 w-4 animate-spin" />
                  </div>
                )}
              </div>
              <Link
                to="/settings/notifications"
                onClick={() => setNotifOpen(false)}
                className="block border-t border-border px-3 py-2 text-center text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
              >
                {t('notifications.viewAll')}
              </Link>
            </DropdownPanel>
          )}
        </div>

        <IconButton
          label={theme === 'dark' ? t('topbar.theme.light') : t('topbar.theme.dark')}
          icon={theme === 'dark' ? Sun : Moon}
          onClick={toggleTheme}
        />

        <div className="relative" ref={profileRef}>
          <button
            onClick={() => {
              setProfileOpen((v) => !v)
              setNotifOpen(false)
              setLangOpen(false)
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
            <DropdownPanel align="right" overflowVisible>
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
                <UserCircle size={16} strokeWidth={1.75} /> {t('topbar.profile.profile')}
              </Link>
              <Link
                to="/settings/license"
                onClick={() => setProfileOpen(false)}
                className="flex items-center gap-2 px-3 py-2 text-sm text-popover-foreground/70 hover:bg-muted hover:text-popover-foreground"
              >
                <KeyRound size={16} strokeWidth={1.75} /> {t('topbar.profile.license')}
              </Link>
              <div className="relative">
                <button
                  type="button"
                  onClick={() => setLangOpen((v) => !v)}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-2 text-sm transition-colors hover:bg-muted',
                    langOpen
                      ? 'bg-muted text-popover-foreground'
                      : 'text-popover-foreground/70 hover:text-popover-foreground',
                  )}
                >
                  <Languages size={16} strokeWidth={1.75} />
                  <span className="flex-1 text-left">{t('common.language')}</span>
                  <ChevronLeft size={14} className="text-muted-foreground" />
                </button>
                {langOpen && (
                  <div className="absolute right-full top-0 z-50 mr-1 min-w-[160px] overflow-hidden rounded-md border border-border bg-popover py-1 text-popover-foreground shadow-lg">
                    {SUPPORTED_LANGUAGES.map((l) => {
                      const active = l.code === currentLang
                      return (
                        <button
                          key={l.code}
                          type="button"
                          onClick={() => {
                            handleLanguageChange(l.code)
                            setLangOpen(false)
                            setProfileOpen(false)
                          }}
                          className={cn(
                            'flex w-full items-center justify-between gap-3 px-3 py-1.5 text-sm transition-colors hover:bg-muted',
                            active ? 'text-popover-foreground' : 'text-popover-foreground/70',
                          )}
                        >
                          {l.label}
                          {active && <Check size={14} className="shrink-0 text-primary" />}
                        </button>
                      )
                    })}
                  </div>
                )}
              </div>
              <button
                onClick={handleLogout}
                className="flex w-full items-center gap-2 rounded-b-md border-t border-border px-3 py-2 text-sm text-destructive/80 hover:bg-destructive/10 hover:text-destructive"
              >
                <LogOut size={16} strokeWidth={1.75} /> {t('topbar.profile.logout')}
              </button>
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
  overflowVisible,
}: {
  align: 'left' | 'right'
  children: React.ReactNode
  // Allow a nested side flyout (e.g. the language submenu) to escape the panel.
  overflowVisible?: boolean
}) {
  return (
    <div
      className={cn(
        'absolute top-full z-50 mt-1 w-72 rounded-md border border-border bg-popover text-popover-foreground shadow-lg',
        overflowVisible ? 'overflow-visible' : 'overflow-hidden',
        align === 'left' ? 'left-0' : 'right-0'
      )}
    >
      {children}
    </div>
  )
}
