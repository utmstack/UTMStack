import { forwardRef, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, ImageOff, Loader2, Paintbrush, RotateCcw, Upload } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useBilling } from '@/features/billing'
import { EnterpriseGate } from '@/shared/components/EnterpriseGate'
import { PlatformBroadcastButton, broadcast, broadcastBrandingAsset, BULK_PATHS } from '@/features/platform-broadcast'
import { brandingHttpService } from '../services/branding-http.service'
import { useBranding } from '../services/branding.context'
import type { Branding, BrandingAssetSlot } from '../types/branding.types'

const PALETTE = ['#8b5cf6', '#3b82f6', '#0ea5e9', '#14b8a6', '#10b981', '#f59e0b', '#f43f5e', '#d946ef']
const DEFAULT_ACCENT = '#8b5cf6'

const EMPTY: Branding = {
  enabled: false,
  productName: 'UTMStack',
  accentColor: '',
  logoUrl: '',
  logoDarkUrl: '',
  faviconUrl: '',
  reportLogoUrl: '',
  reportCoverUrl: '',
}

export function BrandingPage() {
  const { t } = useTranslation()
  const { license } = useBilling()
  const { refresh: refreshGlobal } = useBranding()
  const isEnterprise = license?.edition === 'enterprise'

  const [form, setForm] = useState<Branding>(EMPTY)
  const [initial, setInitial] = useState<Branding>(EMPTY)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [highlightToggle, setHighlightToggle] = useState(false)
  const toggleRef = useRef<HTMLElement>(null)

  useEffect(() => {
    let cancelled = false
    brandingHttpService
      .get()
      .then((b) => {
        if (cancelled) return
        const v = { ...b, productName: b.productName || 'UTMStack' }
        setForm(v)
        setInitial(v)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('branding.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  const patch = (p: Partial<Branding>) => setForm((f) => ({ ...f, ...p }))
  const dirty = JSON.stringify(form) !== JSON.stringify(initial)

  // Restoring is a save of the empty shape with the brand switched off, not a
  // separate endpoint: the uploaded files stay on disk, so turning it back on
  // recovers the previous look without re-uploading anything.
  const restoreDefaults = async () => {
    setSaving(true)
    try {
      const saved = await brandingHttpService.update({ ...EMPTY, enabled: false })
      const v = { ...saved, productName: saved.productName || 'UTMStack' }
      setForm(v)
      setInitial(v)
      await refreshGlobal()
      toast.success(t('branding.restored'))
    } catch {
      toast.error(t('branding.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const save = async () => {
    setSaving(true)
    try {
      const saved = await brandingHttpService.update({
        enabled: form.enabled,
        productName: form.productName,
        accentColor: form.accentColor,
        logoUrl: form.logoUrl,
        logoDarkUrl: form.logoDarkUrl,
        faviconUrl: form.faviconUrl,
        reportLogoUrl: form.reportLogoUrl,
        reportCoverUrl: form.reportCoverUrl,
      })
      const v = { ...saved, productName: saved.productName || 'UTMStack' }
      setForm(v)
      setInitial(v)
      await refreshGlobal()
      toast.success(t('branding.saved'))
    } catch {
      toast.error(t('branding.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const onUploaded = async (b: Branding) => {
    const v = { ...b, productName: b.productName || 'UTMStack' }
    setForm(v)
    setInitial(v)
    await refreshGlobal()
  }

  if (!isEnterprise) {
    return (
      <EnterpriseGate
        header={<header>
            <h1 className="flex items-center gap-2 text-base font-semibold">
              <Paintbrush size={16} strokeWidth={1.75} />
              {t('branding.title')}
            </h1>
          </header>}
        title={t('branding.enterprise.title')}
        body={t('branding.enterprise.body')}
        cta={t('branding.enterprise.upgrade')}
      />
    )
  }

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <Paintbrush size={16} strokeWidth={1.75} />
          {t('branding.title')}
        </h1>
      </header>


      {loading ? (
        <div className="mt-6 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : (
        <fieldset disabled={!isEnterprise || saving} className="mt-6 space-y-5">
          {/* Enable */}
          <Section ref={toggleRef} title={t('branding.enable.title')} subtitle={t('branding.enable.description')}>
            <label className={cn(
              'flex cursor-pointer items-center justify-between gap-4 rounded-md px-2 py-1 transition-all',
              highlightToggle && 'animate-pulse ring-2 ring-primary ring-offset-2 ring-offset-card',
            )}>
              <span className="text-sm">{t('branding.enable.label')}</span>
              <Toggle enabled={form.enabled} onChange={() => patch({ enabled: !form.enabled })} />
            </label>
          </Section>

          <fieldset disabled={!form.enabled} className="m-0 space-y-5 border-0 p-0" onClickCapture={(e) => {
            if (!form.enabled) {
              e.stopPropagation()
              e.preventDefault()
              toggleRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
              setHighlightToggle(true)
              setTimeout(() => setHighlightToggle(false), 1500)
            }
          }}>
          {/* Identity */}
          <Section title={t('branding.identity.title')}>
            <div className="space-y-4">
              <Field label={t('branding.identity.productName')} hint={t('branding.identity.productNameHint')}>
                <Input
                  value={form.productName}
                  onChange={(e) => patch({ productName: e.target.value })}
                  placeholder="UTMStack"
                  className="max-w-sm"
                />
              </Field>

              <Field label={t('branding.identity.accent')} hint={t('branding.identity.accentHint')}>
                <div className="flex flex-wrap items-center gap-2">
                  {PALETTE.map((hex) => {
                    const active = (form.accentColor || DEFAULT_ACCENT).toLowerCase() === hex.toLowerCase()
                    return (
                      <button
                        key={hex}
                        type="button"
                        onClick={() => patch({ accentColor: hex })}
                        className={cn(
                          'flex h-7 w-7 items-center justify-center rounded-full ring-2 ring-offset-2 ring-offset-card transition-all',
                          active ? 'ring-foreground/40' : 'ring-transparent hover:ring-border',
                        )}
                        style={{ background: hex }}
                        aria-label={hex}
                      >
                        {active && <Check size={13} className="text-white" strokeWidth={3} />}
                      </button>
                    )
                  })}
                  <span className="mx-1 h-5 w-px bg-border" />
                  <input
                    type="color"
                    value={form.accentColor || DEFAULT_ACCENT}
                    onChange={(e) => patch({ accentColor: e.target.value })}
                    className="h-7 w-9 cursor-pointer rounded border border-border bg-transparent p-0.5"
                    aria-label={t('branding.identity.custom')}
                  />
                  <Input
                    value={form.accentColor}
                    onChange={(e) => patch({ accentColor: e.target.value })}
                    placeholder={DEFAULT_ACCENT}
                    className="h-7 w-28 font-mono text-xs"
                  />
                </div>
              </Field>
            </div>
          </Section>

          {/* Logos */}
          <Section title={t('branding.logos.title')} subtitle={t('branding.logos.hint')}>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
              <AssetCard slot="logo" label={t('branding.logos.logo')} url={form.logoUrl} disabled={!isEnterprise} onUploaded={onUploaded} t={t} />
              <AssetCard slot="logoDark" label={t('branding.logos.logoDark')} url={form.logoDarkUrl} dark disabled={!isEnterprise} onUploaded={onUploaded} t={t} />
              <AssetCard slot="favicon" label={t('branding.logos.favicon')} url={form.faviconUrl} disabled={!isEnterprise} onUploaded={onUploaded} t={t} />
            </div>
          </Section>

          {/* Reports */}
          <Section title={t('branding.reports.title')} subtitle={t('branding.reports.hint')}>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <AssetCard slot="reportLogo" label={t('branding.reports.reportLogo')} url={form.reportLogoUrl} disabled={!isEnterprise} onUploaded={onUploaded} t={t} />
              <AssetCard slot="reportCover" label={t('branding.reports.reportCover')} url={form.reportCoverUrl} disabled={!isEnterprise} onUploaded={onUploaded} t={t} />
            </div>
          </Section>

          <div className="flex items-center justify-between gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={saving || !isEnterprise}
              onClick={() => void restoreDefaults()}
            >
              <RotateCcw size={13} className="mr-1.5" />
              {t('branding.restore')}
            </Button>
            <div className="flex items-center gap-2">
              <Button size="sm" disabled={!dirty || saving || !isEnterprise} onClick={() => void save()}>
                {saving ? t('branding.saving') : t('branding.save')}
              </Button>
              {dirty && (
                <PlatformBroadcastButton
                  label={t('platformBroadcast.button')}
                  title={t('platformBroadcast.action.update', { resource: t('platformBroadcast.resource.branding') })}
                  disabled={saving || !isEnterprise}
                  excludeDefaultTenant={true}
                  onBroadcast={async (selector) => {
                    return broadcast(BULK_PATHS.branding.update, selector, {
                      enabled: form.enabled,
                      productName: form.productName,
                      accentColor: form.accentColor,
                      logoUrl: form.logoUrl,
                      logoDarkUrl: form.logoDarkUrl,
                      faviconUrl: form.faviconUrl,
                      reportLogoUrl: form.reportLogoUrl,
                      reportCoverUrl: form.reportCoverUrl,
                    })
                  }}
                />
              )}
            </div>
          </div>
          </fieldset>
        </fieldset>
      )}
    </div>
  )
}

/* ─── Asset upload card ────────────────────────────────────────────────── */

function AssetCard({
  slot,
  label,
  url,
  dark,
  disabled,
  onUploaded,
  t,
}: {
  slot: BrandingAssetSlot
  label: string
  url: string
  dark?: boolean
  disabled?: boolean
  onUploaded: (b: Branding) => void | Promise<void>
  t: ReturnType<typeof useTranslation>['t']
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [busy, setBusy] = useState(false)
  const [pickedFile, setPickedFile] = useState<File | null>(null)

  const pick = () => inputRef.current?.click()

  const onFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) {
      e.target.value = ''
      return
    }
    setPickedFile(file)
  }

  const uploadToTenant = async () => {
    if (!pickedFile) return
    setBusy(true)
    try {
      const b = await brandingHttpService.uploadAsset(slot, pickedFile)
      await onUploaded(b)
      setPickedFile(null)
      toast.success(t('branding.uploaded'))
    } catch {
      toast.error(t('branding.uploadError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="flex flex-col gap-2 rounded-lg border border-border bg-card p-3">
      <div className="text-xs font-medium">{label}</div>
      <div
        className={cn(
          'flex h-20 items-center justify-center overflow-hidden rounded-md border border-dashed border-border',
          dark ? 'bg-zinc-900' : 'bg-muted/30',
        )}
      >
        {url ? (
          <img src={url} alt={label} className="max-h-16 max-w-[80%] object-contain" />
        ) : (
          <ImageOff size={18} className="text-muted-foreground" />
        )}
      </div>
      <input ref={inputRef} type="file" accept="image/png,image/jpeg,image/webp,image/gif,image/svg+xml,image/x-icon" hidden onChange={onFile} />
      <div className="flex gap-2">
        <Button variant="outline" size="sm" disabled={disabled || busy} onClick={pick} className="flex-1">
          {busy ? (
            <Loader2 size={13} className="mr-1.5 animate-spin" />
          ) : (
            <Upload size={13} className="mr-1.5" />
          )}
          {pickedFile ? pickedFile.name : url ? t('branding.asset.replace') : t('branding.asset.upload')}
        </Button>
        {pickedFile && (
          <PlatformBroadcastButton
            label={t('platformBroadcast.button')}
            title={t('platformBroadcast.action.upload', { resource: slot })}
            disabled={disabled || busy}
            excludeDefaultTenant={true}
            onBroadcast={async (selector) => {
              return broadcastBrandingAsset(slot, pickedFile, selector)
            }}
          />
        )}
      </div>
      {pickedFile && (
        <Button size="sm" disabled={disabled || busy} onClick={() => void uploadToTenant()} className="w-full">
          {busy ? t('branding.uploading') : 'Upload to tenant'}
        </Button>
      )}
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

const Section = forwardRef<HTMLElement, { title: string; subtitle?: string; children: React.ReactNode }>(
  ({ title, subtitle, children }, ref) => {
    return (
      <section ref={ref} className="rounded-xl border border-border bg-card p-6">
        <header className="mb-4">
          <h2 className="text-sm font-semibold">{title}</h2>
          {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
        </header>
        {children}
      </section>
    )
  },
)

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: () => void }) {
  return (
    <button
      type="button"
      onClick={onChange}
      className={cn('relative h-5 w-9 shrink-0 rounded-full transition-colors', enabled ? 'bg-primary' : 'bg-muted')}
    >
      <span
        className={cn(
          'absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all',
          enabled ? 'left-[1.125rem]' : 'left-0.5',
        )}
      />
    </button>
  )
}
