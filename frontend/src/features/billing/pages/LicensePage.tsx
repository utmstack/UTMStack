import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  Building2,
  Database,
  KeyRound,
  Loader2,
  RefreshCw,
  ShieldCheck,
  Sparkles,
  Upload,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { BillingHttpError, billingHttpService } from '../services/billing-http.service'
import { useBilling } from '../services/billing.context'
import {
  COMMUNITY_DATASOURCE_CAP,
  daysUntilExpiry,
  hasExpiry,
  licenseStatus,
  type License,
  type LicenseStatus,
  type VersionInfo,
} from '../types/billing.types'

const STATUS_META: Record<LicenseStatus, { tone: string; dot: string }> = {
  community: { tone: 'bg-muted text-muted-foreground ring-border', dot: 'bg-muted-foreground' },
  active: { tone: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
  expiring: { tone: 'bg-amber-500/15 text-amber-600 ring-amber-500/30 dark:text-amber-300', dot: 'bg-amber-500' },
  expired: { tone: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300', dot: 'bg-red-500' },
}

export function LicensePage() {
  const { t } = useTranslation()
  const { license, version, isLoading, error, refresh } = useBilling()
  const showLoading = !license && isLoading
  const showError = !license && error

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
      <header className="flex items-start justify-between gap-3">
        <div>
          <h1 className="flex items-center gap-2 text-base font-semibold">
            <KeyRound size={16} strokeWidth={1.75} />
            {t('license.title')}
          </h1>
        </div>
        <Button variant="outline" size="sm" onClick={() => void refresh()} disabled={isLoading}>
          <RefreshCw size={13} className={cn('mr-1.5', isLoading && 'animate-spin')} />
          {t('common.actions.refresh')}
        </Button>
      </header>

      {showLoading && (
        <div className="mt-10 flex items-center justify-center text-muted-foreground">
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {t('license.loading')}
        </div>
      )}

      {showError && (
        <div className="mt-6 flex items-center gap-3 rounded-md border border-red-500/30 bg-red-500/5 p-4 text-sm">
          <AlertTriangle size={16} className="shrink-0 text-red-500" />
          <span>{t('license.loadError')}</span>
          <Button variant="outline" size="sm" className="ml-auto" onClick={() => void refresh()}>
            {t('common.actions.retry')}
          </Button>
        </div>
      )}

      {license && (
        <div className="mt-6 space-y-5">
          <LicenseCard license={license} version={version} />
          <UploadCard onUploaded={() => void refresh()} />
        </div>
      )}
    </div>
  )
}

function LicenseCard({ license, version }: { license: License; version: VersionInfo | null }) {
  const { t, i18n } = useTranslation()
  const status = licenseStatus(license)
  const meta = STATUS_META[status]
  const isEnterprise = license.edition === 'enterprise'
  const days = daysUntilExpiry(license)
  const editionName = isEnterprise ? t('license.enterprise') : t('license.community')
  const datasourceLimit = !isEnterprise
    ? String(COMMUNITY_DATASOURCE_CAP)
    : license.datasources === 0
      ? t('license.unlimited')
      : license.datasources.toLocaleString()

  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_340px]">
        {/* Left — edition + version */}
        <div className="border-b border-border p-6 lg:border-b-0 lg:border-r">
          <div className="flex items-start gap-3">
            <img src="/logo.svg" alt="UTMStack" className="h-12 w-12" />
            <div className="min-w-0 flex-1">
              <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
                {t('license.editionLine', { edition: editionName })}
              </div>
              <div className="mt-0.5 flex items-baseline gap-2">
                <span className="text-3xl font-semibold">UTMStack</span>
                {version?.version && (
                  <span className="font-mono text-lg text-muted-foreground">{version.version}</span>
                )}
              </div>
            </div>
          </div>

          <dl className="mt-6 grid grid-cols-1 gap-y-2.5 text-xs sm:grid-cols-[160px_1fr]">
            <Row k={t('license.edition')}>
              <span className={cn('font-medium', isEnterprise ? 'text-primary' : 'text-foreground')}>
                {editionName}
              </span>
            </Row>
            <Row k={t('license.datasourceLimit')}>
              <span className="inline-flex items-center gap-1.5">
                <Database size={11} className="text-muted-foreground" />
                {datasourceLimit}
              </span>
            </Row>
            {license.mssp && (
              <Row k={t('license.mssp')}>
                <span className="font-medium text-primary">{t('license.enabled')}</span>
              </Row>
            )}
            {isEnterprise && license.type && (
              <Row k={t('license.licenseType')}>
                <span className="capitalize">{license.type}</span>
              </Row>
            )}
            {version?.instanceId && (
              <Row k={t('license.instanceId')}>
                <code className="font-mono text-[11px]">{version.instanceId}</code>
              </Row>
            )}
          </dl>
        </div>

        {/* Right — status */}
        <div className="bg-muted/20 p-6">
          <div className="flex items-center justify-between">
            <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
              {t('license.statusLabel')}
            </div>
            <span
              className={cn(
                'inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset',
                meta.tone,
              )}
            >
              <span className={cn('h-1 w-1 rounded-full', meta.dot)} />
              {t(`license.status.${status}`)}
            </span>
          </div>

          {isEnterprise ? (
            <div className="mt-3 space-y-3">
              <div className="flex items-center gap-2 text-sm">
                <ShieldCheck size={16} className="text-primary" />
                <span className="font-medium">{t('license.enterpriseUnlocked')}</span>
              </div>
              <dl className="space-y-2 text-xs">
                <KV
                  k={t('license.validUntil')}
                  v={
                    hasExpiry(license) ? (
                      <span className="font-mono">
                        {new Date(license.expiresAt as string).toLocaleDateString(i18n.language)}
                      </span>
                    ) : (
                      t('license.noExpiry')
                    )
                  }
                />
                {days != null && (
                  <KV
                    k={t('license.timeRemaining')}
                    v={
                      days < 0 ? (
                        <span className="font-medium text-red-500">
                          {t('license.expiredAgo', { days: Math.abs(days) })}
                        </span>
                      ) : (
                        <span className={cn('font-medium', status === 'expiring' && 'text-amber-500')}>
                          {t('license.daysShort', { days })}
                        </span>
                      )
                    }
                  />
                )}
                {license.mssp && <KV k={t('license.msspShort')} v={t('license.enabled')} />}
              </dl>
            </div>
          ) : (
            <div className="mt-3">
              <div className="flex items-center gap-2 text-sm font-medium">
                <Building2 size={16} className="text-muted-foreground" />
                {t('license.communityEdition')}
              </div>
              <p className="mt-1.5 text-xs text-muted-foreground">{t('license.communityBlurb')}</p>
              <Button size="sm" className="mt-3 w-full" asChild>
                <a href="https://utmstack.com/pricing" target="_blank" rel="noopener">
                  <Sparkles size={13} className="mr-1.5" />
                  {t('license.upgrade')}
                </a>
              </Button>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function UploadCard({ onUploaded }: { onUploaded: () => void }) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [file, setFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)

  const submit = async () => {
    if (!file || uploading) return
    setUploading(true)
    try {
      const lic = await billingHttpService.uploadLicense(file)
      onUploaded()
      setFile(null)
      if (inputRef.current) inputRef.current.value = ''
      toast.success(
        t('license.toast.installed', {
          edition: lic.edition === 'enterprise' ? t('license.enterprise') : t('license.community'),
        }),
      )
    } catch (err) {
      if (err instanceof BillingHttpError) {
        if (err.status === 403) {
          toast.error(t('license.toast.adminRequired'))
        } else if (err.status === 400) {
          toast.error(err.message || t('license.toast.invalid'))
        } else {
          toast.error(err.message || t('license.toast.installFailed'))
        }
      } else {
        toast.error(t('license.toast.installFailed'))
      }
    } finally {
      setUploading(false)
    }
  }

  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <header className="mb-4">
        <h2 className="flex items-center gap-2 text-sm font-semibold">
          <Upload size={15} strokeWidth={1.75} />
          {t('license.uploadTitle')}
        </h2>
        <p className="mt-0.5 text-xs text-muted-foreground">{t('license.uploadBlurb')}</p>
      </header>

      <div className="flex flex-wrap items-center gap-3">
        <input
          ref={inputRef}
          type="file"
          onChange={(e) => setFile(e.target.files?.[0] ?? null)}
          className="hidden"
        />
        <Button type="button" variant="outline" size="sm" onClick={() => inputRef.current?.click()}>
          {t('license.chooseFile')}
        </Button>
        <span className="max-w-[220px] truncate text-xs text-muted-foreground">
          {file ? file.name : t('license.noFileSelected')}
        </span>
        <Button size="sm" onClick={() => void submit()} disabled={!file || uploading}>
          {uploading ? (
            <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
          ) : (
            <Upload size={13} className="mr-1.5" />
          )}
          {t('license.installLicense')}
        </Button>
      </div>
    </section>
  )
}

function KV({ k, v }: { k: string; v: React.ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <span className="text-[11px] text-muted-foreground">{k}</span>
      <span className="text-right">{v}</span>
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}
