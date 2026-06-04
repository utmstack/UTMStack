import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Loader2, Lock, User as UserIcon } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { AuthHttpError } from '../services/auth-http.service'
import { useAuth } from '../services/auth.context'

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: { pathname: string } } | null)?.from?.pathname ?? '/'

  const [loginValue, setLoginValue] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!loginValue || !password || submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await login({ login: loginValue, password })
      navigate(from, { replace: true })
    } catch (err) {
      if (err instanceof AuthHttpError) {
        if (err.status === 429) {
          setError('Too many failed attempts. Try again in a few minutes.')
        } else if (err.status === 403) {
          setError('This account is deactivated.')
        } else if (err.status === 401) {
          setError('Invalid credentials.')
        } else {
          setError(err.message || 'Login failed.')
        }
      } else {
        setError('Login failed. Please try again.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-background px-4 text-foreground">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'radial-gradient(70% 55% at 50% -10%, hsl(232 56% 60% / 0.22) 0%, transparent 65%), radial-gradient(50% 40% at 100% 110%, hsl(280 60% 55% / 0.12) 0%, transparent 60%)',
        }}
      />

      <div className="relative z-10 w-full max-w-sm">
        <div className="mb-10 flex flex-col items-center text-center">
          <img
            src="/logo.svg"
            alt="UTMStack"
            className="mb-6 h-16 w-auto"
            draggable={false}
          />
          <h1 className="text-[22px] font-semibold tracking-tight">Welcome back</h1>
          <p className="mt-1.5 text-sm text-muted-foreground">Sign in to continue to UTMStack</p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="rounded-xl border border-border/80 bg-card/60 px-6 py-7 shadow-xl shadow-black/40 backdrop-blur-md"
        >
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label
                htmlFor="login"
                className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
              >
                Login or email
              </label>
              <div className="relative">
                <UserIcon
                  aria-hidden
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                  id="login"
                  type="text"
                  placeholder="admin"
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
                  Password
                </label>
                <button
                  type="button"
                  className="text-[11px] text-muted-foreground transition-colors hover:text-foreground"
                  onClick={() => {
                    /* TODO: open reset modal once /auth/reset-password lands (Phase 5) */
                  }}
                >
                  Forgot?
                </button>
              </div>
              <div className="relative">
                <Lock
                  aria-hidden
                  className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                  id="password"
                  type="password"
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

          {error && (
            <p className="mt-4 text-xs text-destructive">{error}</p>
          )}

          <Button
            type="submit"
            className="mt-6 h-10 w-full"
            disabled={submitting || !loginValue || !password}
          >
            {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Sign in
          </Button>
        </form>

        <p className="mt-6 text-center text-xs text-muted-foreground">
          By signing in, you agree to UTMStack's terms of service.
        </p>
      </div>
    </div>
  )
}
