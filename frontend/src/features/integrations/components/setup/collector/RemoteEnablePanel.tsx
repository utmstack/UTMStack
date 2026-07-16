import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Check, Copy, Loader2, ShieldAlert, ShieldCheck, Upload } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Section } from '@/features/integrations/components/ui/Section'
import { useAuth } from '@/features/auth/services/auth.context'
import { useCollectorIntegration } from '@/features/integrations/hooks/useCollectorIntegration'
import { defaultPortFor, httpDefaultsFor, type HttpAuth, type Proto } from './protoCatalog'

const ROOT = 'integrations.setup.remoteEnable'

// The Forwarder is about to ship its first real release as 1.0.0 — this is
// the minimum version that understands the generic remote-config channel
// (SetCollectorConfig/GetCollectorConfig). No wire-exposed capability
// negotiation exists yet, so this is a hand-maintained gate, same
// accepted-drift tradeoff as protoCatalog.ts.
export const MIN_REMOTE_CONFIG_VERSION = '1.0.0'

/** Numeric "at least" comparison for dotted version strings (major.minor.patch, ...). */
export function isVersionAtLeast(version: string, min: string): boolean {
  const parse = (v: string) => v.split('.').map((n) => parseInt(n, 10) || 0)
  const a = parse(version)
  const b = parse(min)
  const len = Math.max(a.length, b.length)
  for (let i = 0; i < len; i++) {
    const x = a[i] ?? 0
    const y = b[i] ?? 0
    if (x !== y) return x > y
  }
  return true
}

export interface RemoteEnableSelection {
  proto: Proto
  port: string
}

interface RemoteEnablePanelProps {
  /** OpenSearch source_type / custom integration name this panel configures. */
  dataType: string
  /** Protocols to offer, in display order. */
  availableProtos: Proto[]
  /** Initial protocol selection, when the vendor knows its device needs something other than the first entry (e.g. TCP-only sources). Defaults to availableProtos[0]. */
  defaultProto?: Proto
  step?: number
  /** Fires whenever proto/port change, so the parent can render a matching manual command. */
  onSelectionChange?: (selection: RemoteEnableSelection) => void
}

async function fileToBase64(file: File): Promise<string> {
  // PEM certs/keys are ASCII text, so a plain text read + btoa is safe (no
  // need for the ArrayBuffer/binary-string dance non-ASCII files would need).
  const text = await file.text()
  return btoa(text)
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

  const { forwarders, setDataType, setCertificates, tlsStatus } = useCollectorIntegration()

  const eligible = useMemo(
    () => (forwarders.data ?? []).filter((f) => isVersionAtLeast(f.version, MIN_REMOTE_CONFIG_VERSION)),
    [forwarders.data],
  )

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

  // Pick the first eligible forwarder once the list loads (or clear it if the
  // previously selected one dropped off — e.g. went offline).
  useEffect(() => {
    if (collectorId != null && eligible.some((f) => f.id === collectorId)) return
    setCollectorId(eligible[0]?.id ?? null)
  }, [eligible, collectorId])

  // Reset the port to the catalog default when the protocol changes (still
  // freely editable afterwards).
  useEffect(() => {
    setPort(defaultPortFor(dataType, proto))
    setGeneratedSecret(null)
  }, [proto, dataType])

  useEffect(() => {
    onSelectionChange?.({ proto, port })
  }, [proto, port, onSelectionChange])

  const isHttp = proto === 'http' || proto === 'https'
  const needsCerts = proto === 'tls' || proto === 'https'

  const tlsStatusQuery = tlsStatus(needsCerts ? collectorId : null)

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
  const canSubmit = collectorId != null && !!port

  return (
    <Section title={t(`${ROOT}.title`)} step={step}>
      <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.body`)}</p>

      {forwarders.isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t(`${ROOT}.loadingForwarders`)}
        </div>
      ) : eligible.length === 0 ? (
        <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${ROOT}.noForwarders`, { minVersion: MIN_REMOTE_CONFIG_VERSION })}
        </p>
      ) : (
        <div className="space-y-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <label className="block">
              <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                {t(`${ROOT}.forwarderLabel`)}
              </span>
              <select
                value={collectorId ?? ''}
                onChange={(e) => setCollectorId(Number(e.target.value))}
                className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground"
              >
                {eligible.map((f) => (
                  <option key={f.id} value={f.id}>
                    {f.hostname} ({f.ip})
                  </option>
                ))}
              </select>
            </label>

            <label className="block">
              <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
                {t(`${ROOT}.protoLabel`)}
              </span>
              <select
                value={proto}
                onChange={(e) => setProto(e.target.value as Proto)}
                disabled={availableProtos.length <= 1}
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

          <label className="block max-w-[160px]">
            <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
              {t(`${ROOT}.portLabel`)}
            </span>
            <input
              value={port}
              onChange={(e) => setPort(e.target.value.replace(/[^0-9]/g, ''))}
              inputMode="numeric"
              className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm"
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
                  className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground"
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
                  className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm"
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
                    className="h-9 w-full rounded-md border border-border bg-background px-3 font-mono text-sm"
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
                    <CertFileInput label={t(`${ROOT}.certPemLabel`)} file={certPem} onChange={setCertPem} />
                    <CertFileInput label={t(`${ROOT}.keyPemLabel`)} file={keyPem} onChange={setKeyPem} />
                    <CertFileInput
                      label={t(`${ROOT}.caPemLabel`)}
                      file={caPem}
                      onChange={setCaPem}
                      optional
                    />
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={!certPem || !keyPem || setCertificates.isPending}
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
            <Button size="sm" disabled={!canSubmit || isPending} onClick={() => submit(true)}>
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
}: {
  label: string
  file: File | null
  onChange: (file: File | null) => void
  optional?: boolean
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
        className={cn(
          'block w-full text-xs text-muted-foreground',
          'file:mr-2 file:rounded-md file:border-0 file:bg-muted file:px-2 file:py-1 file:text-[11px] file:font-medium file:text-foreground',
        )}
      />
      {file && <span className="mt-0.5 block truncate text-[10px] text-emerald-600 dark:text-emerald-400">{file.name}</span>}
    </label>
  )
}
