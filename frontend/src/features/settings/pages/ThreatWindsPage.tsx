import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, RadioTower } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { threatIntelHttpService } from '@/features/threat-intel/services/threat-intel-http.service'

export function ThreatWindsPage() {
  const { t } = useTranslation()
  const [status, setStatus] = useState<{ enabled: boolean } | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setFailed(false)
    try {
      setStatus(await threatIntelHttpService.contribution())
    } catch {
      setFailed(true)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const toggle = async (enabled: boolean) => {
    if (!status || busy) return
    setBusy(true)
    try {
      const next = await threatIntelHttpService.setContribution(enabled)
      setStatus(next)
      toast.success(enabled ? t('threatwinds.toast.enabled') : t('threatwinds.toast.disabled'))
    } catch {
      toast.error(t('threatwinds.toast.failed'))
    } finally {
      setBusy(false)
    }
  }

  const enabled = status?.enabled ?? false

  return (
    <div className="mx-auto w-full max-w-3xl px-6 pb-10 pt-4">
      <header className="flex items-center gap-2 text-xs text-muted-foreground">
        <RadioTower size={14} strokeWidth={1.75} />
        <span className="font-medium text-foreground">{t('threatwinds.title')}</span>
      </header>
      <p className="mt-2 text-sm text-muted-foreground">{t('threatwinds.subtitle')}</p>

      <div className="mt-6 rounded-lg border border-border p-4">
        {loading ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 size={16} className="animate-spin" />
            {t('threatwinds.loading')}
          </div>
        ) : failed ? (
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">{t('threatwinds.failed')}</p>
            <button type="button" onClick={() => void load()} className="text-sm underline">
              {t('threatwinds.retry')}
            </button>
          </div>
        ) : (
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <div className="text-sm font-medium">{t('threatwinds.contribute')}</div>
              <p className="mt-0.5 text-xs text-muted-foreground">{t('threatwinds.contributeHint')}</p>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={enabled}
              disabled={busy}
              onClick={() => toggle(!enabled)}
              className={cn(
                'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50',
                enabled ? 'bg-primary' : 'bg-muted-foreground/30',
              )}
            >
              <span
                className={cn(
                  'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
                  enabled ? 'translate-x-4' : 'translate-x-0.5',
                )}
              />
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
