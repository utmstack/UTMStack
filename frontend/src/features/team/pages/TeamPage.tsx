import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Check,
  Crown,
  KeyRound,
  Loader2,
  Pencil,
  Search,
  Shield,
  ShieldCheck,
  ShieldOff,
  UserCheck,
  UserPlus,
  UserX,
  X,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { SUPPORTED_LANGUAGES } from '@/shared/i18n'
import { rolesHttpService, TeamHttpError, usersHttpService } from '../services/team-http.service'
import type {
  PageInfo,
  Role,
  RoleDetail,
  UserBase,
  UserDetail,
  UserListItem,
} from '../types/team.types'

const PAGE_SIZE = 20

/* Roles & permissions arrive from the backend as English DB strings. Translate the
 * stable identifiers (role name, permission resource/action) with a fallback to the
 * backend value so any custom/unknown role still renders something sensible. */
function roleLabel(t: TFunction, name: string, fallback?: string): string {
  return t(`team.roleNames.${name}`, { defaultValue: fallback || name })
}
function roleDesc(t: TFunction, name: string, fallback?: string): string {
  return t(`team.roleDescriptions.${name}`, { defaultValue: fallback || '' })
}
function resourceLabel(t: TFunction, resource: string): string {
  return t(`team.resources.${resource}`, { defaultValue: resource })
}
function actionLabel(t: TFunction, action: string): string {
  return t(`team.actions.${action}`, { defaultValue: action })
}

/* ─────────────────────────────────────────────────────────────────────────
 * Page
 * ───────────────────────────────────────────────────────────────────────── */

type View = 'members' | 'roles'

export function TeamPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<View>('members')
  const [roles, setRoles] = useState<Role[]>([])

  // Roles list is small and shared by the pickers + the roles tab.
  useEffect(() => {
    let cancelled = false
    rolesHttpService
      .list()
      .then((r) => {
        if (!cancelled) setRoles(r)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('team.toast.rolesLoadFailed'))
      })
    return () => {
      cancelled = true
    }
  }, [t])

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <UserCheck size={14} strokeWidth={1.75} />
          <span className="font-medium text-foreground">{t('team.title')}</span>
        </div>
      </header>

      <div className="mt-3 flex items-center gap-1 border-b border-border">
        <TabBtn active={view === 'members'} onClick={() => setView('members')} icon={UserCheck}>
          {t('team.tabs.members')}
        </TabBtn>
        <TabBtn active={view === 'roles'} onClick={() => setView('roles')} icon={Shield}>
          {t('team.tabs.roles')}
        </TabBtn>
      </div>

      {view === 'members' ? <MembersView roles={roles} /> : <RolesView roles={roles} />}
    </div>
  )
}

function TabBtn({
  active,
  onClick,
  icon: Icon,
  children,
}: {
  active: boolean
  onClick: () => void
  icon: LucideIcon
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'relative flex items-center gap-1.5 px-3 py-2 text-sm transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
      )}
    >
      <Icon size={14} strokeWidth={1.75} />
      {children}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Members view
 * ───────────────────────────────────────────────────────────────────────── */

