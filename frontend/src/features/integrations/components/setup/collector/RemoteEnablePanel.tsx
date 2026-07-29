import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Check, ChevronDown, Copy, Loader2, ShieldAlert, ShieldCheck, Upload } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Section } from '@/features/integrations/components/ui/Section'
import { useAuth } from '@/features/auth/services/auth.context'
import { useCollectorIntegration } from '@/features/integrations/hooks/useCollectorIntegration'
import type { ForwarderCollector } from '@/features/integrations/types'
import { defaultPortFor, httpDefaultsFor, type HttpAuth, type Proto } from './protoCatalog'

const ROOT = 'integrations.setup.remoteEnable'

export interface RemoteEnableSelection {
  proto: Proto
  port: string
}

interface RemoteEnablePanelProps {
  dataType: string
  availableProtos: Proto[]
  defaultProto?: Proto
  step?: number
  onSelectionChange?: (selection: RemoteEnableSelection) => void
}

async function fileToBase64(file: File): Promise<string> {
  const text = await file.text()
  return btoa(text)
}

function StatusDot({ status }: { status: ForwarderCollector['status'] }) {
  return (
    <span
      className={cn('h-1.5 w-1.5 shrink-0 rounded-full', status === 'online' ? 'bg-emerald-500' : 'bg-red-500')}
    />
  )
}

