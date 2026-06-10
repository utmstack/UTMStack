import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Mail, ShieldCheck, ShieldOff, Smartphone } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { PasswordInput } from '@/shared/components/ui/password-input'
import { authHttpService, AuthHttpError } from '@/features/auth/services/auth-http.service'
import type { TfaInitResponse, TfaMethod, User } from '@/features/auth/types/auth.types'

type Mode = 'idle' | 'enroll' | 'disable'
type EnrollStep = 'choose' | 'verify'

/**
 * Self-service 2FA: shows the current status and drives the full enrollment
 * (choose method → init → verify code → complete) and teardown (re-auth with
 * password) flows. After any change it calls onChanged() to refresh the user.
 */
export function TwoFactorBody({ user, onChanged }: { user: User; onChanged: () => Promise<void> }) {
  const { t } = useTranslation()
  const enabled = !!user.tfa_enabled
  const [mode, setMode] = useState<Mode>('idle')
  const [busy, setBusy] = useState(false)

  // Enrollment state
  const [method, setMethod] = useState<TfaMethod>('TOTP')
  const [step, setStep] = useState<EnrollStep>('choose')
  const [init, setInit] = useState<TfaInitResponse | null>(null)
  const [code, setCode] = useState('')

  // Disable state
  const [password, setPassword] = useState('')

  const methodLabel = (m: TfaMethod) =>
    m === 'TOTP' ? t('profile.tfa.methodTotp') : t('profile.tfa.methodEmail')

  const enrollError = (err: unknown): string => {
    if (err instanceof AuthHttpError) {
      if (err.status === 401 || err.status === 410) return t('profile.toast.tfaInvalidCode')
      if (err.status === 409) return t('profile.toast.tfaAlreadyEnabled')
      if (err.status === 429) return t('profile.toast.tfaTooManyRequests')
      if (err.status === 400) return err.message || t('profile.toast.tfaVerifyFailed')
    }
    return t('profile.toast.genericError')
  }

  const resetEnroll = () => {
    setMode('idle')
    setStep('choose')
    setInit(null)
    setCode('')
    setMethod('TOTP')
  }

  const handleStartInit = async () => {
    setBusy(true)
    try {
      const resp = await authHttpService.tfaInit({ method })
      setInit(resp)
      setStep('verify')
      if (method === 'EMAIL') toast.success(t('profile.toast.tfaEmailCodeSent'))
    } catch (err) {
      toast.error(enrollError(err))
    } finally {
      setBusy(false)
    }
  }

  const handleResend = async () => {
    setBusy(true)
    try {
      const resp = await authHttpService.tfaRefreshChallenge('EMAIL')
      if (resp.email_sent) toast.success(t('profile.toast.tfaNewCodeSent'))
    } catch (err) {
      if (err instanceof AuthHttpError && err.status === 429) {
        toast.error(t('profile.toast.tfaWaitResend'))
      } else {
        toast.error(t('profile.toast.tfaResendFailed'))
      }
    } finally {
      setBusy(false)
    }
  }

  const handleVerifyComplete = async (e: React.FormEvent) => {
    e.preventDefault()
    if (code.length !== 6) {
      toast.error(t('profile.toast.tfaEnterCode'))
      return
    }
    setBusy(true)
    try {
      await authHttpService.tfaVerify({ method, code })
      await authHttpService.tfaComplete({ method })
      await onChanged()
      toast.success(t('profile.toast.tfaEnabled'))
      resetEnroll()
    } catch (err) {
      toast.error(enrollError(err))
    } finally {
      setBusy(false)
    }
  }

  const handleDisable = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!password) {
      toast.error(t('profile.toast.tfaEnterPassword'))
      return
    }
    setBusy(true)
    try {
      await authHttpService.tfaDisable({ password })
      await onChanged()
      toast.success(t('profile.toast.tfaDisabled'))
      setPassword('')
      setMode('idle')
    } catch (err) {
      const msg =
        err instanceof AuthHttpError && err.status === 401
          ? t('profile.toast.currentPasswordWrong')
          : t('profile.toast.tfaDisableFailed')
      toast.error(msg)
    } finally {
      setBusy(false)
    }
  }

  /* ---- Enabled, idle: show status + disable entry ---- */
  if (enabled && mode === 'idle') {
    return (
      <div className="flex items-center justify-between rounded-md border border-border bg-muted/40 p-4">
        <div className="flex items-center gap-3">
          <ShieldCheck size={18} className="text-emerald-500" />
          <div>
            <div className="text-sm font-medium">{t('profile.tfa.onTitle')}</div>
            <div className="mt-0.5 text-xs text-muted-foreground">
              {t('profile.tfa.method', {
                method: user.tfa_method ? methodLabel(user.tfa_method) : t('profile.tfa.configured'),
              })}
            </div>
          </div>
        </div>
        <Button variant="outline" onClick={() => setMode('disable')}>
          {t('profile.tfa.disable')}
        </Button>
      </div>
    )
  }

  /* ---- Disable: confirm with password ---- */
  if (mode === 'disable') {
    return (
      <form onSubmit={handleDisable} className="rounded-md border border-border bg-muted/40 p-4">
        <div className="flex items-center gap-2 text-sm font-medium">
          <ShieldOff size={16} className="text-amber-500" />
          {t('profile.tfa.disableTitle')}
        </div>
        <p className="mt-1 text-xs text-muted-foreground">{t('profile.tfa.disableDescription')}</p>
        <div className="mt-3 max-w-xs">
          <PasswordInput
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder={t('profile.tfa.currentPasswordPlaceholder')}
            autoComplete="current-password"
          />
        </div>
        <div className="mt-3 flex gap-2">
          <Button type="submit" variant="destructive" disabled={busy}>
            {busy ? t('profile.tfa.disabling') : t('profile.tfa.disableConfirm')}
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={busy}
            onClick={() => {
              setPassword('')
              setMode('idle')
            }}
          >
            {t('profile.actions.cancel')}
          </Button>
        </div>
      </form>
    )
  }

  /* ---- Disabled, idle: show status + enable entry ---- */
  if (!enabled && mode === 'idle') {
    return (
      <div className="flex items-center justify-between rounded-md border border-border bg-muted/40 p-4">
        <div>
          <div className="text-sm font-medium">{t('profile.tfa.offTitle')}</div>
          <div className="mt-0.5 text-xs text-muted-foreground">
            {t('profile.tfa.status')}{' '}
            <span className="text-amber-600 dark:text-amber-300">
              {t('profile.tfa.notConfigured')}
            </span>
          </div>
        </div>
        <Button variant="outline" onClick={() => setMode('enroll')}>
          {t('profile.tfa.enable')}
        </Button>
      </div>
    )
  }

  /* ---- Enroll: choose method ---- */
  if (mode === 'enroll' && step === 'choose') {
    return (
      <div className="rounded-md border border-border bg-muted/40 p-4">
        <div className="text-sm font-medium">{t('profile.tfa.chooseMethod')}</div>
        <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
          <MethodCard
            active={method === 'TOTP'}
            icon={<Smartphone size={16} />}
            title={t('profile.tfa.methodTotp')}
            desc={t('profile.tfa.totpDesc')}
            onClick={() => setMethod('TOTP')}
          />
          <MethodCard
            active={method === 'EMAIL'}
            icon={<Mail size={16} />}
            title={t('profile.tfa.methodEmail')}
            desc={t('profile.tfa.emailDesc')}
            disabled={!user.email}
            onClick={() => setMethod('EMAIL')}
          />
        </div>
        {!user.email && (
          <p className="mt-2 text-[11px] text-muted-foreground">{t('profile.tfa.noEmailHint')}</p>
        )}
        <div className="mt-4 flex gap-2">
          <Button onClick={handleStartInit} disabled={busy}>
            {busy ? t('profile.tfa.starting') : t('profile.tfa.continue')}
          </Button>
          <Button type="button" variant="outline" disabled={busy} onClick={resetEnroll}>
            {t('profile.actions.cancel')}
          </Button>
        </div>
      </div>
    )
  }

  /* ---- Enroll: verify code (after init) ---- */
  if (mode === 'enroll' && step === 'verify') {
    const secret = otpSecret(init?.otp_auth_url)
    return (
      <form onSubmit={handleVerifyComplete} className="rounded-md border border-border bg-muted/40 p-4">
        {method === 'TOTP' ? (
          <div className="flex flex-col items-start gap-4 sm:flex-row">
            {init?.qr_data_url && (
              <img
                src={init.qr_data_url}
                alt="2FA QR code"
                className="h-40 w-40 shrink-0 rounded-md border border-border bg-white p-1"
              />
            )}
            <div className="min-w-0">
              <div className="text-sm font-medium">{t('profile.tfa.scanTitle')}</div>
              <p className="mt-1 text-xs text-muted-foreground">{t('profile.tfa.scanDescription')}</p>
              {secret && (
                <p className="mt-2 text-[11px] text-muted-foreground">
                  {t('profile.tfa.manualKey')}{' '}
                  <code className="break-all font-mono text-foreground">{secret}</code>
                </p>
              )}
            </div>
          </div>
        ) : (
          <div>
            <div className="text-sm font-medium">{t('profile.tfa.checkEmailTitle')}</div>
            <p className="mt-1 text-xs text-muted-foreground">
              {t('profile.tfa.checkEmailDescription', { email: user.email })}{' '}
              <button
                type="button"
                onClick={handleResend}
                disabled={busy}
                className="text-primary hover:underline disabled:opacity-50"
              >
                {t('profile.tfa.resend')}
              </button>
            </p>
          </div>
        )}

        <div className="mt-4 max-w-[200px]">
          <Input
            inputMode="numeric"
            autoComplete="one-time-code"
            maxLength={6}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="000000"
            className="text-center font-mono text-lg tracking-[0.3em]"
          />
        </div>
        <div className="mt-4 flex gap-2">
          <Button type="submit" disabled={busy || code.length !== 6}>
            {busy ? t('profile.tfa.verifying') : t('profile.tfa.verifyEnable')}
          </Button>
          <Button type="button" variant="outline" disabled={busy} onClick={resetEnroll}>
            {t('profile.actions.cancel')}
          </Button>
        </div>
      </form>
    )
  }

  return null
}

function MethodCard({
  active,
  icon,
  title,
  desc,
  disabled,
  onClick,
}: {
  active: boolean
  icon: React.ReactNode
  title: string
  desc: string
  disabled?: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className={[
        'flex flex-col items-start gap-1 rounded-md border p-3 text-left transition-colors',
        active ? 'border-primary bg-primary/5' : 'border-border bg-card hover:border-primary/40',
        disabled ? 'cursor-not-allowed opacity-50' : '',
      ].join(' ')}
    >
      <span className="flex items-center gap-2 text-sm font-medium">
        {icon}
        {title}
      </span>
      <span className="text-[11px] text-muted-foreground">{desc}</span>
    </button>
  )
}

/** Pull the shared secret out of an otpauth:// URL for manual entry. */
function otpSecret(otpAuthUrl: string | undefined): string | null {
  if (!otpAuthUrl) return null
  try {
    return new URL(otpAuthUrl).searchParams.get('secret')
  } catch {
    return null
  }
}