function MembersView({ roles }: { roles: Role[] }) {
  const { t } = useTranslation()
  const [users, setUsers] = useState<UserListItem[] | null>(null)
  const [pageInfo, setPageInfo] = useState<PageInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')

  const [openId, setOpenId] = useState<number | null>(null)
  const [inviteOpen, setInviteOpen] = useState(false)

  // Debounce the search box, and reset to page 1 when the query changes.
  useEffect(() => {
    const handle = setTimeout(() => {
      setDebounced(search.trim())
      setPage(1)
    }, 300)
    return () => clearTimeout(handle)
  }, [search])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const resp = await usersHttpService.list({ page, page_size: PAGE_SIZE, search: debounced })
      setUsers((prev) => (page === 1 ? resp.data : [...(prev ?? []), ...resp.data]))
      setPageInfo(resp.page_info)
    } catch {
      setError(true)
      setUsers([])
    } finally {
      setLoading(false)
    }
  }, [page, debounced])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <>
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('team.members.searchPlaceholder')}
            className="h-9 pl-9"
          />
        </div>
        <Button size="sm" onClick={() => setInviteOpen(true)}>
          <UserPlus size={14} className="mr-1.5" />
          {t('team.members.invite')}
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: MEMBER_COLS }}
        >
          <div>{t('team.members.colUser')}</div>
          <div>{t('team.members.colRoles')}</div>
          <div>{t('team.members.col2fa')}</div>
          <div>{t('team.members.colStatus')}</div>
          <div />
        </div>

        {loading && (!users || users.length === 0) && (
          <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t('team.members.loading')}
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-col items-center gap-3 px-6 py-12 text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('team.members.loadFailed')}
            </span>
            <Button variant="outline" size="sm" onClick={() => void load()}>
              {t('team.members.retry')}
            </Button>
          </div>
        )}

        {!loading && !error && users && users.length === 0 && (
          <div className="px-6 py-16 text-center text-sm text-muted-foreground">
            {debounced ? t('team.members.noSearchMatch') : t('team.members.none')}
          </div>
        )}

        {users && users.length > 0 &&
          users.map((u) => <MemberRow key={u.id} user={u} onOpen={() => setOpenId(u.id)} />)}
      </div>

      {users && users.length > 0 && (
        <InfiniteScrollSentinel
          onReach={() => setPage((p) => p + 1)}
          hasMore={users.length < (pageInfo?.total_items ?? 0)}
          loading={loading}
          endLabel={t('common.allLoaded', { count: pageInfo?.total_items ?? 0 })}
        />
      )}

      {openId != null && (
        <MemberDrawer
          userId={openId}
          roles={roles}
          onClose={() => setOpenId(null)}
          onChanged={() => void load()}
        />
      )}
      {inviteOpen && (
        <InviteDialog
          roles={roles}
          onClose={() => setInviteOpen(false)}
          onCreated={() => {
            setInviteOpen(false)
            setPage(1)
            void load()
          }}
        />
      )}
    </>
  )
}

const MEMBER_COLS = '1.7fr 1fr 90px 130px 40px'

function MemberRow({ user: u, onOpen }: { user: UserListItem; onOpen: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      onClick={onOpen}
      className={cn(
        'grid cursor-pointer items-center gap-3 border-b border-border px-4 py-3 text-xs last:border-b-0 hover:bg-muted/40',
        !u.activated && 'opacity-70'
      )}
      style={{ gridTemplateColumns: MEMBER_COLS }}
    >
      <div className="flex min-w-0 items-center gap-3">
        <Avatar user={u} />
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">{fullName(u)}</div>
          <div className="truncate text-[11px] text-muted-foreground">
            <span className="font-mono">@{u.login}</span>
            {u.email ? ` · ${u.email}` : ''}
          </div>
        </div>
      </div>
      <div className="flex flex-wrap gap-1">
        {(u.roles ?? []).length === 0 ? (
          <span className="text-[11px] italic text-muted-foreground">{t('team.members.noRole')}</span>
        ) : (
          u.roles!.map((r) => <RoleBadge key={r.name} name={r.name} label={r.display_name} />)
        )}
      </div>
      <div>
        {u.tfa_enabled ? (
          <span className="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
            <ShieldCheck size={12} /> {u.tfa_method ?? t('team.members.on')}
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 text-muted-foreground">
            <ShieldOff size={12} /> {t('team.members.off')}
          </span>
        )}
      </div>
      <div>
        <StatusBadge activated={u.activated} defaultPassword={u.default_password} />
      </div>
      <div className="flex justify-end text-muted-foreground">
        <Pencil size={13} />
      </div>
    </div>
  )
}

function StatusBadge({ activated, defaultPassword }: { activated: boolean; defaultPassword: boolean }) {
  const { t } = useTranslation()
  if (!activated) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground ring-1 ring-inset ring-border">
        <span className="h-1.5 w-1.5 rounded-full bg-zinc-400" />
        {t('team.status.inactive')}
      </span>
    )
  }
  if (defaultPassword) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-md bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300">
        <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
        {t('team.status.pendingPassword')}
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5 rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600 ring-1 ring-inset ring-emerald-500/30 dark:text-emerald-300">
      <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
      {t('team.status.active')}
    </span>
  )
}

