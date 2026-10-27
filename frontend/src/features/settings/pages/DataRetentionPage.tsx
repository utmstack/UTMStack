import { useCallback, useEffect, useState } from 'react'
import { AlertTriangle, HardDriveDownload, Loader2, ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { StorageHttpError, storageHttpService } from '../services/storage-http.service'
import type { DatasetUsage, Retention, StoreHealth, Tiering } from '../types/storage.types'
import { DatasetCard } from './DatasetCard'
import { ColdStorageCard } from './ColdStorageCard'
import { DiskBadge } from './DiskBadge'

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
