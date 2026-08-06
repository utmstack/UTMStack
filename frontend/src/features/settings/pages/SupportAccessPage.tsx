import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  Check,
  Eye,
  Loader2,
  Lock,
  LifeBuoy,
  ShieldCheck,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { useAuth } from '@/features/auth'
import type { SupportAccess } from '@/features/tenants/types/tenant.types'
import {
  SupportAccessHttpError,
  supportAccessHttpService,
} from '../services/support-access-http.service'

const LEVELS: { value: SupportAccess; icon: LucideIcon; tone: string }[] = [
  { value: 'NONE', icon: Lock, tone: 'text-muted-foreground' },
  { value: 'READ', icon: Eye, tone: 'text-sky-500' },
  { value: 'FULL', icon: ShieldCheck, tone: 'text-amber-500' },
]

export function SupportAccessPage() {
  const { t } = useTranslation()
  const { tenantId } = useAuth()
  const [current, setCurrent] = useState<SupportAccess | null>(null)
  const [choice, setChoice] = useState<SupportAccess | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    if (!tenantId) return
    setLoading(true)
    setFailed(false)
    try {
      const { supportAccess } = await supportAccessHttpService.get(tenantId)
      setCurrent(supportAccess)
      setChoice(supportAccess)
    } catch {
      setFailed(true)
    } finally {
      setLoading(false)
    }
  }, [tenantId])

  useEffect(() => {
    void load()
  }, [load])

  const dirty = choice !== null && choice !== current

  const save = async () => {
    if (!tenantId || !choice || !dirty || busy) return
    setBusy(true)
    try {
      await supportAccessHttpService.set(tenantId, choice)
      setCurrent(choice)
      toast.success(
        choice === 'NONE' ? t('supportAccess.toast.revoked') : t('supportAccess.toast.granted')
      )
    } catch (err) {
      toast.error(errorMessage(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-3xl px-6 pb-10 pt-4">
      <header className="flex items-center gap-2 text-xs text-muted-foreground">
        <LifeBuoy size={14} strokeWidth={1.75} />
        <span className="font-medium text-foreground">{t('supportAccess.title')}</span>
      </header>
      <p className="mt-2 text-sm text-muted-foreground">{t('supportAccess.intro')}</p>

      {loading && (
        <div className="flex items-center gap-2 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t('supportAccess.loading')}
        </div>
      )}

      {!loading && failed && (
        <div className="mt-6 flex flex-col items-start gap-3 rounded-xl border border-border bg-card px-5 py-6 text-sm">
          <span className="inline-flex items-center gap-2 text-muted-foreground">
            <AlertTriangle size={16} className="text-amber-500" />
            {t('supportAccess.loadFailed')}
          </span>
          <Button variant="outline" size="sm" onClick={() => void load()}>
            {t('supportAccess.retry')}
          </Button>
        </div>
      )}

      {!loading && !failed && (
        <>
          <div className="mt-5 space-y-2">
            {LEVELS.map(({ value, icon: Icon, tone }) => {
              const selected = choice === value
              return (
                <button
                  key={value}
                  type="button"
                  onClick={() => setChoice(value)}
                  className={cn(
                    'flex w-full items-start gap-3 rounded-xl border p-4 text-left transition-colors',
                    selected
                      ? 'border-primary/40 bg-primary/5'
                      : 'border-border hover:bg-muted/40'
                  )}
                >
                  <span
                    className={cn(
                      'mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border',
                      selected ? 'border-primary bg-primary text-primary-foreground' : 'border-input'
                    )}
                  >
                    {selected && <Check size={10} strokeWidth={3} />}
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <Icon size={14} className={tone} />
                      <span className="text-sm font-medium">
                        {t(`supportAccess.levels.${value}.name`)}
                      </span>
                      {current === value && (
                        <span className="rounded bg-muted px-1.5 py-px text-[10px] font-medium text-muted-foreground">
                          {t('supportAccess.currentBadge')}
                        </span>
                      )}
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {t(`supportAccess.levels.${value}.desc`)}
                    </p>
                  </div>
                </button>
              )
            })}
          </div>

          {choice === 'FULL' && (
            <div className="mt-4 flex items-start gap-2 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3 text-xs text-amber-700 dark:text-amber-300">
              <AlertTriangle size={14} className="mt-px shrink-0" />
              <span>{t('supportAccess.fullWarning')}</span>
            </div>
          )}

          <div className="mt-6 flex items-center justify-between gap-3 border-t border-border pt-4">
            <p className="text-[11px] text-muted-foreground">{t('supportAccess.revokeNote')}</p>
            <Button size="sm" disabled={!dirty || busy} onClick={() => void save()}>
              {busy ? t('supportAccess.saving') : t('supportAccess.save')}
            </Button>
          </div>
        </>
      )}
    </div>
  )
}

function errorMessage(err: unknown, t: (k: string) => string): string {
  if (err instanceof SupportAccessHttpError) {
    // The backend refuses this for anyone but an administrator of this very
    // tenant, and refuses it outright for the platform's own tenant.
    if (err.status === 403) return t('supportAccess.toast.forbidden')
    if (err.status === 400) return err.message || t('supportAccess.toast.invalid')
  }
  return err instanceof Error ? err.message : t('supportAccess.toast.failed')
}
