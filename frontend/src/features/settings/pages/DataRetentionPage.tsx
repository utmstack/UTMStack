import { useCallback, useEffect, useState } from 'react'
import {
  AlertTriangle,
  Archive,
  HardDriveDownload,
  Loader2,
  ShieldAlert,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import {
  DataRetentionHttpError,
  dataRetentionHttpService,
} from '../services/data-retention-http.service'
import type { PolicySettings } from '../types/data-retention.types'

/* OpenSearch ISM index-age units. Retention is virtually always in days. */
type Unit = 'd' | 'h'
const DEFAULT_VALUE = 90
const DEFAULT_UNIT: Unit = 'd'

/** Parse an ISM age string like "90d" into a number + unit. */
function parseAge(age: string | undefined): { value: number; unit: Unit } {
  const m = (age ?? '').trim().match(/^(\d+)\s*([a-z]+)$/i)
  if (m) {
    const n = Number(m[1])
    const u = m[2].toLowerCase()
    if (Number.isFinite(n) && (u === 'd' || u === 'h')) return { value: n, unit: u }
    if (Number.isFinite(n)) return { value: n, unit: 'd' }
  }
  return { value: DEFAULT_VALUE, unit: DEFAULT_UNIT }
}

export function DataRetentionPage() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [unavailable, setUnavailable] = useState(false)
  const [error, setError] = useState(false)
  const [saving, setSaving] = useState(false)

  // The loaded policy (null = none configured yet → start from defaults).
  const [loaded, setLoaded] = useState<PolicySettings | null>(null)
  const [value, setValue] = useState<number>(DEFAULT_VALUE)
  const [unit, setUnit] = useState<Unit>(DEFAULT_UNIT)
  const [snapshot, setSnapshot] = useState(false)

  const applyPolicy = (p: PolicySettings | null) => {
    setLoaded(p)
    const { value: v, unit: u } = parseAge(p?.deleteAfter)
    setValue(v)
    setUnit(u)
    setSnapshot(!!p?.snapshotActive)
  }

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    setUnavailable(false)
    try {
      const resp = await dataRetentionHttpService.get()
      // 200 with an empty body → no policy configured yet.
      applyPolicy(resp && typeof resp === 'object' ? (resp as PolicySettings) : null)
    } catch (err) {
      if (err instanceof DataRetentionHttpError && err.status === 404) {
        setUnavailable(true)
      } else {
        setError(true)
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const currentAge = `${value}${unit}`
  const dirty =
    !loaded || loaded.deleteAfter !== currentAge || loaded.snapshotActive !== snapshot
  const valid = Number.isInteger(value) && value > 0

  const save = async () => {
    if (!valid || saving) return
    setSaving(true)
    try {
      await dataRetentionHttpService.update({ deleteAfter: currentAge, snapshotActive: snapshot })
      toast.success(t('dataRetention.toast.saved'))
      await load()
    } catch (err) {
      const msg =
        err instanceof DataRetentionHttpError && err.status === 403
          ? t('dataRetention.toast.noPermission')
          : err instanceof Error
            ? err.message
            : t('dataRetention.toast.saveFailed')
      toast.error(msg)
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <HardDriveDownload size={16} strokeWidth={1.75} />
          {t('dataRetention.title')}
        </h1>
      </header>

      <div className="mt-6">
        {loading ? (
          <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t('dataRetention.loading')}
          </div>
        ) : unavailable ? (
          <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center">
            <ShieldAlert size={28} strokeWidth={1.5} className="text-muted-foreground/60" />
            <div className="text-sm font-medium">{t('dataRetention.unavailable.title')}</div>
            <p className="max-w-md text-xs text-muted-foreground">
              {t('dataRetention.unavailable.body')}
            </p>
          </div>
        ) : error ? (
          <div className="flex flex-col items-center gap-3 rounded-xl border border-border bg-card px-6 py-12 text-sm">
            <span className="inline-flex items-center gap-2 text-muted-foreground">
              <AlertTriangle size={16} className="text-amber-500" />
              {t('dataRetention.loadFailed')}
            </span>
            <Button variant="outline" size="sm" onClick={() => void load()}>
              {t('dataRetention.retry')}
            </Button>
          </div>
        ) : (
          <section className="space-y-5 rounded-xl border border-border bg-card p-6">
            {!loaded && (
              <p className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
                <AlertTriangle size={14} className="mt-0.5 shrink-0" />
                {t('dataRetention.noPolicyNote')}
              </p>
            )}

            {/* Retention period */}
            <div>
              <div className="flex items-center gap-2 text-sm font-medium">
                <Trash2 size={15} className="text-muted-foreground" />
                {t('dataRetention.deleteAfter')}
              </div>
              <p className="mt-1 text-xs text-muted-foreground">{t('dataRetention.deleteAfterHint')}</p>
              <div className="mt-3 flex items-center gap-2">
                <Input
                  type="number"
                  min={1}
                  value={Number.isNaN(value) ? '' : value}
                  onChange={(e) => setValue(parseInt(e.target.value, 10))}
                  className="h-9 w-28"
                />
                <select
                  value={unit}
                  onChange={(e) => setUnit(e.target.value as Unit)}
                  className="h-9 cursor-pointer rounded-md border border-input bg-background/40 px-2 text-sm transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  <option value="d">{t('dataRetention.units.days')}</option>
                  <option value="h">{t('dataRetention.units.hours')}</option>
                </select>
                {!valid && (
                  <span className="text-xs text-red-500">{t('dataRetention.positiveNumber')}</span>
                )}
              </div>
            </div>

            {/* Snapshot before delete */}
            <div className="border-t border-border pt-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Archive size={15} className="text-muted-foreground" />
                    {t('dataRetention.snapshotTitle')}
                  </div>
                  <p className="mt-1 max-w-md text-xs text-muted-foreground">
                    {t('dataRetention.snapshotHint')}
                  </p>
                </div>
                <Toggle checked={snapshot} onChange={setSnapshot} />
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 border-t border-border pt-4">
              <Button size="sm" disabled={!dirty || !valid || saving} onClick={() => void save()}>
                {saving ? t('dataRetention.saving') : t('dataRetention.save')}
              </Button>
            </div>
          </section>
        )}
      </div>
    </div>
  )
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors',
        checked ? 'bg-primary' : 'bg-muted-foreground/30',
      )}
    >
      <span
        className={cn(
          'inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform',
          checked ? 'translate-x-4' : 'translate-x-0.5',
        )}
      />
    </button>
  )
}
