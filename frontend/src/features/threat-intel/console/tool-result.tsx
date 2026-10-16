import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  CircleSlash,
  Crosshair,
  FileText,
  MapPin,
  Network,
  Search,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { AgentStep, IntelEntity } from './types'
import { RelationGraph } from './relation-graph'

export interface ToolResultProps {
  step: AgentStep
  onPivot(target: { id: string; type: string; value: string }): void
}

const TOOL_ICONS: Record<string, typeof Search> = {
  lookup_indicator: Crosshair,
  search_entities: Search,
  get_entity_details: FileText,
  get_relations: Network,
  get_geolocations: MapPin,
}

function reputationVariant(reputation: number | null) {
  return reputation !== null && reputation < 0
    ? 'border-destructive/40 bg-destructive/10 text-destructive'
    : 'border-border/50 text-muted-foreground'
}

function Badge({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span className={cn('shrink-0 rounded-sm border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide', className)}>
      {children}
    </span>
  )
}

function EntityRow({
  entity,
  onPivot,
}: {
  entity: IntelEntity
  onPivot(target: { id: string; type: string; value: string }): void
}) {
  return (
    <button
      type="button"
      onClick={() => onPivot(entity)}
      className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-start transition-colors hover:bg-muted/40"
    >
      <Badge className={reputationVariant(entity.reputation)}>{entity.type}</Badge>
      <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground/90">{entity.value}</span>
      <span className="shrink-0 font-mono text-[10px] text-muted-foreground">rep {entity.reputation}</span>
    </button>
  )
}

function shorten(s: string, n: number) {
  return s.length > n ? `${s.slice(0, n - 1)}…` : s
}

export function ToolResult({ step, onPivot }: ToolResultProps) {
  const { t } = useTranslation()
  const Icon = TOOL_ICONS[step.toolName] ?? Search

  return (
    <div className="overflow-hidden rounded-sm border border-border/50 bg-muted/20">
      <div className="flex items-center gap-2 border-b border-border/30 bg-muted/40 px-3 py-2">
        <Icon className="h-3.5 w-3.5 shrink-0 text-primary/70" />
        <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          {step.toolName.replace(/_/g, ' ')}
        </span>
        <span className="min-w-0 flex-1 truncate font-mono text-[10px] text-muted-foreground/60">
          {Object.entries(step.args)
            .map(([k, v]) => `${k}=${String(v)}`)
            .join(' ')}
        </span>
        {step.status === 'running' && (
          <span className="shrink-0 animate-pulse font-mono text-[10px] uppercase text-primary">{t('tiAnalyst.tool.querying')}</span>
        )}
        {step.durationMs !== undefined && step.status !== 'running' && (
          <span className="shrink-0 font-mono text-[10px] text-muted-foreground/50">{Math.round(step.durationMs)}ms</span>
        )}
      </div>

      <div className="p-3">
        {step.status === 'error' ? (
          <div className="flex items-start gap-2 font-mono text-xs text-destructive">
            <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
            <span>{step.error}</span>
          </div>
        ) : step.payload ? (
          <PayloadView payload={step.payload} onPivot={onPivot} />
        ) : (
          <span className="block whitespace-pre-wrap break-words font-mono text-[10px] text-muted-foreground">
            {step.status === 'running' ? t('tiAnalyst.tool.querying') : (step.result ?? t('tiAnalyst.tool.notRetained'))}
          </span>
        )}
      </div>
    </div>
  )
}

