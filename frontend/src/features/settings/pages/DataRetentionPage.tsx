import { useCallback, useEffect, useState } from 'react'
import {
  AlertTriangle,
  Cloud,
  CloudOff,
  HardDriveDownload,
  Loader2,
  ShieldAlert,
  Snowflake,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { StorageHttpError, storageHttpService } from '../services/storage-http.service'
import type { DatasetUsage, Retention, StoreHealth, Tiering } from '../types/storage.types'

export function DataRetentionPage() {
  const { t } = useTranslation()

  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<'forbidden' | 'failed' | null>(null)
  const [retentions, setRetentions] = useState<Retention[]>([])
  const [usage, setUsage] = useState<DatasetUsage[]>([])
  const [health, setHealth] = useState<StoreHealth | null>(null)
  const [tiering, setTiering] = useState<Tiering | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [r, u, h, ti] = await Promise.all([
        storageHttpService.retention(),
        storageHttpService.usage(),
        storageHttpService.health(),
        storageHttpService.tiering(),
      ])
      setRetentions(r)
      setUsage(u)
      setHealth(h)
      setTiering(ti)
    } catch (err) {
      // Retention belongs to the platform, so a tenant admin gets a 403 here.
      // Saying "could not read the store" would send them looking for an
      // outage that is not happening.
      setError(err instanceof StorageHttpError && (err.status === 403 || err.status === 401)
        ? 'forbidden'
        : 'failed')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header className="flex items-center justify-between gap-3">
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <HardDriveDownload size={16} strokeWidth={1.75} />
          {t('dataRetention.title')}
        </h1>
        {health && <DiskBadge health={health} />}
      </header>

      <div className="mt-6 space-y-6">
        {loading ? (
          <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {t('dataRetention.loading')}
          </div>
        ) : error === 'forbidden' ? (
          <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center">
            <ShieldAlert size={28} strokeWidth={1.5} className="text-muted-foreground/60" />
            <div className="text-sm font-medium">{t('dataRetention.forbidden.title')}</div>
            <p className="max-w-md text-xs text-muted-foreground">
              {t('dataRetention.forbidden.body')}
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
          <>
            {retentions.map((r) => (
              <DatasetCard
                key={r.dataset}
                retention={r}
                usage={usage.find((u) => u.dataset === r.dataset)}
                coldAvailable={!!tiering?.ready}
                onSaved={() => void load()}
              />
            ))}

            <ColdStorageCard tiering={tiering} onSaved={() => void load()} />
          </>
        )}
      </div>
    </div>
  )
}

function DatasetCard({
  retention,
  usage,
  coldAvailable,
  onSaved,
}: {
  retention: Retention
  usage?: DatasetUsage
  coldAvailable: boolean
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [keepDays, setKeepDays] = useState(retention.keepDays)
  const [coldDays, setColdDays] = useState(retention.coldDays)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setKeepDays(retention.keepDays)
    setColdDays(retention.coldDays)
  }, [retention.keepDays, retention.coldDays])

  const dirty = keepDays !== retention.keepDays || coldDays !== retention.coldDays
  const valid =
    Number.isInteger(keepDays) && keepDays > 0 && coldDays >= 0 && (coldDays === 0 || coldDays < keepDays)

  const save = async () => {
    if (!valid || saving) return
    setSaving(true)
    try {
      await storageHttpService.setRetention({ dataset: retention.dataset, keepDays, coldDays })
      toast.success(t('dataRetention.toast.saved'))
      onSaved()
    } catch (err) {
      toast.error(
        err instanceof StorageHttpError && err.status === 403
          ? t('dataRetention.toast.noPermission')
          : err instanceof Error
            ? err.message
            : t('dataRetention.toast.saveFailed'),
      )
      setKeepDays(retention.keepDays)
      setColdDays(retention.coldDays)
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 className="text-sm font-semibold">{t(`dataRetention.datasets.${retention.dataset}`)}</h2>
        {usage && (
          <span className="text-xs text-muted-foreground">
            {t('dataRetention.holding', {
              documents: usage.documents.toLocaleString(),
              size: formatBytes(usage.bytes),
            })}
            {usage.oldest && ` · ${t('dataRetention.since', { date: formatDate(usage.oldest) })}`}
          </span>
        )}
      </div>

      <div className="mt-4 grid gap-5 sm:grid-cols-2">
        <div>
          <div className="flex items-center gap-2 text-sm font-medium">
            <Trash2 size={15} className="text-muted-foreground" />
            {t('dataRetention.keep')}
          </div>
          <p className="mt-1 text-xs text-muted-foreground">{t('dataRetention.keepHint')}</p>
          <div className="mt-3 flex items-center gap-2">
            <Input
              type="number"
              min={1}
              value={Number.isNaN(keepDays) ? '' : keepDays}
              onChange={(e) => setKeepDays(parseInt(e.target.value, 10))}
              className="h-9 w-28"
            />
            <span className="text-sm text-muted-foreground">{t('dataRetention.days')}</span>
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2 text-sm font-medium">
            <Snowflake size={15} className="text-muted-foreground" />
            {t('dataRetention.cold')}
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            {coldAvailable ? t('dataRetention.coldHint') : t('dataRetention.coldUnavailable')}
          </p>
          <div className="mt-3 flex items-center gap-2">
            <Input
              type="number"
              min={0}
              disabled={!coldAvailable}
              value={Number.isNaN(coldDays) ? '' : coldDays}
              onChange={(e) => setColdDays(parseInt(e.target.value, 10) || 0)}
              className="h-9 w-28"
            />
            <span className="text-sm text-muted-foreground">{t('dataRetention.days')}</span>
          </div>
        </div>
      </div>

      {!valid && (
        <p className="mt-3 text-xs text-red-500">{t('dataRetention.coldBeforeDelete')}</p>
      )}
      {retention.tiered && (
        <p className="mt-3 text-xs text-muted-foreground">{t('dataRetention.tieredNote')}</p>
      )}

      <div className="mt-4 flex items-center justify-end border-t border-border pt-4">
        <Button size="sm" disabled={!dirty || !valid || saving} onClick={() => void save()}>
          {saving ? t('dataRetention.saving') : t('dataRetention.save')}
        </Button>
      </div>
    </section>
  )
}

