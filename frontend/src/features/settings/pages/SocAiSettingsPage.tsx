import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bot, Eye, EyeOff, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { socaiHttpService } from '../services/socai-http.service'

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
}

const CUSTOM_MODEL = '__custom__'

const PROVIDERS: Record<string, ProviderDef> = {
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

interface Form {
  provider: string
  model: string
  url: string
  apiKey: string
  authType: string
  authHeaderName: string
  customHeaders: string // JSON text (custom provider only)
  maxTokens: string
  maxToolIterations: string
  autoAnalyze: boolean
  capabilities: string[] // enabled permission group ids
}

function emptyForm(provider = 'openai'): Form {
  const def = PROVIDERS[provider]
  return {
    provider,
    model: def.models[0] ?? '',
    url: '',
    apiKey: '',
    authType: def.authType,
    authHeaderName: def.authHeaderName ?? '',
    customHeaders: '{}',
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
        const provider = cfg.provider && PROVIDERS[cfg.provider] ? cfg.provider : 'openai'
        const def = PROVIDERS[provider]
        const v: Form = {
          provider,
          model: cfg.model || def.models[0] || '',
          url: cfg.url || '',
          apiKey: '',
          authType: cfg.authType || def.authType,
          authHeaderName: cfg.authHeaderName || def.authHeaderName || '',
          customHeaders: JSON.stringify(cfg.customHeaders ?? {}, null, 2),
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
      model: d.models[0] ?? '',
      url: '',
      authType: d.authType,
      authHeaderName: d.authHeaderName ?? '',
      customHeaders: '{}',
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
    let customHeaders: Record<string, string> = {}
    if (def.custom) {
      try {
        const parsed = JSON.parse(form.customHeaders || '{}')
        if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) throw new Error('not an object')
        customHeaders = parsed as Record<string, string>
      } catch {
        toast.error(t('socAi.invalidHeaders'))
        return
      }
    }

    setSaving(true)
    try {
      await socaiHttpService.update({
        provider: form.provider,
        model: form.model.trim(),
        url: def.showUrl ? form.url.trim() : '',
        apiKey: def.apiKey === 'none' ? '' : apiKeyTouched ? form.apiKey : '',
        authType: def.custom ? form.authType : def.authType,
        authHeaderName: def.custom ? form.authHeaderName.trim() : (def.authHeaderName ?? ''),
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

  const showApiKey = def.apiKey !== 'none'
  const hasModelList = def.models.length > 0

  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 pb-6 pt-3">
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

              {def.showUrl && (
                <Field label={t('socAi.url.label')} hint={t('socAi.url.hint')}>
                  <Input value={form.url} onChange={(e) => patch({ url: e.target.value })} placeholder={def.urlPlaceholder} />
                </Field>
              )}

              {showApiKey && (
                <Field
                  label={t('socAi.apiKey.label')}
                  hint={apiKeySet && !apiKeyTouched ? t('socAi.apiKey.setHint') : t('socAi.apiKey.hint')}
                >
                  <div className="relative max-w-sm">
                    <Input
                      type={showKey ? 'text' : 'password'}
                      value={form.apiKey}
                      onChange={(e) => {
                        setForm((f) => ({ ...f, apiKey: e.target.value }))
                        setApiKeyTouched(true)
                      }}
                      placeholder={apiKeySet ? '••••••••' : ''}
                      autoComplete="new-password"
                      className="pr-9"
                    />
                    <button
                      type="button"
                      onClick={() => setShowKey((v) => !v)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                    >
                      {showKey ? <EyeOff size={14} /> : <Eye size={14} />}
                    </button>
                  </div>
                </Field>
              )}
            </div>

            {def.note && <p className="mt-4 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">{t(def.note)}</p>}

            {/* Advanced auth — only for a fully custom gateway. */}
            {def.custom && (
              <div className="mt-4 grid grid-cols-1 gap-4 border-t border-border pt-4 sm:grid-cols-2">
                <Field label={t('socAi.authType.label')} hint={t('socAi.authType.hint')}>
                  <Select value={form.authType} onChange={(e) => patch({ authType: e.target.value })}>
                    {['bearer', 'header', 'none'].map((a) => (
                      <option key={a} value={a}>
                        {a}
                      </option>
                    ))}
                  </Select>
                </Field>
                {form.authType === 'header' && (
                  <Field label={t('socAi.authHeaderName.label')} hint={t('socAi.authHeaderName.hint')}>
                    <Input value={form.authHeaderName} onChange={(e) => patch({ authHeaderName: e.target.value })} placeholder="api-key" />
                  </Field>
                )}
                <div className="sm:col-span-2">
                  <Field label={t('socAi.customHeaders.label')} hint={t('socAi.customHeaders.hint')}>
                    <Textarea
                      value={form.customHeaders}
                      onChange={(e) => patch({ customHeaders: e.target.value })}
                      rows={4}
                      spellCheck={false}
                      placeholder={'{ "api-secret": "..." }'}
                    />
                  </Field>
                </div>
              </div>
            )}
          </Section>

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

function Textarea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const { className, ...rest } = props
  return (
    <textarea
      {...rest}
      className={cn(
        'flex w-full rounded-md border border-input bg-background/40 px-3 py-2 font-mono text-[11px] leading-relaxed',
        'placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
        className,
      )}
    />
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
