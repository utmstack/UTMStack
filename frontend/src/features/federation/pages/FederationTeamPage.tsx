import { useCallback, useEffect, useState } from 'react'
import { Loader2, Mail, RotateCcw, ShieldOff, UserPlus, Users, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useAuth } from '@/features/auth'
import {
  FederationTeamError,
  federationTeamService,
} from '../services/federation-team.service'
import type { FederationTeamUser } from '../types'

/**
 * Flat team management for the FS: every member is an admin (FS_ADMIN), so there
 * are no roles — just invite / activate / deactivate / reset 2FA. Invitations go
 * out by email (the SMTP config must be set first).
 */
export function FederationTeamPage() {
  const { user: me } = useAuth()
  const [users, setUsers] = useState<FederationTeamUser[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [inviteOpen, setInviteOpen] = useState(false)
  const [busyId, setBusyId] = useState<number | null>(null)

  useEffect(() => {
    const id = setTimeout(() => setDebounced(search.trim()), 300)
    return () => clearTimeout(id)
  }, [search])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await federationTeamService.list({ page: 1, page_size: 100, search: debounced })
      setUsers(res.data)
    } catch {
      toast.error('Could not load team members')
    } finally {
      setLoading(false)
    }
  }, [debounced])

  useEffect(() => {
    void load()
  }, [load])

  const act = async (id: number, fn: () => Promise<unknown>, ok: string) => {
    setBusyId(id)
    try {
      await fn()
      toast.success(ok)
      await load()
    } catch (err) {
      toast.error(err instanceof FederationTeamError ? err.message : 'Action failed')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div className="w-full px-6 pb-10 pt-5">
      <header className="flex items-center justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-base font-semibold">
            <Users size={16} strokeWidth={1.75} />
            Team
          </h1>
          <p className="mt-1 text-xs text-muted-foreground">
            Federation administrators. Everyone here can manage instances and settings.
          </p>
        </div>
        <Button size="sm" onClick={() => setInviteOpen(true)}>
          <UserPlus size={15} className="mr-1.5" />
          Invite member
        </Button>
      </header>

      <div className="mt-5">
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search by name, login or email…"
          className="max-w-sm"
        />
      </div>

      <div className="mt-4 overflow-hidden rounded-xl border border-border bg-card">
        {loading ? (
          <div className="flex items-center justify-center py-16 text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin" />
          </div>
        ) : users.length === 0 ? (
          <div className="py-16 text-center text-sm text-muted-foreground">No team members found.</div>
        ) : (
          <ul className="divide-y divide-border">
            {users.map((u) => (
              <MemberRow
                key={u.id}
                user={u}
                isMe={me?.id === u.id}
                busy={busyId === u.id}
                onResend={() =>
                  act(u.id, () => federationTeamService.resendInvite(u.id), 'Invitation resent')
                }
                onResetTfa={() =>
                  act(u.id, () => federationTeamService.resetTfa(u.id), '2FA reset')
                }
                onDeactivate={() =>
                  act(u.id, () => federationTeamService.deactivate(u.id), 'Member deactivated')
                }
                onReactivate={() =>
                  act(
                    u.id,
                    () => federationTeamService.update(u.id, { activated: true }),
                    'Member reactivated',
                  )
                }
              />
            ))}
          </ul>
        )}
      </div>

      {inviteOpen && (
        <InviteModal
          onClose={() => setInviteOpen(false)}
          onInvited={() => {
            setInviteOpen(false)
            void load()
          }}
        />
      )}
    </div>
  )
}

function MemberRow({
  user: u,
  isMe,
  busy,
  onResend,
  onResetTfa,
  onDeactivate,
  onReactivate,
}: {
  user: FederationTeamUser
  isMe: boolean
  busy: boolean
  onResend: () => void
  onResetTfa: () => void
  onDeactivate: () => void
  onReactivate: () => void
}) {
  const name = [u.first_name, u.last_name].filter(Boolean).join(' ') || u.login
  const initials = name
    .split(' ')
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase()

  return (
    <li className="flex items-center justify-between gap-3 px-4 py-3">
      <div className="flex min-w-0 items-center gap-3">
        {u.image_url ? (
          <img src={u.image_url} alt={name} className="h-9 w-9 rounded-full object-cover" />
        ) : (
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">
            {initials || 'U'}
          </span>
        )}
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-sm font-medium">
            <span className="truncate">{name}</span>
            {isMe && (
              <span className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                You
              </span>
            )}
          </div>
          <div className="truncate text-xs text-muted-foreground">{u.email || u.login}</div>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {u.tfa_enabled && (
          <span className="hidden rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600 sm:inline dark:text-emerald-300">
            2FA
          </span>
        )}
        <StatusBadge user={u} />
        {busy ? (
          <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        ) : (
          !isMe && (
            <div className="flex items-center gap-1">
              {u.pending && (
                <RowAction label="Resend invite" icon={Mail} onClick={onResend} />
              )}
              {u.tfa_enabled && u.activated && (
                <RowAction label="Reset 2FA" icon={ShieldOff} onClick={onResetTfa} />
              )}
              {u.activated ? (
                <Button variant="outline" size="sm" onClick={onDeactivate}>
                  Deactivate
                </Button>
              ) : (
                <RowAction label="Reactivate" icon={RotateCcw} onClick={onReactivate} />
              )}
            </div>
          )
        )}
      </div>
    </li>
  )
}

function RowAction({
  label,
  icon: Icon,
  onClick,
}: {
  label: string
  icon: typeof Mail
  onClick: () => void
}) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      onClick={onClick}
      className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
    >
      <Icon size={15} strokeWidth={1.75} />
    </button>
  )
}

function StatusBadge({ user: u }: { user: FederationTeamUser }) {
  if (u.activated) {
    return (
      <span className="rounded-md bg-emerald-500/15 px-2 py-0.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-300">
        Active
      </span>
    )
  }
  if (u.pending) {
    return (
      <span className="rounded-md bg-amber-500/15 px-2 py-0.5 text-[11px] font-medium text-amber-600 dark:text-amber-300">
        Pending invite
      </span>
    )
  }
  return (
    <span className="rounded-md bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground">
      Deactivated
    </span>
  )
}

function InviteModal({ onClose, onInvited }: { onClose: () => void; onInvited: () => void }) {
  const [login, setLogin] = useState('')
  const [email, setEmail] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!login.trim() || !email.trim()) {
      toast.error('Login and email are required')
      return
    }
    setSubmitting(true)
    try {
      await federationTeamService.create({
        login: login.trim(),
        email: email.trim(),
        first_name: firstName.trim() || undefined,
        last_name: lastName.trim() || undefined,
      })
      toast.success('Invitation sent')
      onInvited()
    } catch (err) {
      if (err instanceof FederationTeamError && err.status === 409) {
        toast.error('That login or email is already in use')
      } else {
        toast.error(err instanceof FederationTeamError ? err.message : 'Could not invite member')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onMouseDown={onClose}>
      <div
        className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-xl"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <UserPlus size={16} strokeWidth={1.75} />
            Invite member
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </div>
        <form onSubmit={submit} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="First name">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
            <Field label="Last name">
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </Field>
          </div>
          <Field label="Login">
            <Input value={login} onChange={(e) => setLogin(e.target.value)} autoComplete="off" />
          </Field>
          <Field label="Email">
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <p className="text-[11px] text-muted-foreground">
            We'll email an invitation link to set a password. Configure SMTP first if you haven't.
          </p>
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" disabled={submitting} onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Sending…' : 'Send invitation'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
    </div>
  )
}
