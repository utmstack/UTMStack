import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Camera, KeyRound, Laptop, Shield, Smartphone, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { PasswordInput } from '@/shared/components/ui/password-input'
import { IS_FEDERATION } from '@/shared/config/mode'
import { useAuth } from '@/features/auth'
import { authHttpService, AuthHttpError } from '@/features/auth/services/auth-http.service'
import type { Session } from '@/features/auth/types/auth.types'
import { federationAuthService } from '@/features/federation/services/federation-auth.service'
import { SUPPORTED_LANGUAGES } from '@/shared/i18n'
import { TwoFactorBody } from '../components/TwoFactorBody'
import { ApiKeysSection } from '@/features/settings/pages/ApiKeysPage'

const AVATAR_MAX_BYTES = 2 * 1024 * 1024 // 2 MB
const AVATAR_ACCEPT = 'image/png,image/jpeg,image/webp,image/gif'

// Sessions live in the FS in federation mode and in the instance otherwise; both
// expose the same shape so the section below renders identically either way.
const sessionsService = IS_FEDERATION
  ? {
      list: () => federationAuthService.listSessions(),
      revoke: (id: number) => federationAuthService.revokeSession(id),
      revokeOthers: () => federationAuthService.revokeOtherSessions(),
    }
  : {
      list: () => authHttpService.listSessions(),
      revoke: (id: number) => authHttpService.revokeSession(id),
      revokeOthers: () => authHttpService.revokeOtherSessions(),
    }

