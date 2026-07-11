import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bot, Eye, EyeOff, Loader2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { socaiHttpService } from '../services/socai-http.service'
import { useTiUsage } from '@/features/threat-intel/hooks/use-ti-usage'

// The path key CM's usage endpoint reports for SOC-AI traffic (see
// CustomersManager proxy/handlers/usage.go — keyed by rate-limited prefix).
const THREATWINDS_CHAT_USAGE_KEY = '/api/ai/v1/chat/completions'

/**
 * Provider registry — drives which fields the user sees. For the known cloud
 * providers we already know the endpoint and auth, so the user only supplies an
 * API key and picks a model. URL / auth / custom headers appear only where they
 * genuinely matter (azure, ollama, custom).
 */
interface ProviderDef {
  label: string
  models: string[] // curated; empty => free-text model only
  showUrl: boolean
  urlPlaceholder?: string
  apiKey: 'required' | 'none'
  authType: string // sent to the backend
  authHeaderName?: string // preset (e.g. azure)
  custom?: boolean // show full auth controls (auth type + header + custom headers)
  note?: string // optional i18n key with provider guidance
  fixedModel?: string // no model picker at all; always send this model
}

const CUSTOM_MODEL = '__custom__'

const PROVIDERS: Record<string, ProviderDef> = {
  threatwinds: {
    label: 'ThreatWinds',
    models: [],
    fixedModel: 'silas-1.0',
    showUrl: false,
    apiKey: 'none',
    authType: 'none',
    note: 'socAi.note.threatwinds',
  },
  openai: {
    label: 'OpenAI',
    models: ['gpt-4o', 'gpt-4o-mini', 'gpt-4.1', 'gpt-4.1-mini', 'o3', 'o4-mini'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  anthropic: {
    label: 'Anthropic (Claude)',
    models: ['claude-opus-4-8', 'claude-sonnet-4-6', 'claude-haiku-4-5-20251001'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  gemini: {
    label: 'Google Gemini',
    models: ['gemini-2.5-pro', 'gemini-2.5-flash', 'gemini-2.0-flash'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  groq: {
    label: 'Groq',
    models: ['llama-3.3-70b-versatile', 'llama-3.1-8b-instant'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  mistral: {
    label: 'Mistral',
    models: ['mistral-large-latest', 'mistral-small-latest'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  deepseek: {
    label: 'DeepSeek',
    models: ['deepseek-chat', 'deepseek-reasoner'],
    showUrl: false,
    apiKey: 'required',
    authType: 'bearer',
  },
  azure: {
    label: 'Azure OpenAI',
    models: [],
    showUrl: true,
    urlPlaceholder: 'https://<resource>.openai.azure.com/openai/deployments/<deployment>/chat/completions?api-version=…',
    apiKey: 'required',
    authType: 'header',
    authHeaderName: 'api-key',
    note: 'socAi.note.azure',
  },
  ollama: {
    label: 'Ollama (local)',
    models: [],
    showUrl: true,
    urlPlaceholder: 'http://localhost:11434/v1/chat/completions',
    apiKey: 'none',
    authType: 'none',
    note: 'socAi.note.ollama',
  },
  custom: {
    label: 'Custom (OpenAI-compatible)',
    models: [],
    showUrl: true,
    urlPlaceholder: 'https://my-gateway/v1/chat/completions',
    apiKey: 'required',
    authType: 'bearer',
    custom: true,
    note: 'socAi.note.custom',
  },
}

const PROVIDER_IDS = Object.keys(PROVIDERS)

/**
 * Authentication kind — custom-provider-only UI discriminator. Collapses the
 * backend's authType ('bearer' | 'header' | 'none') plus customHeaders into
 * one mutually-exclusive choice, since a gateway either sends a bearer token,
 * a single named header, a free-form set of headers, or nothing.
 */
type AuthKind = 'apiKey' | 'bearer' | 'customHeaders' | 'none'

const AUTH_KIND_TO_TYPE: Record<AuthKind, string> = {
  apiKey: 'header',
  bearer: 'bearer',
  customHeaders: 'none',
  none: 'none',
}

function authKindForProvider(def: ProviderDef): AuthKind {
  if (def.authType === 'bearer') return 'bearer'
  if (def.authType === 'header') return 'apiKey'
  return 'none'
}

function authKindFromConfig(authType: string, customHeaders: Record<string, string>): AuthKind {
  if (Object.keys(customHeaders).length > 0) return 'customHeaders'
  if (authType === 'bearer') return 'bearer'
  if (authType === 'header') return 'apiKey'
  return 'none'
}

/**
 * Capability groups the admin can grant the agent. IDs MUST match the plugin's
 * capability catalog (plugins/soc-ai/internal/agent/groups.go) and the backend
 * config. Read access is always on; these gate write/action tools per module.
 */
const CAPABILITY_GROUPS: { id: string; danger?: boolean }[] = [
  { id: 'alerts' },
  { id: 'incidents' },
  { id: 'dashboards' },
  { id: 'compliance' },
  { id: 'correlation' },
  { id: 'datasources' },
]

interface HeaderRow {
  key: string
  value: string
}

interface Form {
  provider: string
  model: string
  url: string
  apiKey: string
  authKind: AuthKind // custom provider only
  authHeaderName: string
  customHeadersList: HeaderRow[]
  maxTokens: string
  maxToolIterations: string
  autoAnalyze: boolean
  capabilities: string[] // enabled permission group ids
}

function emptyForm(provider = 'threatwinds'): Form {
  const def = PROVIDERS[provider]
  return {
    provider,
    model: def.fixedModel ?? def.models[0] ?? '',
    url: '',
    apiKey: '',
    authKind: authKindForProvider(def),
    authHeaderName: def.authHeaderName ?? '',
    customHeadersList: [],
    maxTokens: '4096',
    maxToolIterations: '12',
    autoAnalyze: true,
    capabilities: [],
  }
}

export function SocAiSettingsPage() {
  const { t } = useTranslation()
  const [form, setForm] = useState<Form>(() => emptyForm())
  const [initial, setInitial] = useState<Form>(() => emptyForm())
  const [modelCustom, setModelCustom] = useState(false)
  const [apiKeySet, setApiKeySet] = useState(false)
  const [apiKeyTouched, setApiKeyTouched] = useState(false)
  const [showKey, setShowKey] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    socaiHttpService
      .get()
      .then((cfg) => {
        if (cancelled) return
        const provider = cfg.provider && PROVIDERS[cfg.provider] ? cfg.provider : 'threatwinds'
        const def = PROVIDERS[provider]
        const v: Form = {
          provider,
          model: def.fixedModel ?? (cfg.model || def.models[0] || ''),
          url: cfg.url || '',
          apiKey: '',
          authKind: authKindFromConfig(cfg.authType || def.authType, cfg.customHeaders ?? {}),
          authHeaderName: cfg.authHeaderName || def.authHeaderName || '',
          customHeadersList: Object.entries(cfg.customHeaders ?? {}).map(([key, value]) => ({ key, value })),
          maxTokens: String(cfg.maxTokens || 4096),
          maxToolIterations: String(cfg.maxToolIterations || 12),
          autoAnalyze: cfg.autoAnalyze,
          capabilities: cfg.capabilities ?? [],
        }
        setForm(v)
        setInitial(v)
        setModelCustom(def.models.length === 0 || !def.models.includes(v.model))
        setApiKeySet(cfg.apiKeySet)
      })
      .catch(() => {
        if (!cancelled) toast.error(t('socAi.loadError'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [t])

  const def = PROVIDERS[form.provider]
  const patch = (p: Partial<Form>) => setForm((f) => ({ ...f, ...p }))

  const changeProvider = (provider: string) => {
    const d = PROVIDERS[provider]
    setForm((f) => ({
      ...f,
      provider,
      model: d.fixedModel ?? d.models[0] ?? '',
      url: '',
      authKind: authKindForProvider(d),
      authHeaderName: d.authHeaderName ?? '',
      customHeadersList: [],
    }))
    setModelCustom(d.models.length === 0)
  }

  const dirty =
    apiKeyTouched ||
    (Object.keys(initial) as (keyof Form)[]).some((k) => k !== 'apiKey' && form[k] !== initial[k])

  const save = async () => {
    if (!form.model.trim()) {
      toast.error(t('socAi.modelRequired'))
      return
    }

    const customHeaders: Record<string, string> = {}
    if (def.custom && form.authKind === 'customHeaders') {
      for (const row of form.customHeadersList) {
        const key = row.key.trim()
        if (key) customHeaders[key] = row.value
      }
    }

    const authType = def.custom ? AUTH_KIND_TO_TYPE[form.authKind] : def.authType
    const authHeaderName = def.custom
      ? form.authKind === 'apiKey'
        ? form.authHeaderName.trim()
        : ''
      : (def.authHeaderName ?? '')
    const apiKeyEligible = def.custom ? form.authKind === 'apiKey' || form.authKind === 'bearer' : def.apiKey !== 'none'

    setSaving(true)
    try {
      await socaiHttpService.update({
        provider: form.provider,
        model: form.model.trim(),
        url: def.showUrl ? form.url.trim() : '',
        apiKey: apiKeyEligible && apiKeyTouched ? form.apiKey : '',
        authType,
        authHeaderName,
        customHeaders,
        maxTokens: parseInt(form.maxTokens, 10) || 4096,
        maxToolIterations: parseInt(form.maxToolIterations, 10) || 12,
        autoAnalyze: form.autoAnalyze,
        capabilities: form.capabilities,
      })
      const next = { ...form, apiKey: '' }
      setInitial(next)
      setForm(next)
      if (apiKeyTouched && form.apiKey) setApiKeySet(true)
      setApiKeyTouched(false)
      toast.success(t('socAi.saved'))
    } catch (err) {
      // The backend runs a live connection + tool-calling check; its 400 message
      // (bad key, unreachable, no tool-calling) comes through ApiError.message.
      toast.error(err instanceof Error ? err.message : t('socAi.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const showApiKey = def.apiKey !== 'none' && !def.custom
  const hasModelList = def.models.length > 0

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <Bot size={16} strokeWidth={1.75} />
          {t('socAi.title')}
        </h1>
      </header>

      {loading ? (
        <div className="mt-6 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : (
        <div className="mt-6 space-y-5">
          {/* Provider & model */}
          <Section title={t('socAi.section.connection')}>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('socAi.provider.label')} hint={t('socAi.provider.hint')}>
                <Select value={form.provider} onChange={(e) => changeProvider(e.target.value)}>
                  {PROVIDER_IDS.map((id) => (
                    <option key={id} value={id}>
                      {PROVIDERS[id].label}
                    </option>
                  ))}
                </Select>
              </Field>

              {!def.fixedModel && (
                <Field label={t('socAi.model.label')} hint={t('socAi.model.hint')}>
                  {hasModelList ? (
                    <div className="space-y-2">
                      <Select
                        value={modelCustom ? CUSTOM_MODEL : form.model}
                        onChange={(e) => {
                          if (e.target.value === CUSTOM_MODEL) {
                            setModelCustom(true)
                            patch({ model: '' })
                          } else {
                            setModelCustom(false)
                            patch({ model: e.target.value })
                          }
                        }}
                      >
                        {def.models.map((m) => (
                          <option key={m} value={m}>
                            {m}
                          </option>
                        ))}
                        <option value={CUSTOM_MODEL}>{t('socAi.model.customOption')}</option>
                      </Select>
                      {modelCustom && (
                        <Input
                          value={form.model}
                          onChange={(e) => patch({ model: e.target.value })}
                          placeholder={t('socAi.model.customPlaceholder')}
                        />
                      )}
                    </div>
                  ) : (
                    <Input
                      value={form.model}
                      onChange={(e) => patch({ model: e.target.value })}
                      placeholder={t('socAi.model.customPlaceholder')}
                    />
                  )}
                </Field>
              )}

              {def.showUrl && (
                <Field label={t('socAi.url.label')} hint={t('socAi.url.hint')}>
                  <Input value={form.url} onChange={(e) => patch({ url: e.target.value })} placeholder={def.urlPlaceholder} />
                </Field>
              )}

              {showApiKey && (
                <SecretField
                  label={t('socAi.apiKey.label')}
                  hint={apiKeySet && !apiKeyTouched ? t('socAi.apiKey.setHint') : t('socAi.apiKey.hint')}
                  value={form.apiKey}
                  showKey={showKey}
                  onToggleShow={() => setShowKey((v) => !v)}
                  onChange={(v) => {
                    setForm((f) => ({ ...f, apiKey: v }))
                    setApiKeyTouched(true)
                  }}
                  placeholder={apiKeySet ? '••••••••' : ''}
                />
              )}
            </div>

            {def.note && <p className="mt-4 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">{t(def.note)}</p>}
          </Section>

          {/* Usage — only for the ThreatWinds provider, backed by CM's proxy quota. */}
          {form.provider === 'threatwinds' && <ThreatWindsUsageSection />}

          {/* Authentication — only for a fully custom gateway. */}
          {def.custom && (
            <Section title={t('socAi.section.auth')}>
              <Field label={t('socAi.authType.label')} hint={t('socAi.authType.hint')}>
                <Select
                  value={form.authKind}
                  onChange={(e) => patch({ authKind: e.target.value as AuthKind })}
                  className="max-w-xs"
                >
                  <option value="apiKey">{t('socAi.authType.options.apiKey')}</option>
                  <option value="bearer">{t('socAi.authType.options.bearer')}</option>
                  <option value="customHeaders">{t('socAi.authType.options.customHeaders')}</option>
                  <option value="none">{t('socAi.authType.options.none')}</option>
                </Select>
              </Field>

              {form.authKind === 'apiKey' && (
                <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <Field label={t('socAi.authHeaderName.label')} hint={t('socAi.authHeaderName.hint')}>
                    <Input value={form.authHeaderName} onChange={(e) => patch({ authHeaderName: e.target.value })} placeholder="api-key" />
                  </Field>
                  <SecretField
                    label={t('socAi.apiKey.label')}
                    hint={apiKeySet && !apiKeyTouched ? t('socAi.apiKey.setHint') : t('socAi.apiKey.hint')}
                    value={form.apiKey}
                    showKey={showKey}
                    onToggleShow={() => setShowKey((v) => !v)}
                    onChange={(v) => {
                      setForm((f) => ({ ...f, apiKey: v }))
                      setApiKeyTouched(true)
                    }}
                    placeholder={apiKeySet ? '••••••••' : ''}
                  />
                </div>
              )}

              {form.authKind === 'bearer' && (
                <div className="mt-4 max-w-sm">
                  <SecretField
                    label={t('socAi.authToken.label')}
                    hint={apiKeySet && !apiKeyTouched ? t('socAi.authToken.setHint') : t('socAi.authToken.hint')}
                    value={form.apiKey}
                    showKey={showKey}
                    onToggleShow={() => setShowKey((v) => !v)}
                    onChange={(v) => {
                      setForm((f) => ({ ...f, apiKey: v }))
                      setApiKeyTouched(true)
                    }}
                    placeholder={apiKeySet ? '••••••••' : ''}
                  />
                </div>
              )}

              {form.authKind === 'customHeaders' && (
                <div className="mt-4">
                  <Field label={t('socAi.customHeaders.label')} hint={t('socAi.customHeaders.hint')}>
                    <CustomHeadersTable
                      rows={form.customHeadersList}
                      onChange={(rows) => patch({ customHeadersList: rows })}
                    />
                  </Field>
                </div>
              )}
            </Section>
          )}

          {/* Triage on/off */}
          <Section title={t('socAi.section.behavior')}>
            <ToggleRow
              label={t('socAi.autoAnalyze.label')}
              hint={t('socAi.autoAnalyze.hint')}
              checked={form.autoAnalyze}
              onChange={(v) => patch({ autoAnalyze: v })}
            />
          </Section>

          {/* Capabilities — what the agent is allowed to do (read is always on). */}
          <Section title={t('socAi.section.capabilities')}>
            <p className="-mt-2 mb-3 text-xs text-muted-foreground">{t('socAi.capsCaption')}</p>
            <div className="space-y-1">
              {CAPABILITY_GROUPS.map((g) => (
                <ToggleRow
                  key={g.id}
                  label={t(`socAi.cap.${g.id}.label`)}
                  hint={t(`socAi.cap.${g.id}.hint`)}
                  danger={g.danger}
                  checked={form.capabilities.includes(g.id)}
                  onChange={(v) =>
                    patch({
                      capabilities: v
                        ? [...form.capabilities, g.id]
                        : form.capabilities.filter((c) => c !== g.id),
                    })
                  }
                />
              ))}
            </div>
          </Section>

          {/* Advanced */}
          <Section title={t('socAi.section.advanced')}>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field label={t('socAi.maxTokens.label')} hint={t('socAi.maxTokens.hint')}>
                <Input type="number" value={form.maxTokens} onChange={(e) => patch({ maxTokens: e.target.value })} className="max-w-[140px]" />
              </Field>
              <Field label={t('socAi.maxToolIterations.label')} hint={t('socAi.maxToolIterations.hint')}>
                <Input
                  type="number"
                  value={form.maxToolIterations}
                  onChange={(e) => patch({ maxToolIterations: e.target.value })}
                  className="max-w-[140px]"
                />
              </Field>
            </div>
          </Section>

          <div className="flex justify-end">
            <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
              {saving ? t('socAi.saving') : t('socAi.save')}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

/* ─── ThreatWinds usage ─────────────────────────────────────────────────── */

function ThreatWindsUsageSection() {
  const { t } = useTranslation()
  const { data, isLoading } = useTiUsage()

  const chat = data?.kind === 'ok' ? data.value[THREATWINDS_CHAT_USAGE_KEY] : undefined

  return (
    <Section title={t('socAi.usage.title')}>
      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
        </div>
      ) : !chat ? (
        <p className="text-xs text-muted-foreground">{t('socAi.usage.unavailable')}</p>
      ) : (
        <div className="max-w-sm space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span>{t('socAi.usage.chat')}</span>
            <span className="font-mono text-xs text-muted-foreground">
              {chat.used} / {chat.limit}
            </span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className={cn(
                'h-full rounded-full transition-all',
                chat.used >= chat.limit ? 'bg-destructive' : 'bg-primary',
              )}
              style={{ width: `${chat.limit > 0 ? Math.min(100, (chat.used / chat.limit) * 100) : 0}%` }}
            />
          </div>
          {chat.resets_in_seconds > 0 && (
            <p className="text-[11px] text-muted-foreground">
              {t('socAi.usage.resetsIn', { time: formatDuration(chat.resets_in_seconds) })}
            </p>
          )}
        </div>
      )}
    </Section>
  )
}

function formatDuration(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <h2 className="mb-4 text-sm font-semibold">{title}</h2>
      {children}
    </section>
  )
}

function Field({ label, hint, children }: { label?: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      {label && <label className="block text-xs font-medium text-foreground/80">{label}</label>}
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Select(props: React.SelectHTMLAttributes<HTMLSelectElement>) {
  const { className, ...rest } = props
  return (
    <select
      {...rest}
      className={cn(
        'h-9 w-full rounded-md border border-input bg-background px-3 text-sm',
        'focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
        className,
      )}
    />
  )
}

function SecretField({
  label,
  hint,
  value,
  showKey,
  onToggleShow,
  onChange,
  placeholder,
}: {
  label: string
  hint?: string
  value: string
  showKey: boolean
  onToggleShow: () => void
  onChange: (v: string) => void
  placeholder?: string
}) {
  return (
    <Field label={label} hint={hint}>
      <div className="relative max-w-sm">
        <Input
          type={showKey ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          autoComplete="new-password"
          className="pr-9"
        />
        <button
          type="button"
          onClick={onToggleShow}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          {showKey ? <EyeOff size={14} /> : <Eye size={14} />}
        </button>
      </div>
    </Field>
  )
}

function CustomHeadersTable({
  rows,
  onChange,
}: {
  rows: HeaderRow[]
  onChange: (rows: HeaderRow[]) => void
}) {
  const { t } = useTranslation()
  const update = (i: number, patch: Partial<HeaderRow>) =>
    onChange(rows.map((r, idx) => (idx === i ? { ...r, ...patch } : r)))
  const remove = (i: number) => onChange(rows.filter((_, idx) => idx !== i))
  const add = () => onChange([...rows, { key: '', value: '' }])

  return (
    <div className="space-y-2">
      {rows.map((row, i) => (
        <div key={i} className="flex items-center gap-2">
          <Input
            value={row.key}
            onChange={(e) => update(i, { key: e.target.value })}
            placeholder={t('socAi.customHeaders.keyPlaceholder')}
            className="max-w-[220px] font-mono text-xs"
          />
          <Input
            value={row.value}
            onChange={(e) => update(i, { value: e.target.value })}
            placeholder={t('socAi.customHeaders.valuePlaceholder')}
            className="font-mono text-xs"
          />
          <button
            type="button"
            onClick={() => remove(i)}
            aria-label={t('socAi.customHeaders.remove')}
            className="shrink-0 rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-destructive"
          >
            <Trash2 size={14} />
          </button>
        </div>
      ))}
      <Button type="button" variant="outline" size="sm" onClick={add}>
        <Plus size={14} className="mr-1.5" />
        {t('socAi.customHeaders.addRow')}
      </Button>
    </div>
  )
}

function ToggleRow({
  label,
  hint,
  checked,
  danger,
  onChange,
}: {
  label: string
  hint?: string
  checked: boolean
  danger?: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-border py-3 last:border-0">
      <div className="min-w-0">
        <div className="flex items-center gap-1.5 text-sm font-medium">
          {label}
          {danger && (
            <span className="rounded bg-amber-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase tracking-wide text-amber-600 dark:text-amber-400">
              ⚠
            </span>
          )}
        </div>
        {hint && <p className="mt-0.5 text-xs text-muted-foreground">{hint}</p>}
      </div>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={cn(
          'relative mt-0.5 inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors',
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
    </div>
  )
}
