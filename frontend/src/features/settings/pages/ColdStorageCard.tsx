import { useState } from 'react'
import { AlertTriangle, Cloud, CloudOff } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { storageHttpService } from '../services/storage-http.service'
import type { ColdKind, Tiering } from '../types/storage.types'

export function ColdStorageCard({ tiering, onSaved }: { tiering: Tiering | null; onSaved: () => void }) {
  const { t } = useTranslation()
  const [kind, setKind] = useState<ColdKind>(tiering?.kind ?? 's3')
  const [endpoint, setEndpoint] = useState(tiering?.kind === 'local' ? '' : (tiering?.endpoint ?? ''))
  const [container, setContainer] = useState('')
  const [useRole, setUseRole] = useState(false)
  const [accessKey, setAccessKey] = useState('')
  const [secretKey, setSecretKey] = useState('')
  const [saving, setSaving] = useState(false)

  const configured = !!tiering?.configured
  const ready = !!tiering?.ready
  const hasKeys = accessKey.trim() !== '' && secretKey.trim() !== ''
  const valid =
    kind === 'local' ||
    (kind === 'azure'
      ? endpoint.trim() !== '' && container.trim() !== '' && hasKeys
      : endpoint.trim() !== '' && (useRole || hasKeys))

  const save = async () => {
    if (!valid || saving) return
    setSaving(true)
    try {
      await storageHttpService.enableTiering(
        kind === 'local'
          ? { kind }
          : kind === 'azure'
            ? { kind, endpoint, container, accessKey, secretKey }
            : useRole
              ? { kind, endpoint, useRole }
              : { kind, endpoint, accessKey, secretKey },
      )
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

      <div className="mt-4 flex flex-wrap gap-2" role="radiogroup" aria-label={t('dataRetention.coldStorage.kind')}>
        {(['s3', 'azure', 'local'] as const).map((k) => (
          <button
            key={k}
            type="button"
            role="radio"
            aria-checked={kind === k}
            onClick={() => setKind(k)}
            className={
              'rounded-md border px-3 py-1.5 text-xs transition-colors ' +
              (kind === k
                ? 'border-primary bg-primary/10 text-foreground'
                : 'border-border text-muted-foreground hover:bg-muted')
            }
          >
            {t(`dataRetention.coldStorage.kind_${k}`)}
          </button>
        ))}
      </div>

      {kind === 'local' && (
        <p className="mt-4 text-xs text-muted-foreground">
          {t('dataRetention.coldStorage.localHint')}
        </p>
      )}

      {kind !== 'local' && (
        <div className="mt-4 grid gap-4 sm:grid-cols-3">
          {kind === 's3' && (
            <p className="text-xs text-muted-foreground sm:col-span-3">
              {t('dataRetention.coldStorage.s3Hint')}
            </p>
          )}
          <div className={kind === 'azure' ? 'sm:col-span-2' : 'sm:col-span-3'}>
            <label className="mb-1 block text-xs font-medium text-muted-foreground">
              {kind === 'azure'
                ? t('dataRetention.coldStorage.accountUrl')
                : t('dataRetention.coldStorage.endpoint')}
            </label>
            <Input
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder={
                kind === 'azure'
                  ? 'https://myaccount.blob.core.windows.net'
                  : 'https://s3.eu-west-1.amazonaws.com/utmstack-cold/'
              }
              className="h-9"
            />
          </div>
          {kind === 'azure' && (
            <div>
              <label className="mb-1 block text-xs font-medium text-muted-foreground">
                {t('dataRetention.coldStorage.container')}
              </label>
              <Input value={container} onChange={(e) => setContainer(e.target.value)} className="h-9" />
            </div>
          )}
          {kind === 's3' && (
            <label className="flex items-center gap-2 text-xs sm:col-span-3">
              <input
                type="checkbox"
                checked={useRole}
                onChange={(e) => setUseRole(e.target.checked)}
              />
              {t('dataRetention.coldStorage.useRole')}
            </label>
          )}
          {!(kind === 's3' && useRole) && (
            <>
              <div>
                <label className="mb-1 block text-xs font-medium text-muted-foreground">
                  {kind === 'azure'
                    ? t('dataRetention.coldStorage.accountName')
                    : t('dataRetention.coldStorage.accessKey')}
                </label>
                <Input value={accessKey} onChange={(e) => setAccessKey(e.target.value)} className="h-9" />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-muted-foreground">
                  {kind === 'azure'
                    ? t('dataRetention.coldStorage.accountKey')
                    : t('dataRetention.coldStorage.secretKey')}
                </label>
                <Input
                  type="password"
                  autoComplete="off"
                  value={secretKey}
                  onChange={(e) => setSecretKey(e.target.value)}
                  className="h-9"
                />
              </div>
            </>
          )}
        </div>
      )}

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