function PayloadView({
  payload,
  onPivot,
}: {
  payload: NonNullable<AgentStep['payload']>
  onPivot: ToolResultProps['onPivot']
}) {
  const { t } = useTranslation()
  switch (payload.kind) {
    case 'lookup':
      return payload.entity ? (
        <EntityRow entity={payload.entity} onPivot={onPivot} />
      ) : (
        <div className="flex items-center gap-2 font-mono text-xs text-muted-foreground">
          <CircleSlash className="h-3.5 w-3.5 shrink-0" />
          <span>{t('tiAnalyst.tool.noIntel', { value: payload.value })}</span>
        </div>
      )

    case 'search':
      return (
        <div className="space-y-1">
          <div className="mb-1.5 font-mono text-[10px] text-muted-foreground">
            {t('tiAnalyst.tool.matches', { items: payload.items, query: payload.query })}
          </div>
          {payload.results.length === 0 && (
            <div className="font-mono text-xs text-muted-foreground">{t('tiAnalyst.tool.noMatches')}</div>
          )}
          {payload.results.slice(0, 10).map((entity) => (
            <button
              key={entity.id}
              type="button"
              onClick={() => onPivot(entity)}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-start transition-colors hover:bg-muted/40"
            >
              <Badge className={reputationVariant(entity.reputation)}>{entity.type}</Badge>
              <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground/90">{shorten(entity.value, 72)}</span>
              <span className="shrink-0 font-mono text-[10px] text-muted-foreground">rep {entity.reputation}</span>
            </button>
          ))}
          {payload.results.length > 10 && (
            <div className="pt-1 font-mono text-[10px] text-muted-foreground">
              {t('tiAnalyst.tool.more', { count: payload.results.length - 10 })}
            </div>
          )}
        </div>
      )

    case 'details': {
      const a = payload.details.attributes
      return (
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Badge className={a.reputation_score < 0 ? 'border-destructive/40 bg-destructive/10 text-destructive' : undefined}>
              {a.type}
            </Badge>
            <span className="truncate font-mono text-sm text-foreground">{a.value}</span>
          </div>
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <Metric label={t('tiAnalyst.entity.reputation')} value={a.reputation} score={a.reputation_score} />
            <Metric label={t('tiAnalyst.entity.accuracy')} value={a.accuracy} score={a.accuracy_score} />
            <Metric label={t('tiAnalyst.entity.worst')} value={a.worst_reputation} score={a.worst_reputation_score} />
            <Metric label={t('tiAnalyst.entity.best')} value={a.best_reputation} score={a.best_reputation_score} />
          </div>
          {payload.details.metadata && (
            <div className="flex flex-wrap gap-1.5 pt-1">
              {Object.entries(payload.details.metadata).map(([k, v]) => (
                <span
                  key={k}
                  className="rounded-sm border border-border/40 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground"
                >
                  {k}: <span className="text-foreground/80">{String(v)}</span>
                </span>
              ))}
            </div>
          )}
          {payload.details.latest_associations.length > 0 && (
            <div className="space-y-1 pt-1">
              <div className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
                {t('tiAnalyst.entity.associations')}
              </div>
              {payload.details.latest_associations.slice(0, 6).map((assoc) => (
                <button
                  key={assoc.id}
                  type="button"
                  onClick={() => onPivot(assoc)}
                  className="flex w-full items-center gap-2 rounded-sm px-2 py-1 text-start transition-colors hover:bg-muted/40"
                >
                  <Badge className={reputationVariant(assoc.reputation)}>{assoc.type}</Badge>
                  <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground/90">{assoc.value}</span>
                </button>
              ))}
            </div>
          )}
        </div>
      )
    }

    case 'relations':
      return (
        <RelationGraph
          graph={payload.graph}
          rootId={payload.id}
          height={340}
          onSelect={(node) => onPivot({ id: node.id, type: node.type, value: node.value })}
        />
      )

    case 'geo':
      return payload.locations.length === 0 ? (
        <div className="font-mono text-xs text-muted-foreground">{t('tiAnalyst.canvas.noGeo')}</div>
      ) : (
        <div className="space-y-1">
          {payload.locations.map((geo, i) => (
            <div key={`${geo.lat}-${geo.lng}-${i}`} className="flex items-center gap-2 rounded-sm px-2 py-1.5">
              <MapPin className="h-3.5 w-3.5 shrink-0 text-primary/60" />
              <span className="font-mono text-xs text-foreground/90">
                {geo.lat.toFixed(4)}°, {geo.lng.toFixed(4)}°
              </span>
              <span className="font-mono text-[10px] text-muted-foreground">
                {t('tiAnalyst.entity.radius', { radius: geo.radius })}
              </span>
            </div>
          ))}
        </div>
      )
  }
}

function Metric({ label, value, score }: { label: string; value: string; score: number }) {
  return (
    <div className="rounded-sm border border-border/40 bg-card px-2.5 py-2">
      <div className="font-mono text-[9px] uppercase tracking-widest text-muted-foreground">{label}</div>
      <div className="truncate font-mono text-xs text-foreground/90">{value || '—'}</div>
      <div className="font-mono text-[10px] text-muted-foreground">{score}</div>
    </div>
  )
}