// Custom dropdown (not a native <select>) so each row — open or closed — can
// show a real colored status dot next to it. Native <option> elements can't
// host styled children, only plain text, so a native select can't do this.
function ForwarderPicker({
  forwarders,
  value,
  onChange,
  disabled,
}: {
  forwarders: ForwarderCollector[]
  value: number | null
  onChange: (id: number) => void
  disabled?: boolean
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const selected = forwarders.find((f) => f.id === value) ?? null

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        disabled={disabled}
        className="flex h-9 w-full items-center justify-between gap-2 rounded-md border border-border bg-background px-3 text-sm text-foreground disabled:opacity-70"
      >
        <span className="flex min-w-0 items-center gap-2">
          {selected && <StatusDot status={selected.status} />}
          <span className="truncate">{selected ? `${selected.hostname} (${selected.ip})` : ''}</span>
        </span>
        <ChevronDown size={14} className="shrink-0 text-muted-foreground" />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-full rounded-md border border-border bg-popover py-1 shadow-lg">
          {forwarders.map((f) => (
            <button
              key={f.id}
              type="button"
              onClick={() => {
                onChange(f.id)
                setOpen(false)
              }}
              className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted"
            >
              <StatusDot status={f.status} />
              <span className="truncate">
                {f.hostname} ({f.ip})
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function ForwarderStatusLegend({ forwarders }: { forwarders: ForwarderCollector[] }) {
  const { t } = useTranslation()
  const online = forwarders.filter((f) => f.status === 'online').length
  const offline = forwarders.length - online

  return (
    <div className="flex items-center gap-2.5 text-[11px] text-muted-foreground">
      <span className="flex items-center gap-1">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
        {t(`${ROOT}.legendOnline`, { count: online })}
      </span>
      <span className="flex items-center gap-1">
        <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
        {t(`${ROOT}.legendOffline`, { count: offline })}
      </span>
      <span className="flex items-center gap-1">
        <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground/50" />
        {t(`${ROOT}.legendTotal`, { count: forwarders.length })}
      </span>
    </div>
  )
}

export function RemoteEnablePanel({
  dataType,
  availableProtos,
  defaultProto,
  step,
  onSelectionChange,
}: RemoteEnablePanelProps) {
  const { t } = useTranslation()
  const { hasPermission } = useAuth()
  const canManage = hasPermission('integrations.write')

  const { forwarders, setDataType, setCertificates, tlsStatus, dataTypeConfig } = useCollectorIntegration()

  const allForwarders = forwarders.data ?? []

  const initialProto = defaultProto ?? availableProtos[0]
  const [collectorId, setCollectorId] = useState<number | null>(null)
  const [proto, setProto] = useState<Proto>(initialProto)
  const [port, setPort] = useState(() => defaultPortFor(dataType, initialProto))
  const httpDefaults = httpDefaultsFor(dataType)
  const [auth, setAuth] = useState<HttpAuth>(httpDefaults?.auth ?? '')
  const [path, setPath] = useState(httpDefaults?.path ?? '')
  const [signatureHeader, setSignatureHeader] = useState(httpDefaults?.signatureHeader ?? '')
  const [generatedSecret, setGeneratedSecret] = useState<string | null>(null)
  const [secretCopied, setSecretCopied] = useState(false)

  const handleCopySecret = () => {
    if (!generatedSecret) return
    navigator.clipboard.writeText(generatedSecret)
    setSecretCopied(true)
    setTimeout(() => setSecretCopied(false), 2000)
  }

  const [certPem, setCertPem] = useState<File | null>(null)
  const [keyPem, setKeyPem] = useState<File | null>(null)
  const [caPem, setCaPem] = useState<File | null>(null)

  const selectedForwarder = allForwarders.find((f) => f.id === collectorId) ?? null
  const isSelectedForwarderOffline = selectedForwarder != null && selectedForwarder.status !== 'online'

  useEffect(() => {
    if (collectorId != null && allForwarders.some((f) => f.id === collectorId)) return
    const preferred = allForwarders.find((f) => f.status === 'online') ?? allForwarders[0]
    setCollectorId(preferred?.id ?? null)
  }, [allForwarders, collectorId])

  useEffect(() => {
    onSelectionChange?.({ proto, port })
  }, [proto, port, onSelectionChange])

  const isHttp = proto === 'http' || proto === 'https'
  const needsCerts = proto === 'tls' || proto === 'https'

  const tlsStatusQuery = tlsStatus(needsCerts ? collectorId : null)

  const currentConfig = dataTypeConfig(collectorId, dataType)
  const isSyncingConfig = collectorId != null && currentConfig.isLoading

  const isLiveOnSelectedForwarder = !!(currentConfig.data?.configured && currentConfig.data?.enabled)
  const isConfigLocked = isSyncingConfig || isLiveOnSelectedForwarder

  useEffect(() => {
    setGeneratedSecret(null)
  }, [])

  useEffect(() => {
    if (collectorId == null) return
    if (!currentConfig.isSuccess) return

    const data = currentConfig.data

    if (data.configured) {
      const nextProto = (data.proto as Proto) || initialProto
      setProto(nextProto)
      setPort(data.port || defaultPortFor(dataType, nextProto))
      setAuth((data.auth as HttpAuth) ?? '')
      setPath(data.path ?? '')
      setSignatureHeader(data.signatureHeader ?? '')
      return
    }

    setProto(initialProto)
    setPort(defaultPortFor(dataType, initialProto))
    setAuth(httpDefaults?.auth ?? '')
    setPath(httpDefaults?.path ?? '')
    setSignatureHeader(httpDefaults?.signatureHeader ?? '')
  }, [collectorId, currentConfig.isSuccess, currentConfig.data])

  if (!canManage) {
    return (
      <Section title={t(`${ROOT}.title`)} step={step}>
        <p className="text-sm text-muted-foreground">{t(`${ROOT}.noPermission`)}</p>
      </Section>
    )
  }

  const handleUploadCerts = async () => {
    if (!collectorId || !certPem || !keyPem) return
    try {
      const data = {
        certPem: await fileToBase64(certPem),
        keyPem: await fileToBase64(keyPem),
        ...(caPem ? { caPem: await fileToBase64(caPem) } : {}),
      }
      setCertificates.mutate(
        { collectorId, data },
        {
          onSuccess: (resp) => {
            if (resp.accepted) {
              toast.success(t(`${ROOT}.certUploadSuccess`))
              setCertPem(null)
              setKeyPem(null)
              setCaPem(null)
            } else {
              toast.error(resp.errorMessage || t(`${ROOT}.certUploadError`))
            }
          },
          onError: (err) => toast.error(err instanceof Error ? err.message : t(`${ROOT}.certUploadError`)),
        },
      )
    } catch {
      toast.error(t(`${ROOT}.certReadError`))
    }
  }

  const submit = (enabled: boolean) => {
    if (!collectorId || !port) return
    setDataType.mutate(
      {
        collectorId,
        dataType,
        data: {
          enabled,
          proto,
          port,
          ...(isHttp ? { auth: auth || undefined, path: path || undefined } : {}),
          ...(isHttp && auth === 'hmac' ? { signatureHeader: signatureHeader || undefined } : {}),
        },
      },
      {
        onSuccess: (resp) => {
          if (resp.accepted) {
            toast.success(enabled ? t(`${ROOT}.enableSuccess`) : t(`${ROOT}.disableSuccess`))
            setGeneratedSecret(resp.generatedSecret || null)
          } else {
            toast.error(resp.errorMessage || t(`${ROOT}.enableError`))
          }
        },
        onError: (err) => toast.error(err instanceof Error ? err.message : t(`${ROOT}.enableError`)),
      },
    )
  }

  const isPending = setDataType.isPending
  const canSubmit = collectorId != null && !!port && !isSyncingConfig && !isSelectedForwarderOffline

  return (
    <Section title={t(`${ROOT}.title`)} step={step}>
      <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.body`)}</p>

      {forwarders.isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t(`${ROOT}.loadingForwarders`)}
        </div>
      ) : allForwarders.length === 0 ? (
        <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${ROOT}.noForwarders`)}
        </p>
      ) : (
        <div className="space-y-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <label className="block">
              <div className="mb-1 flex items-center justify-between gap-2">
                <span className="text-[11px] font-medium text-muted-foreground">
                  {t(`${ROOT}.forwarderLabel`)}
                </span>
                <ForwarderStatusLegend forwarders={allForwarders} />
              </div>
              <ForwarderPicker
                forwarders={allForwarders}
                value={collectorId}
                onChange={setCollectorId}
                disabled={isSyncingConfig}
              />
              {isSelectedForwarderOffline && (
                <p className="mt-1 text-[11px] text-red-600 dark:text-red-400">
                  {t(`${ROOT}.forwarderOffline`)}
                </p>
              )}
            </label>

            <label className="block">
              <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                {t(`${ROOT}.protoLabel`)}
              </span>
              <select
                value={proto}
                onChange={(e) => {
                  const next = e.target.value as Proto
                  setProto(next)
                  setPort(defaultPortFor(dataType, next))
                  setGeneratedSecret(null)
                }}
                disabled={availableProtos.length <= 1 || isConfigLocked}
                className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground disabled:opacity-70"
              >
                {availableProtos.map((p) => (
                  <option key={p} value={p}>
                    {p.toUpperCase()}
                  </option>
                ))}
              </select>
            </label>
          </div>

          {isSyncingConfig ? (
            <div className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <Loader2 className="h-3 w-3 animate-spin" />
              {t(`${ROOT}.currentStateLoading`)}
            </div>
          ) : (
            currentConfig.data?.configured && (
              <p className="text-[11px] text-muted-foreground">
                {t(`${ROOT}.${currentConfig.data.enabled ? 'currentlyEnabled' : 'currentlyDisabled'}`, {
                  proto: currentConfig.data.proto,
                  port: currentConfig.data.port,
                })}
              </p>
            )
          )}

          <label className="block max-w-[160px]">
            <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
              {t(`${ROOT}.portLabel`)}
            </span>
            <input
              value={port}
              onChange={(e) => setPort(e.target.value.replace(/[^0-9]/g, ''))}
              inputMode="numeric"
              disabled={isConfigLocked}
              className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm disabled:opacity-70"
              placeholder="7100"
            />
          </label>

          {isHttp && (
            <div className="grid gap-3 sm:grid-cols-2">
              <label className="block">
                <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                  {t(`${ROOT}.authLabel`)}
                </span>
                <select
                  value={auth}
                  onChange={(e) => setAuth(e.target.value as HttpAuth)}
                  disabled={isConfigLocked}
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground disabled:opacity-70"
                >
                  <option value="">{t(`${ROOT}.authNone`)}</option>
                  <option value="bearer">{t(`${ROOT}.authBearer`)}</option>
                  <option value="hmac">{t(`${ROOT}.authHmac`)}</option>
                </select>
              </label>

              <label className="block">
                <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                  {t(`${ROOT}.pathLabel`)}
                </span>
                <input
                  value={path}
                  onChange={(e) => setPath(e.target.value)}
                  disabled={isConfigLocked}
                  className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm disabled:opacity-70"
                  placeholder="/logs"
                />
              </label>

              {auth === 'hmac' && (
                <label className="block sm:col-span-2">
                  <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                    {t(`${ROOT}.signatureHeaderLabel`)}
                  </span>
                  <input
                    value={signatureHeader}
                    onChange={(e) => setSignatureHeader(e.target.value)}
                    disabled={isConfigLocked}
                    className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm disabled:opacity-70"
                    placeholder="X-Hub-Signature-256"
                  />
                </label>
              )}
            </div>
          )}

          {needsCerts && collectorId != null && (
            <div className="rounded-lg border border-border bg-muted/20 p-3">
              <p className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                {t(`${ROOT}.certTitle`)}
              </p>
              {tlsStatusQuery.isLoading ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  {t(`${ROOT}.certLoading`)}
                </div>
              ) : tlsStatusQuery.data?.available ? (
                <div className="flex items-start gap-2 text-sm text-emerald-600 dark:text-emerald-400">
                  <ShieldCheck size={16} className="mt-0.5 shrink-0" />
                  <span>{t(`${ROOT}.certLoaded`)}</span>
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="flex items-start gap-2 text-sm text-amber-700 dark:text-amber-400">
                    <ShieldAlert size={16} className="mt-0.5 shrink-0" />
                    <span>
                      {tlsStatusQuery.data?.error
                        ? t(`${ROOT}.certInvalid`, { error: tlsStatusQuery.data.error })
                        : t(`${ROOT}.certMissing`)}
                    </span>
                  </div>
                  <div className="grid gap-2 sm:grid-cols-3">
                    <CertFileInput
                      label={t(`${ROOT}.certPemLabel`)}
                      file={certPem}
                      onChange={setCertPem}
                      disabled={isConfigLocked}
                    />
                    <CertFileInput
                      label={t(`${ROOT}.keyPemLabel`)}
                      file={keyPem}
                      onChange={setKeyPem}
                      disabled={isConfigLocked}
                    />
                    <CertFileInput
                      label={t(`${ROOT}.caPemLabel`)}
                      file={caPem}
                      onChange={setCaPem}
                      optional
                      disabled={isConfigLocked}
                    />
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={!certPem || !keyPem || setCertificates.isPending || isConfigLocked}
                    onClick={handleUploadCerts}
                  >
                    {setCertificates.isPending ? (
                      <Loader2 size={13} className="mr-1.5 animate-spin" />
                    ) : (
                      <Upload size={13} className="mr-1.5" />
                    )}
                    {t(`${ROOT}.uploadCertsButton`)}
                  </Button>
                </div>
              )}
            </div>
          )}

          {generatedSecret && (
            <div className="rounded-md border border-amber-500/30 bg-amber-500/[0.07] px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">
              <p className="mb-1 font-semibold">{t(`${ROOT}.secretWarning`)}</p>
              <div className="flex items-center gap-2">
                <code className="flex-1 break-all font-mono text-xs text-foreground">{generatedSecret}</code>
                <button
                  type="button"
                  title={secretCopied ? t('integrations.codeBlock.copied') : t('integrations.codeBlock.copy')}
                  onClick={handleCopySecret}
                  className="flex shrink-0 items-center gap-1 rounded px-1.5 py-1 text-[10px] text-amber-700/80 transition-colors hover:bg-amber-500/10 dark:text-amber-300/80"
                >
                  {secretCopied ? (
                    <Check size={11} className="text-emerald-500" />
                  ) : (
                    <Copy size={11} />
                  )}
                  <span className={secretCopied ? 'text-emerald-500' : ''}>
                    {secretCopied ? t('integrations.codeBlock.copied') : t('integrations.codeBlock.copy')}
                  </span>
                </button>
              </div>
            </div>
          )}

          <div className="flex items-center gap-2 pt-1">
            <Button
              size="sm"
              disabled={!canSubmit || isPending || isLiveOnSelectedForwarder}
              onClick={() => submit(true)}
            >
              {isPending && <Loader2 size={13} className="mr-1.5 animate-spin" />}
              {t(`${ROOT}.enableButton`)}
            </Button>
            <Button
              size="sm"
              variant="outline"
              disabled={!canSubmit || isPending}
              onClick={() => submit(false)}
            >
              {t(`${ROOT}.disableButton`)}
            </Button>
          </div>
        </div>
      )}
    </Section>
  )
}

function CertFileInput({
  label,
  file,
  onChange,
  optional,
  disabled,
}: {
  label: string
  file: File | null
  onChange: (file: File | null) => void
  optional?: boolean
  disabled?: boolean
}) {
  return (
    <label className="block">
      <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
        {label}
        {optional ? ' (opt.)' : ''}
      </span>
      <input
        type="file"
        accept=".pem,.crt,.key,.cer"
        onChange={(e) => onChange(e.target.files?.[0] ?? null)}
        disabled={disabled}
        className={cn(
          'block w-full text-xs text-muted-foreground disabled:opacity-70',
          'file:mr-2 file:rounded-md file:border-0 file:bg-muted file:px-2 file:py-1 file:text-[11px] file:font-medium file:text-foreground',
        )}
      />
      {file && <span className="mt-0.5 block truncate text-[10px] text-emerald-600 dark:text-emerald-400">{file.name}</span>}
    </label>
  )
}
