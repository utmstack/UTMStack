import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Lock, Search, Server } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { AgentConsole } from '@/features/datasources/components/AgentConsole'
import { deriveStatus, STATUS_META } from '@/features/datasources/lib/status'
import type { Datasource } from '@/features/datasources/types/datasource.types'

interface AgentItem {
  id: string // the datasource uuid — React key and selection identity
  agentId: string
  name: string
  osPlatform?: string
  lastPingAt?: string
  noRemoteControl: boolean
}

/** Only agent-kind datasources linked to a real agent record can open a console. */
function toAgentItem(d: Datasource): AgentItem | null {
  const agentId = d.metadata?.agentId
  if (d.sourceKind !== 'agent' || agentId == null) return null
  return {
    id: d.id,
    agentId: String(agentId),
    name: d.name,
    osPlatform: typeof d.metadata?.osPlatform === 'string' ? d.metadata.osPlatform : undefined,
    lastPingAt: d.lastPingAt,
    // The agent-manager writes it as a string, and an agent registered before
    // the column existed does not write it at all — both mean "accepts commands".
    noRemoteControl: d.metadata?.noRemoteControl === 'true',
  }
}

export function InteractiveConsolePage() {
  const { t } = useTranslation()
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<AgentItem | null>(null)

  useEffect(() => {
    setLoading(true)
    setError(false)
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: 'agent', sort: 'asset_name.asc' })
      .then((r) => setAgents((r.items ?? []).map(toAgentItem).filter((a): a is AgentItem => a !== null)))
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return agents
    return agents.filter((a) => a.name.toLowerCase().includes(q))
  }, [agents, search])

  const selectedStatus = selected ? deriveStatus(selected.lastPingAt) : null

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-20 pt-3 ">
      <div className="flex min-h-0 flex-1 gap-4">
        <aside className="flex w-72 shrink-0 flex-col rounded-xl border border-border bg-card">
          <div className="border-b border-border p-3">
            <div className="relative">
              <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t('soar.editor.searchAgents')}
                className="pl-8"
              />
            </div>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto">
            {loading ? (
              <ListCenter><Loader2 className="h-4 w-4 animate-spin" /> {t('soar.console.loading')}</ListCenter>
            ) : error ? (
              <ListCenter><AlertTriangle size={14} className="text-amber-500" /> {t('soar.console.loadError')}</ListCenter>
            ) : filtered.length === 0 ? (
              <ListCenter>{t('soar.editor.noAgents')}</ListCenter>
            ) : (
              filtered.map((a) => (
                <AgentRow key={a.id} agent={a} active={selected?.id === a.id} onClick={() => setSelected(a)} />
              ))
            )}
          </div>
        </aside>

        {selected && selectedStatus ? (
          <div className="flex min-h-0 flex-1 flex-col">
            <AgentConsole
              key={selected.agentId}
              variant="inline"
              agentId={selected.agentId}
              hostname={selected.name}
              osPlatform={selected.osPlatform}
              offline={selectedStatus === 'offline'}
              noRemoteControl={selected.noRemoteControl}
              onClose={() => setSelected(null)}
            />
          </div>
        ) : (
          <div className="flex min-h-0 flex-1 flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card px-6 text-center">
            <Server size={28} strokeWidth={1.5} className="text-muted-foreground/60" />
            <div className="mt-3 text-sm font-medium">{t('soar.console.pickAgent')}</div>
            <p className="mt-1 max-w-sm text-xs text-muted-foreground">{t('soar.console.pickAgentHint')}</p>
          </div>
        )}
      </div>
    </div>
  )
}

function AgentRow({ agent, active, onClick }: { agent: AgentItem; active: boolean; onClick: () => void }) {
  const { t } = useTranslation()
  const status = deriveStatus(agent.lastPingAt)
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-2 border-b border-border/50 px-3 py-2 text-left text-[13px] transition-colors last:border-b-0 hover:bg-muted/40',
        active && 'bg-primary/10',
      )}
    >
      <span className={cn('h-1.5 w-1.5 shrink-0 rounded-full', STATUS_META[status].dot)} />
      <span className="min-w-0 flex-1 truncate">{agent.name}</span>
      {agent.noRemoteControl && (
        <Lock size={11} className="shrink-0 text-muted-foreground" aria-label={t('soar.console.lockedAgent')} />
      )}
      {agent.osPlatform && (
        <span className="shrink-0 truncate rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
          {agent.osPlatform}
        </span>
      )}
    </button>
  )
}

function ListCenter({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-4 py-10 text-center text-xs text-muted-foreground">{children}</div>
}
