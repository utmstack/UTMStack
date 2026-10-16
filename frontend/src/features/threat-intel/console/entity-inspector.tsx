import { useQuery } from '@tanstack/react-query'
import { MapPin, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { proxyFetch } from './proxy'
import { normalizeDetails, relationsToGraph } from './agent'
import type { EntityDetails } from './types'
import type { RawDetails, RawRelation } from './agent'
import { RelationGraph } from './relation-graph'

interface EntityInspectorProps {
  entityId: string
  onPivot(target: { id: string; type: string; value: string }): void
  onAskAgent(question: string): void
}

export function EntityInspector({ entityId, onPivot, onAskAgent }: EntityInspectorProps) {
  const { t } = useTranslation()

  const details = useQuery({
    queryKey: ['ti-analyst-details', entityId],
    queryFn: async () => normalizeDetails(await proxyFetch<RawDetails>(`/entity/${encodeURIComponent(entityId)}`)),
    staleTime: 10 * 60 * 1000,
  })

  const relations = useQuery({
    queryKey: ['ti-analyst-relations', entityId],
    queryFn: async () => relationsToGraph(entityId, '', (await proxyFetch<RawRelation[]>(`/entity/${encodeURIComponent(entityId)}/relations`)) ?? []),
    staleTime: 10 * 60 * 1000,
  })

  if (details.isLoading) {
    return (
      <div className="space-y-2 p-4">
        <div className="h-24 animate-pulse rounded bg-muted/50" />
        <div className="h-56 animate-pulse rounded bg-muted/50" />
      </div>
    )
  }

  const d: EntityDetails | undefined = details.data
  if (details.isError || !d) {
    return (
      <div className="flex h-full items-center justify-center p-4 font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
        {t('tiAnalyst.canvas.failed')}
      </div>
    )
  }

  const a = d.attributes
  const malicious = a.reputation_score < 0

  return (
    <div className="h-full space-y-4 overflow-y-auto px-3 py-4 sm:px-4">
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={
            'rounded-sm border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide ' +
            (malicious ? 'border-destructive/40 bg-destructive/10 text-destructive' : 'border-border/50 text-muted-foreground')
          }
        >
          {a.type}
        </span>
        <span className="min-w-0 break-words font-mono text-sm text-foreground">{a.value}</span>
        <button
          type="button"
          onClick={() => onAskAgent(t('tiAnalyst.entity.askAgent', { type: a.type, value: a.value, id: a.id }))}
          className="ms-auto flex items-center gap-1.5 rounded-sm border border-primary/30 px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider text-primary transition-colors hover:bg-primary/10"
        >
          <Sparkles className="h-3 w-3" />
          {t('tiAnalyst.entity.askAgentBtn')}
        </button>
      </div>

      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <Stat label={t('tiAnalyst.entity.reputation')} value={a.reputation} score={a.reputation_score} />
        <Stat label={t('tiAnalyst.entity.accuracy')} value={a.accuracy} score={a.accuracy_score} />
        <Stat label={t('tiAnalyst.entity.worst')} value={a.worst_reputation} score={a.worst_reputation_score} />
        <Stat label={t('tiAnalyst.entity.best')} value={a.best_reputation} score={a.best_reputation_score} />
      </div>

      {a.description && (
        <p className="font-mono text-xs leading-relaxed text-muted-foreground">{a.description}</p>
      )}

      {d.metadata && (
        <div className="flex flex-wrap gap-1.5">
          {Object.entries(d.metadata).map(([k, v]) => (
            <span
              key={k}
              className="rounded-sm border border-border/40 px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground"
            >
              {k}: <span className="text-foreground/80">{String(v)}</span>
            </span>
          ))}
        </div>
      )}

      <section className="space-y-2">
        <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          {t('tiAnalyst.entity.relations')}
        </h3>
        {relations.isLoading ? (
          <div className="h-64 animate-pulse rounded bg-muted/50" />
        ) : relations.data && relations.data.nodes.length > 1 ? (
          <RelationGraph
            graph={relations.data}
            rootId={entityId}
            height={360}
            onSelect={(node) => onPivot({ id: node.id, type: node.type, value: node.value })}
          />
        ) : (
          <div className="rounded-sm border border-dashed border-border/50 px-3 py-6 text-center font-mono text-xs text-muted-foreground">
            {t('tiAnalyst.canvas.noRelations')}
          </div>
        )}
      </section>

      {d.geolocations.length > 0 && (
        <section className="space-y-2">
          <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            {t('tiAnalyst.entity.geolocations')}
          </h3>
          {d.geolocations.map((geo, i) => (
            <div
              key={`${geo.lat}-${geo.lng}-${i}`}
              className="flex items-center gap-2 rounded-sm border border-border/40 px-3 py-2"
            >
              <MapPin className="h-3.5 w-3.5 shrink-0 text-primary/60" />
              <span className="font-mono text-xs text-foreground/90">
                {geo.lat.toFixed(4)}°, {geo.lng.toFixed(4)}°
              </span>
              <span className="font-mono text-[10px] text-muted-foreground">
                {t('tiAnalyst.entity.radius', { radius: geo.radius })}
              </span>
            </div>
          ))}
        </section>
      )}

      {d.latest_associations.length > 0 && (
        <section className="space-y-1">
          <h3 className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
            {t('tiAnalyst.entity.associations')}
          </h3>
          {d.latest_associations.map((assoc) => (
            <button
              key={assoc.id}
              type="button"
              onClick={() => onPivot({ id: assoc.id, type: assoc.type, value: assoc.value })}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-start transition-colors hover:bg-muted/40"
            >
              <span className="shrink-0 rounded-sm border border-border/50 px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground">
                {assoc.type}
              </span>
              <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground/90">{assoc.value}</span>
            </button>
          ))}
        </section>
      )}
    </div>
  )
}

function Stat({ label, value, score }: { label: string; value: string; score: number }) {
  return (
    <div className="rounded-sm border border-border/40 bg-card px-2.5 py-2">
      <div className="font-mono text-[9px] uppercase tracking-widest text-muted-foreground">{label}</div>
      <div className="truncate font-mono text-xs text-foreground/90">{value || '—'}</div>
      <div className="font-mono text-[10px] text-muted-foreground">{score}</div>
    </div>
  )
}
