import { useEffect, useState } from 'react'
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom'
import { useBranding } from '@/features/branding'
import { ssoHttpService, type PublicIdentityProvider } from '../services/sso-http.service'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Building2, KeyRound, Loader2, Lock, Mail, ShieldCheck, User as UserIcon } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { PasswordInput } from '@/shared/components/ui/password-input'
import { LanguageSwitcher } from '@/shared/components/LanguageSwitcher'
import { AuthHttpError } from '../services/auth-http.service'
import { useAuth } from '../services/auth.context'
import type { TfaFactorType } from '../types/auth.types'

type Mode = 'credentials' | 'tfa' | 'forgot' | 'reset'

interface TfaChallenge {
  method: TfaFactorType
  preAuthToken: string
}

export function LoginPage() {
  const { branding } = useBranding()
  const [providers, setProviders] = useState<PublicIdentityProvider[]>([])
  // A chosen directory only narrows the bind; the form stays the same, which is
  // the whole point of a directory not being a redirect.
  const [directory, setDirectory] = useState<PublicIdentityProvider | null>(null)

  // A provider list that cannot be read is not an error worth showing: the
  // password form still works, and an install with no SSO is the common case.
  useEffect(() => {
    void ssoHttpService
      .list()
      .then(setProviders)
      .catch(() => setProviders([]))
  }, [])
  const brandActive = !!branding?.enabled
  const brandName = (brandActive && branding?.productName) || 'UTMStack'
  const brandLogo = (brandActive && branding?.logoUrl) || '/logo.svg'

  const { t } = useTranslation()
  const { login, verifyTfaCode, requestPasswordReset, finishPasswordReset, adoptSession } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()
  const from = (location.state as { from?: { pathname: string } } | null)?.from?.pathname ?? '/'

  const resetKey = searchParams.get('key')
  const ssoToken = searchParams.get('token')
  const ssoError = searchParams.get('error')

  const [mode, setMode] = useState<Mode>(resetKey ? 'reset' : 'credentials')
  const [loginValue, setLoginValue] = useState('')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [resetEmail, setResetEmail] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [challenge, setChallenge] = useState<TfaChallenge | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // The provider sent the browser back here with a token on the URL. Adopt it
  // and drop it from the address bar: a session token in the history, or in the
  // referer of the next request, outlives the reason it was put there.
  useEffect(() => {
    if (!ssoToken) return
    void adoptSession(ssoToken)
      .then(() => {
        window.history.replaceState({}, '', window.location.pathname)
        goAuthenticated()
      })
      .catch(() => setError(t('auth.errors.ssoFailed')))
  }, [ssoToken])

  useEffect(() => {
    if (ssoError) setError(t('auth.errors.ssoFailed'))
  }, [ssoError])


  // An invitation / reset link lands here with ?key=… → show the set-password form.
  useEffect(() => {
    if (resetKey) setMode('reset')
  }, [resetKey])

  const goAuthenticated = () => navigate(from, { replace: true })

  const mapError = (err: unknown, fallback: string): string => {
    if (err instanceof AuthHttpError) {
      switch (err.status) {
        case 429:
          return t('auth.errors.tooManyAttempts')
        case 403:
          return t('auth.errors.deactivated')
        case 401:
          return t('auth.errors.invalidCredentials')
        case 410:
          return t('auth.errors.codeExpired')
        default:
          return err.message || fallback
      }
    }
    return fallback
  }

  const handleCredentials = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!loginValue || !password || submitting) return
    setSubmitting(true)
    setError(null)
    try {
      const result = await login({ login: loginValue, password, provider_id: directory?.id })
      if (result.status === 'tfa_required') {
        setChallenge({ method: result.method, preAuthToken: result.preAuthToken })
        setMode('tfa')
        setCode('')
        return
      }
      goAuthenticated()
    } catch (err) {
      setError(mapError(err, t('auth.errors.loginFailed')))
    } finally {
      setSubmitting(false)
    }
  }

  const handleTfa = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!challenge || code.length !== 6 || submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await verifyTfaCode(challenge.preAuthToken, code)
      goAuthenticated()
    } catch (err) {
      if (err instanceof AuthHttpError && err.status === 401) {
        setError(t('auth.errors.invalidCode'))
      } else {
        setError(mapError(err, t('auth.errors.verifyFailed')))
      }
    } finally {
      setSubmitting(false)
    }
  }

  const handleForgot = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!resetEmail || submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await requestPasswordReset(resetEmail)
      toast.success(t('auth.forgot.sent'))
      backToCredentials()
    } catch (err) {
      // The backend returns 204 even when the email is unknown (anti-enumeration);
      // a real error here is a bad request or a server issue.
      setError(mapError(err, t('auth.errors.resetFailed')))
    } finally {
      setSubmitting(false)
    }
  }

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!resetKey || submitting) return
    if (newPassword.length < 8) {
      setError(t('auth.errors.passwordTooShort', 'Password must be at least 8 characters'))
      return
    }
    if (newPassword !== confirmPassword) {
      setError(t('auth.errors.passwordsNoMatch', 'Passwords do not match'))
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      await finishPasswordReset({ key: resetKey, new_password: newPassword })
      toast.success(t('auth.reset.done', 'Password set. You can sign in now.'))
      setNewPassword('')
      setConfirmPassword('')
      // Drop the key from the URL and return to the sign-in form.
      searchParams.delete('key')
      setSearchParams(searchParams, { replace: true })
      backToCredentials()
    } catch (err) {
      if (err instanceof AuthHttpError && (err.status === 410 || err.status === 400)) {
        setError(t('auth.reset.invalid', 'This link is invalid or has expired.'))
      } else {
        setError(mapError(err, t('auth.errors.resetFailed')))
      }
    } finally {
      setSubmitting(false)
    }
  }

  const backToCredentials = () => {
    setMode('credentials')
    setError(null)
    setCode('')
    setChallenge(null)
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-4 text-foreground">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'radial-gradient(70% 55% at 50% -10%, hsl(211 95% 60% / 0.20) 0%, transparent 65%), radial-gradient(50% 40% at 100% 110%, hsl(211 80% 45% / 0.10) 0%, transparent 60%)',
        }}
      />

      <div className="absolute right-4 top-4 z-20">
        <LanguageSwitcher align="right" />
      </div>

      <div className="relative z-10 w-full max-w-sm">
        <div className="rounded-xl border border-border/80 bg-card/60 px-6 py-8 shadow-xl shadow-black/40 backdrop-blur-md">
          <div className="mb-9 flex flex-col items-center text-center">
            <img src={brandLogo} alt={brandName} className="mb-5 h-20 w-auto" draggable={false} />
            {mode === 'credentials' && (
              <>
                <h1 className="text-[22px] font-semibold tracking-tight">{t('auth.login.title')}</h1>
                <p className="mt-1.5 text-sm text-muted-foreground">{t('auth.login.subtitle', { brand: brandName })}</p>
              </>
            )}
            {mode === 'tfa' && (
              <>
                <h1 className="text-[22px] font-semibold tracking-tight">{t('auth.tfa.title')}</h1>
                <p className="mt-1.5 text-sm text-muted-foreground">
                  {challenge?.method === 'totp'
                    ? t('auth.tfa.subtitleTotp')
                    : t('auth.tfa.subtitleEmail')}
                </p>
              </>
            )}
            {mode === 'forgot' && (
              <>
                <h1 className="text-[22px] font-semibold tracking-tight">{t('auth.forgot.title')}</h1>
                <p className="mt-1.5 text-sm text-muted-foreground">{t('auth.forgot.subtitle')}</p>
              </>
            )}
            {mode === 'reset' && (
              <>
                <h1 className="text-[22px] font-semibold tracking-tight">
                  {t('auth.reset.title', 'Set your password')}
                </h1>
                <p className="mt-1.5 text-sm text-muted-foreground">
                  {t('auth.reset.subtitle', 'Choose a password to activate your account.')}
                </p>
              </>
            )}
          </div>

          {mode === 'credentials' && (
            <form onSubmit={handleCredentials}>
              {directory && (
                <button
                  type="button"
                  onClick={() => setDirectory(null)}
                  className="mb-4 flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                >
                  <ArrowLeft size={13} />
                  {t('auth.login.useAnother')}
                </button>
              )}
              {directory && (
                <p className="mb-4 flex items-center gap-1.5 text-sm font-medium">
                  <Building2 size={14} className="text-muted-foreground" />
                  {t('auth.login.signingInTo', { name: directory.name })}
                </p>
              )}
              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label
                    htmlFor="login"
                    className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                  >
                    {t('auth.login.loginLabel')}
                  </label>
                  <div className="relative">
                    <UserIcon
                      aria-hidden
                      className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                    />
                    <Input
                      id="login"
                      type="text"
                      placeholder="you@company.com"
                      autoComplete="username"
                      required
                      autoFocus
                      value={loginValue}
                      onChange={(e) => setLoginValue(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <label
                      htmlFor="password"
                      className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                    >
                      {t('auth.login.passwordLabel')}
                    </label>
                    <button
                      type="button"
                      className="text-[11px] text-muted-foreground transition-colors hover:text-foreground"
                      onClick={() => {
                        setMode('forgot')
                        setError(null)
                        setResetEmail(loginValue.includes('@') ? loginValue : '')
                      }}
                    >
                      {t('auth.login.forgot')}
                    </button>
                  </div>
                  <div className="relative">
                    <Lock
                      aria-hidden
                      className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                    />
                    <PasswordInput
                      id="password"
                      placeholder="••••••••"
                      autoComplete="current-password"
                      required
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>
              </div>

              {error && <p className="mt-4 text-xs text-destructive">{error}</p>}

              <Button
                type="submit"
                className="mt-6 h-10 w-full"
                disabled={submitting || !loginValue || !password}
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {t('auth.login.submit')}
              </Button>

              {providers.length > 0 && !directory && (
                <>
                  <div className="mt-6 flex items-center gap-3">
                    <span className="h-px flex-1 bg-border" />
                    <span className="text-[11px] uppercase tracking-wide text-muted-foreground">
                      {t('auth.login.orContinueWith')}
                    </span>
                    <span className="h-px flex-1 bg-border" />
                  </div>
                  <div className="mt-4 space-y-2">
                    {providers.map((p) =>
                      p.loginUrl ? (
                        <a
                          key={p.id}
                          href={p.loginUrl}
                          className="flex h-10 w-full items-center justify-center gap-2 rounded-md border border-input bg-background/40 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                        >
                          <KeyRound size={14} />
                          {t('auth.login.continueWith', { name: p.name })}
                        </a>
                      ) : (
                        <button
                          key={p.id}
                          type="button"
                          onClick={() => setDirectory(p)}
                          className="flex h-10 w-full items-center justify-center gap-2 rounded-md border border-input bg-background/40 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground"
                        >
                          <Building2 size={14} />
                          {t('auth.login.continueWith', { name: p.name })}
                        </button>
                      ),
                    )}
                  </div>
                </>
              )}
            </form>
          )}

          {mode === 'tfa' && (
            <form onSubmit={handleTfa}>
              <div className="space-y-1.5">
                <label
                  htmlFor="tfa-code"
                  className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                >
                  {t('auth.tfa.codeLabel')}
                </label>
                <div className="relative">
                  <ShieldCheck
                    aria-hidden
                    className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  />
                  <Input
                    id="tfa-code"
                    type="text"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    placeholder="000000"
                    maxLength={6}
                    required
                    autoFocus
                    value={code}
                    onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    className="pl-9 tracking-[0.4em]"
                  />
                </div>
              </div>

              {error && <p className="mt-4 text-xs text-destructive">{error}</p>}

              <Button
                type="submit"
                className="mt-6 h-10 w-full"
                disabled={submitting || code.length !== 6}
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {t('auth.tfa.submit')}
              </Button>

              <button
                type="button"
                onClick={backToCredentials}
                className="mt-4 flex w-full items-center justify-center gap-1.5 text-[11px] text-muted-foreground transition-colors hover:text-foreground"
              >
                <ArrowLeft className="h-3 w-3" />
                {t('auth.tfa.back')}
              </button>
            </form>
          )}

          {mode === 'forgot' && (
            <form onSubmit={handleForgot}>
              <div className="space-y-1.5">
                <label
                  htmlFor="reset-email"
                  className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                >
                  {t('auth.forgot.emailLabel')}
                </label>
                <div className="relative">
                  <Mail
                    aria-hidden
                    className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                  />
                  <Input
                    id="reset-email"
                    type="email"
                    placeholder="you@company.com"
                    autoComplete="email"
                    required
                    autoFocus
                    value={resetEmail}
                    onChange={(e) => setResetEmail(e.target.value)}
                    className="pl-9"
                  />
                </div>
              </div>

              {error && <p className="mt-4 text-xs text-destructive">{error}</p>}

              <Button
                type="submit"
                className="mt-6 h-10 w-full"
                disabled={submitting || !resetEmail}
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {t('auth.forgot.submit')}
              </Button>

              <button
                type="button"
                onClick={backToCredentials}
                className="mt-4 flex w-full items-center justify-center gap-1.5 text-[11px] text-muted-foreground transition-colors hover:text-foreground"
              >
                <ArrowLeft className="h-3 w-3" />
                {t('auth.tfa.back')}
              </button>
            </form>
          )}

          {mode === 'reset' && (
            <form onSubmit={handleReset}>
              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label
                    htmlFor="new-password"
                    className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                  >
                    {t('auth.reset.newPassword', 'New password')}
                  </label>
                  <div className="relative">
                    <Lock
                      aria-hidden
                      className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                    />
                    <PasswordInput
                      id="new-password"
                      placeholder="••••••••"
                      autoComplete="new-password"
                      required
                      autoFocus
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>
                <div className="space-y-1.5">
                  <label
                    htmlFor="confirm-password"
                    className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                  >
                    {t('auth.reset.confirmPassword', 'Confirm password')}
                  </label>
                  <div className="relative">
                    <Lock
                      aria-hidden
                      className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                    />
                    <PasswordInput
                      id="confirm-password"
                      placeholder="••••••••"
                      autoComplete="new-password"
                      required
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>
              </div>

              {error && <p className="mt-4 text-xs text-destructive">{error}</p>}

              <Button
                type="submit"
                className="mt-6 h-10 w-full"
                disabled={submitting || !newPassword || !confirmPassword}
              >
                {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {t('auth.reset.submit', 'Set password')}
              </Button>

              <button
                type="button"
                onClick={() => {
                  searchParams.delete('key')
                  setSearchParams(searchParams, { replace: true })
                  backToCredentials()
                }}
                className="mt-4 flex w-full items-center justify-center gap-1.5 text-[11px] text-muted-foreground transition-colors hover:text-foreground"
              >
                <ArrowLeft className="h-3 w-3" />
                {t('auth.reset.backToSignIn', 'Back to sign in')}
              </button>
            </form>
          )}
        </div>

        <p className="mt-6 text-center text-xs text-muted-foreground">{t('auth.login.terms', { brand: brandName })}</p>
      </div>
    </div>
  )
}