function ColdStorageCard({ tiering, onSaved }: { tiering: Tiering | null; onSaved: () => void }) {
  const { t } = useTranslation()
  const [endpoint, setEndpoint] = useState(tiering?.endpoint ?? '')
  const [accessKey, setAccessKey] = useState('')
  const [secretKey, setSecretKey] = useState('')
  const [saving, setSaving] = useState(false)

  const configured = !!tiering?.configured
  const ready = !!tiering?.ready
  const valid = endpoint.trim() !== '' && accessKey.trim() !== '' && secretKey.trim() !== ''

  const save = async () => {
    if (!valid || saving) return
    setSaving(true)
    try {
      await storageHttpService.enableTiering({ endpoint, accessKey, secretKey })
      toast.success(t('dataRetention.coldStorage.toast.saved'))
      setAccessKey('')
      setSecretKey('')
      onSaved()
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t('dataRetention.coldStorage.toast.failed'),
      )
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h2 className="flex items-center gap-2 text-sm font-semibold">
            {ready ? (
              <Cloud size={15} className="text-muted-foreground" />
            ) : (
              <CloudOff size={15} className="text-muted-foreground" />
            )}
            {t('dataRetention.coldStorage.title')}
          </h2>
          <p className="mt-1 max-w-xl text-xs text-muted-foreground">
            {t('dataRetention.coldStorage.hint')}
          </p>
        </div>
        <span
          className={
            'shrink-0 rounded-full px-2 py-0.5 text-[11px] ' +
            (ready
              ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
              : 'bg-muted text-muted-foreground')
          }
        >
          {ready
            ? t('dataRetention.coldStorage.active')
            : t('dataRetention.coldStorage.inactive')}
        </span>
      </div>

      {configured && !ready && (
        <p className="mt-3 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          {t('dataRetention.coldStorage.writtenNotReady')}
        </p>
      )}

      <div className="mt-4 grid gap-4 sm:grid-cols-3">
        <div className="sm:col-span-3">
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            {t('dataRetention.coldStorage.endpoint')}
          </label>
          <Input
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="https://s3.eu-west-1.amazonaws.com/utmstack-cold/"
            className="h-9"
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            {t('dataRetention.coldStorage.accessKey')}
          </label>
          <Input value={accessKey} onChange={(e) => setAccessKey(e.target.value)} className="h-9" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-muted-foreground">
            {t('dataRetention.coldStorage.secretKey')}
          </label>
          <Input
            type="password"
            autoComplete="off"
            value={secretKey}
            onChange={(e) => setSecretKey(e.target.value)}
            className="h-9"
          />
        </div>
      </div>

      {configured && (
        <p className="mt-3 text-xs text-muted-foreground">
          {t('dataRetention.coldStorage.bucketLocked')}
        </p>
      )}

      <div className="mt-4 flex items-center justify-end border-t border-border pt-4">
        <Button size="sm" disabled={!valid || saving} onClick={() => void save()}>
          {saving
            ? t('dataRetention.saving')
            : configured
              ? t('dataRetention.coldStorage.update')
              : t('dataRetention.coldStorage.enable')}
        </Button>
      </div>
    </section>
  )
}

function DiskBadge({ health }: { health: StoreHealth }) {
  const { t } = useTranslation()
  const warn = health.status !== 'ok'
  return (
    <span
      className={
        'rounded-full px-2 py-0.5 text-[11px] ' +
        (warn
          ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300'
          : 'bg-muted text-muted-foreground')
      }
      title={health.message}
    >
      {t('dataRetention.diskUsed', { pct: health.diskUsedPct.toFixed(1) })}
    </span>
  )
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleDateString()
}
