import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Camera, KeyRound, Laptop, Shield, Smartphone, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useAuth } from '@/features/auth'
import { authHttpService, AuthHttpError } from '@/features/auth/services/auth-http.service'
import type { Session } from '@/features/auth/types/auth.types'

const AVATAR_MAX_BYTES = 2 * 1024 * 1024 // 2 MB
const AVATAR_ACCEPT = 'image/png,image/jpeg,image/webp,image/gif'

export function ProfilePage() {
  const { user, updateMe, changePassword, uploadAvatar, removeAvatar } = useAuth()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadingAvatar, setUploadingAvatar] = useState(false)

  const [firstName, setFirstName] = useState(user?.first_name ?? '')
  const [lastName, setLastName] = useState(user?.last_name ?? '')
  const [email, setEmail] = useState(user?.email ?? '')
  const [savingProfile, setSavingProfile] = useState(false)

  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [pwSubmitting, setPwSubmitting] = useState(false)

  const [sessions, setSessions] = useState<Session[] | null>(null)
  const [sessionsLoading, setSessionsLoading] = useState(true)
  const [revokingId, setRevokingId] = useState<number | null>(null)
  const [revokingAll, setRevokingAll] = useState(false)

  // Sync form fields when the user object hydrates after mount.
  useEffect(() => {
    setFirstName(user?.first_name ?? '')
    setLastName(user?.last_name ?? '')
    setEmail(user?.email ?? '')
  }, [user?.first_name, user?.last_name, user?.email])

  const loadSessions = useCallback(async () => {
    setSessionsLoading(true)
    try {
      const list = await authHttpService.listSessions()
      setSessions(list)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not load sessions')
      setSessions([])
    } finally {
      setSessionsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const fullName =
    [user?.first_name, user?.last_name].filter(Boolean).join(' ') || user?.login || 'User'
  const initials = fullName
    .split(' ')
    .map((p) => p[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase()

  const profileDirty = useMemo(() => {
    return (
      firstName !== (user?.first_name ?? '') ||
      lastName !== (user?.last_name ?? '') ||
      email !== (user?.email ?? '')
    )
  }, [firstName, lastName, email, user])

  const handleSavePersonal = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!profileDirty) return
    setSavingProfile(true)
    try {
      await updateMe({
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        email: email.trim(),
      })
      toast.success('Profile updated')
    } catch (err) {
      const msg =
        err instanceof AuthHttpError && err.status === 409
          ? 'That email is already in use'
          : err instanceof Error
            ? err.message
            : 'Could not update profile'
      toast.error(msg)
    } finally {
      setSavingProfile(false)
    }
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match')
      return
    }
    if (newPassword.length < 8) {
      toast.error('Password must be at least 8 characters')
      return
    }
    setPwSubmitting(true)
    try {
      await changePassword({ current_password: currentPassword, new_password: newPassword })
      toast.success('Password changed — other sessions were signed out')
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      await loadSessions()
    } catch (err) {
      const msg =
        err instanceof AuthHttpError && err.status === 401
          ? 'Current password is incorrect'
          : err instanceof Error
            ? err.message
            : 'Could not change password'
      toast.error(msg)
    } finally {
      setPwSubmitting(false)
    }
  }

  const handleRevokeSession = async (id: number) => {
    setRevokingId(id)
    try {
      await authHttpService.revokeSession(id)
      toast.success('Session revoked')
      setSessions((s) => (s ? s.filter((x) => x.id !== id) : s))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not revoke session')
    } finally {
      setRevokingId(null)
    }
  }

  const handleAvatarFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // allow re-selecting same file later
    if (!file) return
    if (!AVATAR_ACCEPT.split(',').includes(file.type)) {
      toast.error('Image type not allowed (use PNG, JPEG, WEBP or GIF)')
      return
    }
    if (file.size > AVATAR_MAX_BYTES) {
      toast.error('Image is too large (max 2MB)')
      return
    }
    setUploadingAvatar(true)
    try {
      await uploadAvatar(file)
      toast.success('Avatar updated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not upload avatar')
    } finally {
      setUploadingAvatar(false)
    }
  }

  const handleRemoveAvatar = async () => {
    setUploadingAvatar(true)
    try {
      await removeAvatar()
      toast.success('Avatar removed')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not remove avatar')
    } finally {
      setUploadingAvatar(false)
    }
  }

  const handleRevokeOthers = async () => {
    setRevokingAll(true)
    try {
      await authHttpService.revokeOtherSessions()
      toast.success('Other sessions signed out')
      await loadSessions()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not revoke sessions')
    } finally {
      setRevokingAll(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-4xl px-6 py-10">
      {/* Hero */}
      <header className="flex flex-col items-center gap-4 sm:flex-row sm:items-center sm:gap-6">
        <div className="relative">
          {user?.image_url ? (
            <img
              src={user.image_url}
              alt={fullName}
              className="h-20 w-20 rounded-full object-cover ring-1 ring-border"
            />
          ) : (
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-primary text-2xl font-semibold text-primary-foreground">
              {initials}
            </div>
          )}
          <input
            ref={fileInputRef}
            type="file"
            accept={AVATAR_ACCEPT}
            onChange={handleAvatarFileChange}
            className="hidden"
          />
          <button
            className="absolute bottom-0 right-0 flex h-7 w-7 items-center justify-center rounded-full border border-border bg-card text-muted-foreground shadow-sm hover:bg-muted hover:text-foreground disabled:opacity-50"
            aria-label="Change avatar"
            disabled={uploadingAvatar}
            onClick={() => fileInputRef.current?.click()}
          >
            <Camera size={14} strokeWidth={1.75} />
          </button>
          {user?.image_url && !uploadingAvatar && (
            <button
              className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full border border-border bg-card text-muted-foreground shadow-sm hover:bg-destructive hover:text-destructive-foreground hover:border-destructive"
              aria-label="Remove avatar"
              onClick={handleRemoveAvatar}
            >
              <X size={10} strokeWidth={2} />
            </button>
          )}
        </div>
        <div className="flex-1 text-center sm:text-left">
          <h1 className="text-xl font-semibold">{fullName}</h1>
          <p className="text-sm text-muted-foreground">{user?.email}</p>
          <div className="mt-1.5 flex items-center justify-center gap-1.5 sm:justify-start">
            <span className="rounded-md bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary ring-1 ring-inset ring-primary/20">
              Administrator
            </span>
            <span className="font-mono text-xs text-muted-foreground">@{user?.login}</span>
          </div>
        </div>
      </header>

      {/* Personal info */}
      <Section title="Personal information" description="Update your name and contact info.">
        <form onSubmit={handleSavePersonal} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="First name">
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
            <Field label="Last name">
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </Field>
          </div>
          <Field label="Email">
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <Field label="Username" hint="Username cannot be changed.">
            <Input value={user?.login ?? ''} disabled className="font-mono" />
          </Field>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              disabled={!profileDirty || savingProfile}
              onClick={() => {
                setFirstName(user?.first_name ?? '')
                setLastName(user?.last_name ?? '')
                setEmail(user?.email ?? '')
              }}
            >
              Reset
            </Button>
            <Button type="submit" disabled={!profileDirty || savingProfile}>
              {savingProfile ? 'Saving…' : 'Save changes'}
            </Button>
          </div>
        </form>
      </Section>

      {/* Password */}
      <Section title="Password" description="Use a strong password you don't reuse anywhere else." icon={KeyRound}>
        <form onSubmit={handleChangePassword} className="space-y-4">
          <Field label="Current password">
            <Input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
            />
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="New password" hint="At least 8 characters.">
              <Input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                autoComplete="new-password"
              />
            </Field>
            <Field label="Confirm new password">
              <Input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                autoComplete="new-password"
              />
            </Field>
          </div>
          <div className="flex justify-end pt-2">
            <Button type="submit" disabled={pwSubmitting}>
              {pwSubmitting ? 'Updating…' : 'Update password'}
            </Button>
          </div>
        </form>
      </Section>

      {/* Two-factor */}
      <Section title="Two-factor authentication" description="Add an extra layer of security to your account." icon={Shield}>
        <div className="flex items-center justify-between rounded-md border border-border bg-muted/40 p-4">
          <div>
            <div className="text-sm font-medium">Authenticator app</div>
            <div className="mt-0.5 text-xs text-muted-foreground">
              Status:{' '}
              <span className="text-amber-600 dark:text-amber-300">Not configured</span>
            </div>
          </div>
          <Button variant="outline" onClick={() => toast.info('2FA enrollment not wired yet')}>
            Enable
          </Button>
        </div>
      </Section>

      {/* Sessions */}
      <Section title="Active sessions" description="Devices currently signed in to your account.">
        {sessionsLoading ? (
          <div className="rounded-md border border-border p-6 text-center text-sm text-muted-foreground">
            Loading sessions…
          </div>
        ) : !sessions || sessions.length === 0 ? (
          <div className="rounded-md border border-border p-6 text-center text-sm text-muted-foreground">
            No active sessions.
          </div>
        ) : (
          <>
            <ul className="divide-y divide-border rounded-md border border-border">
              {sessions.map((s) => (
                <SessionItem
                  key={s.id}
                  session={s}
                  revoking={revokingId === s.id}
                  onRevoke={() => handleRevokeSession(s.id)}
                />
              ))}
            </ul>
            {sessions.some((s) => !s.current) && (
              <div className="mt-3 flex justify-end">
                <Button variant="outline" disabled={revokingAll} onClick={handleRevokeOthers}>
                  {revokingAll ? 'Signing out…' : 'Sign out of all other sessions'}
                </Button>
              </div>
            )}
          </>
        )}
      </Section>
    </div>
  )
}

function SessionItem({
  session,
  revoking,
  onRevoke,
}: {
  session: Session
  revoking: boolean
  onRevoke: () => void
}) {
  const ua = parseUserAgent(session.user_agent)
  return (
    <li className="flex items-center justify-between gap-3 px-4 py-3">
      <div className="flex items-center gap-3">
        <div className="flex h-9 w-9 items-center justify-center rounded-md bg-muted text-muted-foreground">
          {ua.kind === 'desktop' ? (
            <Laptop size={16} strokeWidth={1.75} />
          ) : (
            <Smartphone size={16} strokeWidth={1.75} />
          )}
        </div>
        <div>
          <div className="flex items-center gap-2 text-sm font-medium">
            {ua.label}
            {session.current && (
              <span className="rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600 dark:text-emerald-300">
                This device
              </span>
            )}
          </div>
          <div className="text-xs text-muted-foreground">
            {session.ip ? <span className="font-mono">{session.ip}</span> : 'Unknown IP'}
            {' · '}
            Signed in {formatRelative(session.created_at)}
          </div>
        </div>
      </div>
      {!session.current && (
        <Button variant="outline" size="sm" disabled={revoking} onClick={onRevoke}>
          {revoking ? 'Revoking…' : 'Revoke'}
        </Button>
      )}
    </li>
  )
}

function parseUserAgent(ua: string | undefined): { kind: 'desktop' | 'mobile'; label: string } {
  if (!ua) return { kind: 'desktop', label: 'Unknown device' }
  const isMobile = /Android|iPhone|iPad|iPod|Mobile/i.test(ua)
  const browser = ua.match(/Firefox|Edg|Chrome|Safari/i)?.[0] ?? 'Browser'
  const os =
    /Mac OS X|macOS/i.test(ua) ? 'macOS'
      : /Windows/i.test(ua) ? 'Windows'
      : /Android/i.test(ua) ? 'Android'
      : /iPhone|iPad|iOS/i.test(ua) ? 'iOS'
      : /Linux/i.test(ua) ? 'Linux'
      : 'Unknown'
  return { kind: isMobile ? 'mobile' : 'desktop', label: `${browser === 'Edg' ? 'Edge' : browser} on ${os}` }
}

function formatRelative(iso: string): string {
  const then = new Date(iso).getTime()
  const diff = Date.now() - then
  if (diff < 60_000) return 'just now'
  const m = Math.floor(diff / 60_000)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  if (d < 30) return `${d}d ago`
  return new Date(iso).toLocaleDateString()
}

interface SectionProps {
  title: string
  description?: string
  icon?: React.ComponentType<{ size?: number; strokeWidth?: number; className?: string }>
  children: React.ReactNode
}

function Section({ title, description, icon: Icon, children }: SectionProps) {
  return (
    <section className="mt-8 rounded-xl border border-border bg-card p-6">
      <header className="mb-4">
        <h2 className="flex items-center gap-2 text-base font-semibold">
          {Icon && <Icon size={16} strokeWidth={1.75} />}
          {title}
        </h2>
        {description && <p className="mt-1 text-xs text-muted-foreground">{description}</p>}
      </header>
      {children}
    </section>
  )
}

function Field({
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
