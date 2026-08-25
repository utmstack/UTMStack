import { useEffect, useMemo, useState } from 'react'
import { Check, ChevronDown, Plus, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { COMMON_PLATFORMS, defaultShellForPlatform, shellsForPlatform } from '../lib/alert-fields'

const SELECT = 'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface AgentOption {
  name: string
  platform: string
}

interface Props {
  platform?: string
  agent?: string
  excludedAgents?: string[]
  shell?: string
  readOnly?: boolean
  onChange: (patch: { platform?: string; agent?: string; excludedAgents?: string[]; shell?: string }) => void
}

type Scope = 'auto' | 'specific' | 'except'

function scopeFrom(agent?: string, excluded?: string[]): Scope {
  if (agent) return 'specific'
  if (excluded && excluded.length) return 'except'
  return 'auto'
}

/** Per-node platform + agent + shell picker for `shell` nodes. Three modes:
 *
 *   - auto: run on `$(alert.dataSource)` (the box that raised the alert)
 *   - specific: pick one hostname (or a template)
 *   - except: run auto, but skip if the alert source is on the exclusion list
 *
 *  Platform is a UI-only filter (limits the agent list); the runtime uses
 *  `agent` + `excludedAgents`. */
export function AgentPicker({ platform, agent, excludedAgents, shell, readOnly, onChange }: Props) {
  const [agents, setAgents] = useState<AgentOption[]>([])
  const [scope, setScope] = useState<Scope>(() => scopeFrom(agent, excludedAgents))

  // Keep the tab in sync when the caller changes agent/excluded externally.
  useEffect(() => {
    setScope(scopeFrom(agent, excludedAgents))
  }, [agent, excludedAgents])

  useEffect(() => {
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: 'agent' })
      .then((r) =>
        setAgents(
          (r.items ?? []).map((d) => ({
            name: d.name,
            platform: typeof d.metadata?.osPlatform === 'string' ? d.metadata.osPlatform : '',
          })),
        ),
      )
      .catch(() => setAgents([]))
  }, [])

  const platforms = useMemo(() => {
    const seen = new Set<string>()
    const out: string[] = []
    for (const p of [...COMMON_PLATFORMS, ...agents.map((a) => a.platform), platform ?? '']) {
      const v = p.trim()
      if (v && !seen.has(v.toLowerCase())) {
        seen.add(v.toLowerCase())
        out.push(v)
      }
    }
    return out
  }, [agents, platform])

  const platformAgents = useMemo(() => {
    const p = (platform ?? '').trim().toLowerCase()
    if (!p) return agents
    const m = agents.filter((a) => a.platform && (a.platform.toLowerCase().includes(p) || p.includes(a.platform.toLowerCase())))
    return m.length ? m : agents
  }, [agents, platform])

  const shells = shellsForPlatform(platform ?? '')

  const onPlatformChange = (v: string) => {
    const nextShell = shells.includes(shell ?? '') ? shell : defaultShellForPlatform(v)
    onChange({ platform: v || undefined, shell: nextShell || undefined })
  }

  const switchScope = (next: Scope) => {
    if (next === scope) return
    setScope(next)
    if (next === 'auto') onChange({ agent: undefined, excludedAgents: undefined })
    else if (next === 'specific') onChange({ agent: undefined, excludedAgents: undefined })
    else onChange({ agent: undefined, excludedAgents: excludedAgents ?? [] })
  }

  const options = platformAgents.map((a) => a.name)

  return (
    <div className="space-y-2">
      <div className="grid grid-cols-2 gap-2">
        <Field label="Platform">
          <select value={platform ?? ''} disabled={readOnly} onChange={(e) => onPlatformChange(e.target.value)} className={SELECT + ' w-full'}>
            <option value="">any</option>
            {platforms.map((p) => (
              <option key={p} value={p}>{p}</option>
            ))}
          </select>
        </Field>
        <Field label="Shell">
          <select value={shell ?? ''} disabled={readOnly} onChange={(e) => onChange({ shell: e.target.value || undefined })} className={SELECT + ' w-full'}>
            <option value="">auto</option>
            {shells.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </Field>
      </div>

      <Field label="Run on">
        <div className="inline-flex w-full rounded-md border border-border p-0.5">
          {(['auto', 'specific', 'except'] as const).map((s) => (
            <button
              key={s}
              type="button"
              disabled={readOnly}
              onClick={() => switchScope(s)}
              className={cn(
                'flex-1 rounded px-2 py-1 text-[11px] capitalize transition-colors disabled:cursor-not-allowed',
                scope === s ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
              title={
                s === 'auto'
                  ? 'Runs on the host that raised the alert'
                  : s === 'specific'
                  ? 'Pin to one host (or a template like $(alert.foo))'
                  : 'Runs on the host that raised the alert, but skips ones in the list'
              }
            >
              {s === 'auto' ? 'alert source' : s === 'specific' ? 'one host' : 'all except'}
            </button>
          ))}
        </div>
      </Field>

      {scope === 'auto' && (
        <p className="rounded-md bg-muted/40 px-2 py-1.5 text-[10px] text-muted-foreground">
          Runs on <code className="rounded bg-background px-1 font-mono">$(alert.dataSource)</code> — the host that
          raised the matched alert.
        </p>
      )}

      {scope === 'specific' && (
        <div className="space-y-1.5">
          <select
            value={options.includes(agent ?? '') ? agent : ''}
            disabled={readOnly}
            onChange={(e) => onChange({ agent: e.target.value || undefined, excludedAgents: undefined })}
            className={SELECT + ' w-full'}
          >
            <option value="">pick a host…</option>
            {options.map((name) => (
              <option key={name} value={name}>{name}</option>
            ))}
            {agent && !options.includes(agent) && <option value={agent}>{agent}</option>}
          </select>
          <p className="text-[10px] text-muted-foreground">
            Or type a template — <code className="rounded bg-muted px-1 font-mono">$(alert.field)</code> resolves per
            execution.
          </p>
          <Input
            value={agent ?? ''}
            readOnly={readOnly}
            onChange={(e) => onChange({ agent: e.target.value || undefined, excludedAgents: undefined })}
            placeholder="$(alert.dataSource)"
            className="h-8 font-mono text-xs"
          />
        </div>
      )}

      {scope === 'except' && (
        <AgentMultiSelect
          options={options}
          values={excludedAgents ?? []}
          readOnly={readOnly}
          onChange={(v) => onChange({ agent: undefined, excludedAgents: v.length ? v : undefined })}
        />
      )}
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{label}</label>
      {children}
    </div>
  )
}

/** Searchable multi-select of agent hostnames. Used for the exclusion list on
 *  the shell node. Kept local so it doesn't accidentally bleed into other
 *  parts of the UI. */
function AgentMultiSelect({
  options,
  values,
  readOnly,
  onChange,
}: {
  options: string[]
  values: string[]
  readOnly?: boolean
  onChange: (v: string[]) => void
}) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')

  const toggle = (name: string) => onChange(values.includes(name) ? values.filter((x) => x !== name) : [...values, name])
  const filtered = options.filter((o) => (q ? o.toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div className="relative">
      <div className="flex min-h-9 flex-wrap items-center gap-1 rounded-md border border-input bg-background/40 p-1.5">
        {values.map((v) => (
          <span key={v} className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">
            {v}
            {!readOnly && (
              <button type="button" onClick={() => toggle(v)} className="text-muted-foreground hover:text-red-500">
                <X size={11} />
              </button>
            )}
          </span>
        ))}
        {!readOnly && (
          <button type="button" onClick={() => setOpen((o) => !o)} className="inline-flex items-center gap-1 px-1 text-[11px] text-muted-foreground hover:text-foreground">
            <Plus size={11} /> {values.length ? '' : 'add exclusion'}
            <ChevronDown size={10} className="opacity-60" />
          </button>
        )}
      </div>
      {open && !readOnly && (
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input
              value={q}
              onChange={(e) => setQ(e.target.value)}
              autoFocus
              placeholder="search agents…"
              className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>
          <div className="max-h-52 overflow-y-auto">
            {filtered.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">no agents</div>}
            {filtered.map((o) => {
              const on = values.includes(o)
              return (
                <button key={o} type="button" onClick={() => toggle(o)} className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted">
                  <span className="min-w-0 flex-1 truncate font-mono">{o}</span>
                  {on && <Check size={13} className="shrink-0 text-primary" />}
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
