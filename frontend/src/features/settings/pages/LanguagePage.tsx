import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Globe2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { SUPPORTED_LANGUAGES } from '@/shared/i18n'
import { configHttpService } from '../services/config-http.service'

const LANGUAGE_KEY = 'utmstack.system.language'
const DEFAULT_LANGUAGE = 'en'

export function LanguagePage() {
  const { t } = useTranslation()
  const [value, setValue] = useState(DEFAULT_LANGUAGE)
  const [initial, setInitial] = useState(DEFAULT_LANGUAGE)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    configHttpService
      .get(LANGUAGE_KEY)
      .then((entry) => {
        if (cancelled) return
        const lang = entry.value || DEFAULT_LANGUAGE
        setValue(lang)
        setInitial(lang)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('languageSettings.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  const dirty = value !== initial

  const save = async () => {
    setSaving(true)
    try {
      await configHttpService.set(LANGUAGE_KEY, { value })
      setInitial(value)
      toast.success(t('languageSettings.saved'))
    } catch {
      toast.error(t('languageSettings.saveError'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <Globe2 size={16} strokeWidth={1.75} />
          {t('languageSettings.title')}
        </h1>
      </header>

      <div className="mt-6 rounded-xl border border-border bg-card p-5">
        <div className="text-sm font-medium">{t('languageSettings.label')}</div>
        <p className="mt-0.5 text-xs text-muted-foreground">{t('languageSettings.hint')}</p>

        {loading ? (
          <div className="mt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        ) : (
          <div className="mt-4 grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {SUPPORTED_LANGUAGES.map((l) => {
              const active = l.code === value
              return (
                <button
                  key={l.code}
                  type="button"
                  onClick={() => setValue(l.code)}
                  className={cn(
                    'flex items-center justify-between rounded-lg border px-3 py-2.5 text-left text-sm transition-colors',
                    active
                      ? 'border-primary bg-primary/5 text-foreground'
                      : 'border-border hover:bg-muted/50',
                  )}
                >
                  <span className="flex items-center gap-2">
                    <span>{l.label}</span>
                    <span className="font-mono text-[11px] text-muted-foreground">{l.code}</span>
                  </span>
                  {active && <Check size={15} className="text-primary" />}
                </button>
              )
            })}
          </div>
        )}

        <div className="mt-5 flex justify-end">
          <Button size="sm" disabled={!dirty || saving || loading} onClick={() => void save()}>
            {saving ? t('languageSettings.saving') : t('languageSettings.save')}
          </Button>
        </div>
      </div>
    </div>
  )
}