function RoleBadge({ name, label }: { name: string; label: string }) {
  const { t } = useTranslation()
  const isAdmin = name === 'ROLE_ADMIN'
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
        isAdmin
          ? 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300'
          : 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300'
      )}
    >
      {isAdmin ? <Crown size={10} /> : <Shield size={10} />}
      {roleLabel(t, name, label)}
    </span>
  )
}

function Avatar({ user: u, size = 36 }: { user: UserBase; size?: number }) {
  if (u.image_url) {
    return (
      <img
        src={u.image_url}
        alt={fullName(u)}
        className="shrink-0 rounded-full object-cover"
        style={{ height: size, width: size }}
      />
    )
  }
  return (
    <span
      className="flex shrink-0 items-center justify-center rounded-full bg-primary text-[11px] font-semibold text-primary-foreground"
      style={{ height: size, width: size }}
    >
      {initials(u)}
    </span>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Member drawer — load detail, edit roles, activate/deactivate
 * ───────────────────────────────────────────────────────────────────────── */

function MemberDrawer({
  userId,
  roles,
  onClose,
  onChanged,
}: {
  userId: number
  roles: Role[]
  onClose: () => void
  onChanged: () => void
}) {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [roleNames, setRoleNames] = useState<string[]>([])
  const [confirmResetTfa, setConfirmResetTfa] = useState(false)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    usersHttpService
      .get(userId)
      .then((u) => {
        if (cancelled) return
        setUser(u)
        setRoleNames((u.roles ?? []).map((r) => r.name))
      })
      .catch(() => {
        if (!cancelled) {
          toast.error(t('team.toast.userLoadFailed'))
          onClose()
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [userId, onClose, t])

  // Pending invite = created but never accepted (still on the temporary password,
  // not yet activated). It can't be reactivated — the user must accept first.
  const isPending = !!user && !user.activated && user.default_password

  // Admins only manage roles + account status. Profile details (name, email,
  // language) are edited by the user themselves on their own Profile page.
  const baseRoles = (user?.roles ?? []).map((r) => r.name)
  const dirty = roleNames.slice().sort().join(',') !== baseRoles.slice().sort().join(',')

  const toggleRole = (name: string) =>
    setRoleNames((rs) => (rs.includes(name) ? rs.filter((r) => r !== name) : [...rs, name]))

  const save = async () => {
    if (!user || !dirty) return
    setBusy(true)
    try {
      await usersHttpService.assignRoles(user.id, roleNames)
      toast.success(t('team.toast.rolesUpdated'))
      onChanged()
      onClose()
    } catch (err) {
      toast.error(userError(err, t))
    } finally {
      setBusy(false)
    }
  }

  const resetTfa = async () => {
    if (!user) return
    setBusy(true)
    try {
      await usersHttpService.resetTfa(user.id)
      toast.success(t('team.toast.tfaReset'))
      setConfirmResetTfa(false)
      onChanged()
      onClose()
    } catch (err) {
      toast.error(userError(err, t))
    } finally {
      setBusy(false)
    }
  }

  const setActivated = async (activated: boolean) => {
    if (!user) return
    setBusy(true)
    try {
      if (activated) {
        await usersHttpService.update(user.id, { activated: true })
        toast.success(t('team.toast.userReactivated'))
      } else {
        await usersHttpService.deactivate(user.id)
        toast.success(t('team.toast.userDeactivated'))
      }
      onChanged()
      onClose()
    } catch (err) {
      toast.error(userError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[640px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {loading || !user ? (
          <div className="flex flex-1 items-center justify-center text-muted-foreground">
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            {t('team.drawer.loading')}
          </div>
        ) : (
          <>
            <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-5">
              <div className="flex min-w-0 items-center gap-3">
                <Avatar user={user} size={48} />
                <div className="min-w-0">
                  <h2 className="truncate text-lg font-semibold">{fullName(user)}</h2>
                  <div className="truncate text-xs text-muted-foreground">
                    <span className="font-mono">@{user.login}</span>
                    {user.created_by ? ` · ${t('team.drawer.createdBy', { actor: user.created_by })}` : ''}
                    {user.created_date ? ` ${relativeTime(user.created_date, t)}` : ''}
                  </div>
                </div>
              </div>
              <button
                onClick={onClose}
                className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
              >
                <X size={16} />
              </button>
            </header>

            <div className="flex-1 space-y-6 overflow-y-auto p-6">
              <DrawerSection title={t('team.drawer.profileTitle')} subtitle={t('team.drawer.profileSubtitle')}>
                <dl className="grid grid-cols-[110px_1fr] gap-y-2.5 text-xs">
                  <InfoRow k={t('team.drawer.name')}>
                    {[user.first_name, user.last_name].filter(Boolean).join(' ') || '—'}
                  </InfoRow>
                  <InfoRow k={t('team.drawer.username')}>
                    <span className="font-mono">@{user.login}</span>
                  </InfoRow>
                  <InfoRow k={t('team.drawer.email')}>{user.email || '—'}</InfoRow>
                  <InfoRow k={t('team.drawer.language')}>
                    {user.lang_key ? langLabel(user.lang_key) : t('team.langDefault')}
                  </InfoRow>
                </dl>
              </DrawerSection>

              <DrawerSection title={t('team.drawer.accountTitle')}>
                <ul className="space-y-1 text-[11px] text-muted-foreground">
                  <li>
                    {t('team.drawer.status')}:{' '}
                    {user.activated ? (
                      <span className="text-emerald-600 dark:text-emerald-400">
                        {t('team.drawer.statusActive')}
                      </span>
                    ) : (
                      <span>{t('team.drawer.statusInactive')}</span>
                    )}
                  </li>
                  <li className="flex items-center gap-2">
                    <span>
                      {t('team.drawer.tfa')}:{' '}
                      {user.tfa_enabled ? (
                        <span className="text-emerald-600 dark:text-emerald-400">
                          {t('team.drawer.tfaEnabled', { method: user.tfa_method })}
                        </span>
                      ) : (
                        t('team.drawer.tfaNotConfigured')
                      )}
                    </span>
                    {user.tfa_enabled && (
                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => setConfirmResetTfa(true)}
                        className="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium text-red-500 hover:bg-red-500/10 disabled:opacity-50"
                        title={t('team.drawer.resetTfaHint')}
                      >
                        <ShieldOff size={11} />
                        {t('team.drawer.resetTfa')}
                      </button>
                    )}
                  </li>
                  {user.default_password && (
                    <li className="text-amber-600 dark:text-amber-300">
                      {t('team.drawer.pendingPasswordNote')}
                    </li>
                  )}
                  {user.last_modified_date && (
                    <li>{t('team.drawer.lastModified', { time: relativeTime(user.last_modified_date, t) })}</li>
                  )}
                </ul>
              </DrawerSection>

              <DrawerSection title={t('team.drawer.rolesTitle')} subtitle={t('team.drawer.rolesSubtitle')}>
                <RolePicker roles={roles} selected={roleNames} onToggle={toggleRole} />
                {roleNames.length === 0 && (
                  <p className="mt-2 text-[11px] text-amber-600 dark:text-amber-300">
                    {t('team.drawer.noRoleWarning')}
                  </p>
                )}
              </DrawerSection>
            </div>

            <footer className="flex items-center justify-between gap-3 border-t border-border px-6 py-3">
              {user.activated ? (
                <Button
                  variant="outline"
                  size="sm"
                  className="text-red-500 hover:bg-red-500/10"
                  disabled={busy}
                  onClick={() => void setActivated(false)}
                >
                  <UserX size={13} className="mr-1.5" />
                  {t('team.drawer.deactivate')}
                </Button>
              ) : isPending ? (
                <span className="text-[11px] text-muted-foreground">{t('team.drawer.waitingInvite')}</span>
              ) : (
                <Button variant="outline" size="sm" disabled={busy} onClick={() => void setActivated(true)}>
                  <UserCheck size={13} className="mr-1.5" />
                  {t('team.drawer.reactivate')}
                </Button>
              )}
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm" onClick={onClose}>
                  {t('team.drawer.cancel')}
                </Button>
                <Button size="sm" disabled={!dirty || busy} onClick={() => void save()}>
                  {busy ? t('team.drawer.saving') : t('team.drawer.saveRoles')}
                </Button>
              </div>
            </footer>

            {confirmResetTfa && (
              <div
                className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
                onClick={() => !busy && setConfirmResetTfa(false)}
              >
                <div
                  className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
                  onClick={(e) => e.stopPropagation()}
                >
                  <header className="flex items-center gap-2 border-b border-border px-6 py-4">
                    <ShieldOff size={17} className="text-red-500" />
                    <h2 className="text-base font-semibold">{t('team.drawer.resetTfaConfirmTitle')}</h2>
                  </header>
                  <div className="px-6 py-5 text-sm text-muted-foreground">
                    {t('team.drawer.resetTfaConfirmBody', { name: fullName(user) })}
                  </div>
                  <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
                    <Button variant="outline" size="sm" disabled={busy} onClick={() => setConfirmResetTfa(false)}>
                      {t('team.drawer.cancel')}
                    </Button>
                    <Button variant="destructive" size="sm" disabled={busy} onClick={() => void resetTfa()}>
                      {busy ? t('team.drawer.saving') : t('team.drawer.resetTfaConfirm')}
                    </Button>
                  </footer>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Invite dialog — create user (no password; invite-by-email like the backend)
 * ───────────────────────────────────────────────────────────────────────── */

function InviteDialog({
  roles,
  onClose,
  onCreated,
}: {
  roles: Role[]
  onClose: () => void
  onCreated: () => void
}) {
  const { t } = useTranslation()
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [roleNames, setRoleNames] = useState<string[]>(['ROLE_USER'])
  const [busy, setBusy] = useState(false)

  const valid = login.trim().length >= 3 && /.+@.+\..+/.test(email)

  const toggleRole = (name: string) =>
    setRoleNames((rs) => (rs.includes(name) ? rs.filter((r) => r !== name) : [...rs, name]))

  const submit = async () => {
    if (!valid || busy) return
    setBusy(true)
    try {
      await usersHttpService.create({
        login: login.trim(),
        email: email.trim(),
        first_name: firstName.trim() || undefined,
        last_name: lastName.trim() || undefined,
        // No lang_key: the user inherits the platform-default language until they
        // pick their own in their profile.
        role_names: roleNames,
      })
      toast.success(t('team.toast.invitationSent'))
      onCreated()
    } catch (err) {
      toast.error(userError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <UserPlus size={18} />
              {t('team.invite.title')}
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">{t('team.invite.description')}</p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="space-y-4 px-6 py-5">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DrawerField label={t('team.invite.username')} hint={t('team.invite.usernameHint')}>
              <Input value={login} onChange={(e) => setLogin(e.target.value)} className="font-mono" placeholder="jsmith" />
            </DrawerField>
            <DrawerField label={t('team.invite.email')}>
              <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="jsmith@company.com" />
            </DrawerField>
          </div>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DrawerField label={t('team.invite.firstName')}>
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </DrawerField>
            <DrawerField label={t('team.invite.lastName')}>
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </DrawerField>
          </div>
          <div>
            <label className="mb-1.5 block text-xs font-medium text-foreground/80">
              {t('team.invite.roles')}
            </label>
            <RolePicker roles={roles} selected={roleNames} onToggle={toggleRole} compact />
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('team.invite.cancel')}
          </Button>
          <Button size="sm" disabled={!valid || busy} onClick={() => void submit()}>
            <UserPlus size={13} className="mr-1.5" />
            {busy ? t('team.invite.sending') : t('team.invite.send')}
          </Button>
        </footer>
      </div>
    </div>
  )
}

function RolePicker({
  roles,
  selected,
  onToggle,
  compact,
}: {
  roles: Role[]
  selected: string[]
  onToggle: (name: string) => void
  compact?: boolean
}) {
  const { t } = useTranslation()
  if (roles.length === 0) {
    return <p className="text-[11px] text-muted-foreground">{t('team.rolePicker.none')}</p>
  }
  return (
    <div className="space-y-2">
      {roles.map((r) => {
        const checked = selected.includes(r.name)
        const isAdmin = r.name === 'ROLE_ADMIN'
        return (
          <button
            key={r.name}
            type="button"
            onClick={() => onToggle(r.name)}
            className={cn(
              'flex w-full items-start gap-3 rounded-md border p-2.5 text-left transition-colors',
              checked ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
            )}
          >
            <span
              className={cn(
                'mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded border',
                checked ? 'border-primary bg-primary text-primary-foreground' : 'border-input'
              )}
            >
              {checked && <Check size={11} strokeWidth={3} />}
            </span>
            <span className="min-w-0">
              <span className="flex items-center gap-1.5 text-sm font-medium">
                {isAdmin ? <Crown size={12} className="text-amber-500" /> : <Shield size={12} className="text-sky-500" />}
                {roleLabel(t, r.name, r.display_name)}
                {!compact && <code className="font-mono text-[10px] text-muted-foreground">{r.name}</code>}
              </span>
              {roleDesc(t, r.name, r.description) && (
                <span className="text-[11px] text-muted-foreground">{roleDesc(t, r.name, r.description)}</span>
              )}
            </span>
          </button>
        )
      })}
    </div>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Roles view — read-only catalog + permission matrix (loaded from the backend)
 * ───────────────────────────────────────────────────────────────────────── */

function RolesView({ roles }: { roles: Role[] }) {
  const { t } = useTranslation()
  const [details, setDetails] = useState<RoleDetail[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  useEffect(() => {
    if (roles.length === 0) return
    let cancelled = false
    setLoading(true)
    setError(false)
    Promise.all(roles.map((r) => rolesHttpService.get(r.name)))
      .then((d) => {
        if (!cancelled) setDetails(d)
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [roles])

  // Union of all permissions across roles, grouped by resource (rows of the matrix).
  const byResource = useMemo(() => {
    const map = new Map<string, Map<string, { name: string; description?: string }>>()
    details?.forEach((d) =>
      d.permissions.forEach((p) => {
        if (!map.has(p.resource)) map.set(p.resource, new Map())
        map.get(p.resource)!.set(p.name, { name: p.name, description: p.description })
      })
    )
    return [...map.entries()].map(([resource, perms]) => ({
      resource,
      perms: [...perms.values()].sort((a, b) => a.name.localeCompare(b.name)),
    }))
  }, [details])

  const has = (roleName: string, permName: string) =>
    details?.find((d) => d.name === roleName)?.permissions.some((p) => p.name === permName) ?? false

  if (loading) {
    return (
      <div className="mt-10 flex items-center justify-center gap-2 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {t('team.roles.loading')}
      </div>
    )
  }
  if (error || !details) {
    return (
      <div className="mt-6 rounded-md border border-amber-500/30 bg-amber-500/5 p-4 text-sm text-muted-foreground">
        {t('team.roles.loadFailed')}
      </div>
    )
  }

  return (
    <div className="mt-5 space-y-5">
      <div className="rounded-md border border-border bg-muted/30 px-4 py-3 text-xs text-muted-foreground">
        {t('team.roles.readonlyNote')}
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {details.map((d) => (
          <div key={d.name} className="rounded-xl border border-border bg-card p-5">
            <div className="flex items-center gap-2">
              {d.name === 'ROLE_ADMIN' ? (
                <Crown size={18} className="text-amber-500" />
              ) : (
                <Shield size={18} className="text-sky-500" />
              )}
              <div>
                <div className="text-sm font-semibold">{roleLabel(t, d.name, d.display_name)}</div>
                <code className="font-mono text-[10px] text-muted-foreground">{d.name}</code>
              </div>
            </div>
            {roleDesc(t, d.name, d.description) && (
              <p className="mt-2 text-xs text-muted-foreground">{roleDesc(t, d.name, d.description)}</p>
            )}
            <div className="mt-3 flex items-center gap-3 text-[11px] text-muted-foreground">
              <span className="inline-flex items-center gap-1">
                <KeyRound size={11} />
                {t('team.roles.permissionsCount', { count: d.permissions.length })}
              </span>
              <span>·</span>
              <span>
                {t('team.roles.modulesCount', { count: new Set(d.permissions.map((p) => p.resource)).size })}
              </span>
            </div>
          </div>
        ))}
      </div>

      <div className="overflow-hidden rounded-xl border border-border bg-card">
        <div className="border-b border-border px-5 py-3">
          <h3 className="text-sm font-semibold">{t('team.roles.matrixTitle')}</h3>
          <p className="mt-0.5 text-xs text-muted-foreground">{t('team.roles.matrixSubtitle')}</p>
        </div>
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-5 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: matrixCols(details.length) }}
        >
          <div>{t('team.roles.permission')}</div>
          {details.map((d) => (
            <div key={d.name} className="text-center">
              {roleLabel(t, d.name, d.display_name)}
            </div>
          ))}
        </div>
        {byResource.map(({ resource, perms }) => (
          <div key={resource}>
            <div className="bg-muted/20 px-5 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
              {resourceLabel(t, resource)}
            </div>
            {perms.map((p) => (
              <div
                key={p.name}
                className="grid items-center gap-3 border-b border-border/50 px-5 py-2 text-xs last:border-b-0"
                style={{ gridTemplateColumns: matrixCols(details.length) }}
              >
                <div>
                  <code className="font-mono text-[11px]">{p.name}</code>
                  <div className="text-[11px] text-muted-foreground">
                    {actionLabel(t, p.name.split('.')[1] ?? '')}
                  </div>
                </div>
                {details.map((d) => (
                  <PermTick key={d.name} on={has(d.name, p.name)} />
                ))}
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}

function matrixCols(n: number): string {
  return `1fr ${Array.from({ length: n }, () => '110px').join(' ')}`
}

function PermTick({ on }: { on: boolean }) {
  return (
    <div className="flex justify-center">
      {on ? (
        <Check size={14} className="text-emerald-500" strokeWidth={3} />
      ) : (
        <span className="text-muted-foreground/40">—</span>
      )}
    </div>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Small shared parts + helpers
 * ───────────────────────────────────────────────────────────────────────── */

function DrawerSection({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: React.ReactNode
}) {
  return (
    <section>
      <div className="mb-3">
        <h3 className="text-sm font-semibold">{title}</h3>
        {subtitle && <p className="mt-0.5 text-[11px] text-muted-foreground">{subtitle}</p>}
      </div>
      <div className="space-y-4">{children}</div>
    </section>
  )
}

function DrawerField({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function InfoRow({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

function langLabel(code: string): string {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code)?.label ?? code
}

function fullName(u: UserBase): string {
  return [u.first_name, u.last_name].filter(Boolean).join(' ') || u.login
}

function initials(u: UserBase): string {
  return (
    fullName(u)
      .split(' ')
      .map((p) => p[0])
      .filter(Boolean)
      .slice(0, 2)
      .join('')
      .toUpperCase() || 'U'
  )
}

function relativeTime(iso: string, t: TFunction): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return t('team.relative.justNow')
  if (m < 60) return t('team.relative.minutesAgo', { count: m })
  const h = Math.round(m / 60)
  if (h < 24) return t('team.relative.hoursAgo', { count: h })
  const d = Math.round(h / 24)
  if (d < 30) return t('team.relative.daysAgo', { count: d })
  return new Date(iso).toLocaleDateString()
}

function userError(err: unknown, t: TFunction): string {
  if (err instanceof TeamHttpError) {
    if (err.status === 409) return t('team.toast.loginOrEmailInUse')
    if (err.status === 400) return err.message || t('team.toast.invalidRequest')
    if (err.status === 403) return t('team.toast.noPermission')
    if (err.status === 404) return t('team.toast.userNotFound')
  }
  return err instanceof Error ? err.message : t('team.toast.operationFailed')
}
