import { useEffect, useState } from 'react'
import { Snowflake, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { StorageHttpError, storageHttpService } from '../services/storage-http.service'
import type { DatasetUsage, Retention } from '../types/storage.types'

export function DatasetCard({
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
