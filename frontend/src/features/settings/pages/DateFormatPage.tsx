import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Clock, Globe2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { createApiClient } from '@/shared/lib/api-client'
import { setDateFormatPrefs, type DateFormatPrefs, type DateFormatStyle } from '@/shared/lib/datetime'
import { configHttpService } from '../services/config-http.service'

const KEY_TZ = 'utmstack.time.zone'
const KEY_FMT = 'utmstack.time.dateformat'

const STYLES: DateFormatStyle[] = ['short', 'medium', 'long', 'full']

const ALL_ZONES: string[] = (() => {
  try {
    const supported = (Intl as unknown as { supportedValuesOf?: (k: string) => string[] }).supportedValuesOf
    const zones = supported ? supported('timeZone') : []
    return Array.from(new Set(['UTC', ...zones]))
  } catch {
    return ['UTC']
  }
})()

function previewIn(tz: string, style: DateFormatStyle): string {
  try {
    return new Intl.DateTimeFormat(undefined, {
      dateStyle: style,
      timeStyle: 'medium',
      timeZone: tz === 'system' ? undefined : tz,
    }).format(new Date())
  } catch {
    return '—'
  }
}

const api = createApiClient()

export function DateFormatPage() {
  const { t } = useTranslation()
  const [timezone, setTimezone] = useState('UTC')
  const [dateFormat, setDateFormat] = useState<DateFormatStyle>('medium')
  const [initial, setInitial] = useState<DateFormatPrefs>({ timezone: 'UTC', dateFormat: 'medium' })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    api
      .get<DateFormatPrefs>('/date-format')
      .then((d) => {
        if (cancelled) return
        const fmt = (STYLES.includes(d.dateFormat) ? d.dateFormat : 'medium') as DateFormatStyle
        setTimezone(d.timezone || 'UTC')
        setDateFormat(fmt)
        setInitial({ timezone: d.timezone || 'UTC', dateFormat: fmt })
      })
      .catch(() => {
        if (!cancelled) toast.error(t('dateFormat.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  const dirty = timezone !== initial.timezone || dateFormat !== initial.dateFormat

  const save = async () => {
    setSaving(true)
    try {
      const updates: Promise<unknown>[] = []
      if (timezone !== initial.timezone) updates.push(configHttpService.set(KEY_TZ, { value: timezone }))
      if (dateFormat !== initial.dateFormat) updates.push(configHttpService.set(KEY_FMT, { value: dateFormat }))
      await Promise.all(updates)
      setInitial({ timezone, dateFormat })
      // Apply immediately to this session (the rest of the app re-renders).
      setDateFormatPrefs({ timezone, dateFormat })
      toast.success(t('dateFormat.saved'))
    } catch {
      toast.error(t('dateFormat.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const zoneOptions = useMemo(() => ['system', ...ALL_ZONES], [])

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <header>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <Clock size={18} strokeWidth={1.75} />
          {t('dateFormat.title')}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{t('dateFormat.subtitle')}</p>
      </header>

      {loading ? (
        <div className="mt-6 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : (
        <div className="mt-6 space-y-5">
          {/* Timezone */}
          <Section title={t('dateFormat.timezone.title')} subtitle={t('dateFormat.timezone.hint')}>
            <div className="relative max-w-md">
              <Globe2 size={14} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <select
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                className="h-9 w-full appearance-none rounded-md border border-border bg-background pl-9 pr-3 text-sm"
              >
                {zoneOptions.map((z) => (
                  <option key={z} value={z}>
                    {z === 'system' ? t('dateFormat.timezone.system') : z}
                  </option>
                ))}
              </select>
            </div>
            <p className="mt-2 text-[11px] text-muted-foreground">
              {t('dateFormat.timezone.now', { value: previewIn(timezone, dateFormat) })}
            </p>
          </Section>

          {/* Format */}
          <Section title={t('dateFormat.format.title')} subtitle={t('dateFormat.format.hint')}>
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              {STYLES.map((s) => {
                const active = dateFormat === s
                return (
                  <button
                    key={s}
                    type="button"
                    onClick={() => setDateFormat(s)}
                    className={cn(
                      'flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5 text-left transition-colors',
                      active ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted/50',
                    )}
                  >
                    <span className="min-w-0">
                      <span className="block text-sm font-medium">{t(`dateFormat.styles.${s}`)}</span>
                      <span className="block truncate font-mono text-[11px] text-muted-foreground">
                        {previewIn(timezone, s)}
                      </span>
                    </span>
                    {active && <Check size={15} className="shrink-0 text-primary" />}
                  </button>
                )
              })}
            </div>
          </Section>

          <div className="flex justify-end">
            <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
              {saving ? t('dateFormat.saving') : t('dateFormat.save')}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function Section({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <header className="mb-4">
        <h2 className="text-sm font-semibold">{title}</h2>
        {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
      </header>
      {children}
    </section>
  )
}
