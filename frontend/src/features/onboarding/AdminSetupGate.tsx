import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Mail, ShieldAlert } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { IS_FEDERATION } from '@/shared/config/mode'
import { useAuth } from '@/features/auth'

// The seeded admin ships with the placeholder "admin"; treat it (and an empty
// email) as "not configured yet".
const DEFAULT_ADMIN_EMAIL = 'admin'

function needsAdminEmail(email: string | undefined): boolean {
  const e = (email ?? '').trim().toLowerCase()
  return e === '' || e === DEFAULT_ADMIN_EMAIL
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

/**
 * First-run blocking gate (ported from v11): an admin still on the default email
 * ("admin") must set a real address before using the app — password
 * reset, notifications, alert emails and threat-intel registration all depend on
 * it. The modal can't be dismissed; after saving, the admin is guided to the
 * mail-server settings (non-blocking). Instance mode only — the federation
 * console has its own onboarding.
 */
export function AdminSetupGate() {
  const { user, isAdmin, updateMe } = useAuth()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (IS_FEDERATION || !isAdmin || !user || !needsAdminEmail(user.email)) return null

  const valid = EMAIL_RE.test(email.trim())

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!valid || saving) return
    setSaving(true)
    setError(null)
    try {
      // Send the existing name/lang too so a partial update never wipes them.
      await updateMe({
        email: email.trim(),
        name: user.name,
        lang_key: user.lang_key,
      })
      toast.success(t('onboarding.adminEmail.success'))
      // Guide the admin to configure the mail server next (like v11). The gate
      // closes on its own once the user's email updates in context.
      navigate('/settings/email')
    } catch {
      setError(t('onboarding.adminEmail.error'))
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-2xl">
        <div className="mb-4 flex items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/15 text-primary">
            <ShieldAlert size={20} strokeWidth={1.75} />
          </span>
          <div className="min-w-0">
            <h2 className="text-base font-semibold">{t('onboarding.adminEmail.title')}</h2>
            <p className="text-xs text-muted-foreground">{t('onboarding.adminEmail.subtitle')}</p>
          </div>
        </div>

        <p className="mb-4 text-sm text-muted-foreground">{t('onboarding.adminEmail.description')}</p>

        <form onSubmit={submit} className="space-y-3">
          <label htmlFor="admin-email" className="block text-xs font-medium text-foreground/80">
            {t('onboarding.adminEmail.emailLabel')}
          </label>
          <div className="relative">
            <Mail
              aria-hidden
              className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="admin-email"
              type="email"
              autoFocus
              required
              autoComplete="email"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                setError(null)
              }}
              placeholder="you@company.com"
              className="pl-9"
            />
          </div>
          {error && <p className="text-xs text-destructive">{error}</p>}
          <Button type="submit" className="mt-2 h-10 w-full" disabled={!valid || saving}>
            {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {saving ? t('onboarding.adminEmail.saving') : t('onboarding.adminEmail.save')}
          </Button>
        </form>
      </div>
    </div>
  )
}