export function ProfilePage() {
  const { t } = useTranslation()
  const { user, updateMe, changePassword, uploadAvatar, removeAvatar, refreshUser } = useAuth()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploadingAvatar, setUploadingAvatar] = useState(false)

  const [firstName, setFirstName] = useState(user?.first_name ?? '')
  const [lastName, setLastName] = useState(user?.last_name ?? '')
  const [email, setEmail] = useState(user?.email ?? '')
  const [langKey, setLangKey] = useState(user?.lang_key ?? '')
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
    setLangKey(user?.lang_key ?? '')
  }, [user?.first_name, user?.last_name, user?.email, user?.lang_key])

  const loadSessions = useCallback(async () => {
    setSessionsLoading(true)
    try {
      const list = await sessionsService.list()
      setSessions(list)
    } catch {
      toast.error(t('profile.toast.sessionsLoadFailed'))
      setSessions([])
    } finally {
      setSessionsLoading(false)
    }
  }, [t])

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
      email !== (user?.email ?? '') ||
      langKey !== (user?.lang_key ?? '')
    )
  }, [firstName, lastName, email, langKey, user])

  const handleSavePersonal = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!profileDirty) return
    setSavingProfile(true)
    try {
      await updateMe({
        first_name: firstName.trim(),
        last_name: lastName.trim(),
        email: email.trim(),
        lang_key: langKey || undefined,
      })
      toast.success(t('profile.toast.profileUpdated'))
    } catch (err) {
      const msg =
        err instanceof AuthHttpError && err.status === 409
          ? t('profile.toast.emailInUse')
          : t('profile.toast.profileUpdateFailed')
      toast.error(msg)
    } finally {
      setSavingProfile(false)
    }
  }

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (newPassword !== confirmPassword) {
      toast.error(t('profile.toast.passwordsNoMatch'))
      return
    }
    if (newPassword.length < 8) {
      toast.error(t('profile.toast.passwordTooShort'))
      return
    }
    setPwSubmitting(true)
    try {
      await changePassword({ current_password: currentPassword, new_password: newPassword })
      toast.success(t('profile.toast.passwordChanged'))
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      await loadSessions()
    } catch (err) {
      const msg =
        err instanceof AuthHttpError && err.status === 401
          ? t('profile.toast.currentPasswordWrong')
          : t('profile.toast.passwordChangeFailed')
      toast.error(msg)
    } finally {
      setPwSubmitting(false)
    }
  }

  const handleRevokeSession = async (id: number) => {
    setRevokingId(id)
    try {
      await sessionsService.revoke(id)
      toast.success(t('profile.toast.sessionRevoked'))
      setSessions((s) => (s ? s.filter((x) => x.id !== id) : s))
    } catch {
      toast.error(t('profile.toast.sessionRevokeFailed'))
    } finally {
      setRevokingId(null)
    }
  }

  const handleAvatarFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    e.target.value = '' // allow re-selecting same file later
    if (!file) return
    if (!AVATAR_ACCEPT.split(',').includes(file.type)) {
      toast.error(t('profile.toast.avatarTypeNotAllowed'))
      return
    }
    if (file.size > AVATAR_MAX_BYTES) {
      toast.error(t('profile.toast.avatarTooLarge'))
      return
    }
    setUploadingAvatar(true)
    try {
      await uploadAvatar(file)
      toast.success(t('profile.toast.avatarUpdated'))
    } catch {
      toast.error(t('profile.toast.avatarUploadFailed'))
    } finally {
      setUploadingAvatar(false)
    }
  }

  const handleRemoveAvatar = async () => {
    setUploadingAvatar(true)
    try {
      await removeAvatar()
      toast.success(t('profile.toast.avatarRemoved'))
    } catch {
      toast.error(t('profile.toast.avatarRemoveFailed'))
    } finally {
      setUploadingAvatar(false)
    }
  }

  const handleRevokeOthers = async () => {
    setRevokingAll(true)
    try {
      await sessionsService.revokeOthers()
      toast.success(t('profile.toast.othersSignedOut'))
      await loadSessions()
    } catch {
      toast.error(t('profile.toast.othersRevokeFailed'))
    } finally {
      setRevokingAll(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pt-5 pb-10">
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
            aria-label={t('profile.hero.changeAvatar')}
            disabled={uploadingAvatar}
            onClick={() => fileInputRef.current?.click()}
          >
            <Camera size={14} strokeWidth={1.75} />
          </button>
          {user?.image_url && !uploadingAvatar && (
            <button
              className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full border border-border bg-card text-muted-foreground shadow-sm hover:bg-destructive hover:text-destructive-foreground hover:border-destructive"
              aria-label={t('profile.hero.removeAvatar')}
              onClick={handleRemoveAvatar}
            >
              <X size={10} strokeWidth={2} />
            </button>
          )}
        </div>
        <div className="flex-1 text-center sm:text-left">
          <h1 className="text-base font-semibold">{fullName}</h1>
          <div className="mt-1.5 flex items-center justify-center gap-1.5 sm:justify-start">
            <span className="rounded-md bg-primary/15 px-2 py-0.5 text-[11px] font-medium text-primary ring-1 ring-inset ring-primary/20">
              {t('profile.hero.administrator')}
            </span>
            <span className="font-mono text-xs text-muted-foreground">@{user?.login}</span>
          </div>
        </div>
      </header>

      {/* Personal info */}
      <Section title={t('profile.personal.title')} description={t('profile.personal.description')}>
        <form onSubmit={handleSavePersonal} className="space-y-4">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label={t('profile.personal.firstName')}>
              <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} />
            </Field>
            <Field label={t('profile.personal.lastName')}>
              <Input value={lastName} onChange={(e) => setLastName(e.target.value)} />
            </Field>
          </div>
          <Field label={t('profile.personal.email')}>
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label={t('profile.personal.language')} hint={t('profile.personal.languageHint')}>
              <select
                value={langKey}
                onChange={(e) => setLangKey(e.target.value)}
                className="flex h-10 w-full cursor-pointer rounded-md border border-input bg-background/40 px-3 py-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="">{t('profile.personal.systemDefault')}</option>
                {SUPPORTED_LANGUAGES.map((l) => (
                  <option key={l.code} value={l.code}>
                    {l.label}
                  </option>
                ))}
              </select>
            </Field>
            <Field label={t('profile.personal.username')} hint={t('profile.personal.usernameHint')}>
              <Input value={user?.login ?? ''} disabled className="font-mono" />
            </Field>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              disabled={!profileDirty || savingProfile}
              onClick={() => {
                setFirstName(user?.first_name ?? '')
                setLastName(user?.last_name ?? '')
                setEmail(user?.email ?? '')
                setLangKey(user?.lang_key ?? '')
              }}
            >
              {t('profile.personal.reset')}
            </Button>
            <Button type="submit" disabled={!profileDirty || savingProfile}>
              {savingProfile ? t('profile.personal.saving') : t('profile.personal.save')}
            </Button>
          </div>
        </form>
      </Section>

      {/* Password */}
      <Section
        title={t('profile.password.title')}
        description={t('profile.password.description')}
        icon={KeyRound}
      >
        <form onSubmit={handleChangePassword} className="space-y-4">
          <Field label={t('profile.password.current')}>
            <PasswordInput
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
            />
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label={t('profile.password.new')} hint={t('profile.password.newHint')}>
              <PasswordInput
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                autoComplete="new-password"
              />
            </Field>
            <Field label={t('profile.password.confirm')}>
              <PasswordInput
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                autoComplete="new-password"
              />
            </Field>
          </div>
          <div className="flex justify-end pt-2">
            <Button type="submit" disabled={pwSubmitting}>
              {pwSubmitting ? t('profile.password.updating') : t('profile.password.update')}
            </Button>
          </div>
        </form>
      </Section>

      {/* Two-factor. Federation is TOTP-only (no mailer), so EMAIL is hidden there. */}
      <Section
        title={t('profile.tfa.title')}
        description={t('profile.tfa.description')}
        icon={Shield}
      >
        {user ? (
          <TwoFactorBody user={user} onChanged={refreshUser} allowEmail={!IS_FEDERATION} />
        ) : (
          <div className="rounded-md border border-border p-6 text-center text-sm text-muted-foreground">
            {t('profile.tfa.loading')}
          </div>
        )}
      </Section>

      {/* API keys (per-user → embedded here, not in system settings). Not in federation. */}
      {!IS_FEDERATION && <ApiKeysSection />}

      {/* Active sessions (FS sessions in federation mode, instance sessions otherwise) */}
      <Section title={t('profile.sessions.title')} description={t('profile.sessions.description')}>
        {sessionsLoading ? (
          <div className="rounded-md border border-border p-6 text-center text-sm text-muted-foreground">
            {t('profile.sessions.loading')}
          </div>
        ) : !sessions || sessions.length === 0 ? (
          <div className="rounded-md border border-border p-6 text-center text-sm text-muted-foreground">
            {t('profile.sessions.empty')}
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
                  {revokingAll
                    ? t('profile.sessions.signingOut')
                    : t('profile.sessions.signOutOthers')}
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
  const { t, i18n } = useTranslation()
  const ua = parseUserAgent(session.user_agent)
  const deviceLabel =
    ua.browser && ua.os
      ? t('profile.sessions.deviceLabel', { browser: ua.browser, os: ua.os })
      : t('profile.sessions.unknownDevice')
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
            {deviceLabel}
            {session.current && (
              <span className="rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[10px] font-medium text-emerald-600 dark:text-emerald-300">
                {t('profile.sessions.thisDevice')}
              </span>
            )}
          </div>
          <div className="text-xs text-muted-foreground">
            {session.ip ? (
              <span className="font-mono">{session.ip}</span>
            ) : (
              t('profile.sessions.unknownIp')
            )}
            {' · '}
            {t('profile.sessions.signedIn', {
              time: formatRelative(session.created_at, i18n.language),
            })}
          </div>
        </div>
      </div>
      {!session.current && (
        <Button variant="outline" size="sm" disabled={revoking} onClick={onRevoke}>
          {revoking ? t('profile.sessions.revoking') : t('profile.sessions.revoke')}
        </Button>
      )}
    </li>
  )
}

function parseUserAgent(ua: string | undefined): {
  kind: 'desktop' | 'mobile'
  browser: string | null
  os: string | null
} {
  if (!ua) return { kind: 'desktop', browser: null, os: null }
  const isMobile = /Android|iPhone|iPad|iPod|Mobile/i.test(ua)
  const raw = ua.match(/Firefox|Edg|Chrome|Safari/i)?.[0]
  const browser = raw ? (raw === 'Edg' ? 'Edge' : raw) : 'Browser'
  const os = /Mac OS X|macOS/i.test(ua)
    ? 'macOS'
    : /Windows/i.test(ua)
      ? 'Windows'
      : /Android/i.test(ua)
        ? 'Android'
        : /iPhone|iPad|iOS/i.test(ua)
          ? 'iOS'
          : /Linux/i.test(ua)
            ? 'Linux'
            : 'Unknown'
  return { kind: isMobile ? 'mobile' : 'desktop', browser, os }
}

/** Localized relative time ("5 minutes ago" / "hace 5 minutos") via Intl. */
function formatRelative(iso: string, locale: string): string {
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const diffSec = Math.round((then - Date.now()) / 1000)
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' })
  if (Math.abs(diffSec) < 60) return rtf.format(diffSec, 'second')
  const min = Math.round(diffSec / 60)
  if (Math.abs(min) < 60) return rtf.format(min, 'minute')
  const hr = Math.round(diffSec / 3600)
  if (Math.abs(hr) < 24) return rtf.format(hr, 'hour')
  const day = Math.round(diffSec / 86400)
  if (Math.abs(day) < 30) return rtf.format(day, 'day')
  return new Date(iso).toLocaleDateString(locale)
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

