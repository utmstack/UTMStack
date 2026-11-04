import { useQuery } from '@tanstack/react-query'
import { Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { proxyFetch } from './proxy'
import { normalizeDetails, relationsToGraph } from './agent'
import type { EntityDetails, EntityHistoryRecord } from './types'
import type { RawDetails, RawRelations } from './agent'
import { RelationGraph } from './relation-graph'
import { EntityVerdict, ExtendedMetadataCard, GeoCard, HistoryRecordCard, MetadataCard } from './entity-cards'

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
    queryFn: async () => relationsToGraph(entityId, '', (await proxyFetch<RawRelations>(`/entity/${encodeURIComponent(entityId)}/relations`)) ?? []),
    staleTime: 10 * 60 * 1000,
  })

  // History endpoint shape is upstream-dependent: accept an array or a wrapper.
  const history = useQuery({
    queryKey: ['ti-analyst-history', entityId],
    queryFn: async (): Promise<EntityHistoryRecord[]> => {
      const raw = await proxyFetch<unknown>(`/entity/${encodeURIComponent(entityId)}/history`)
      const arr = Array.isArray(raw) ? raw : ((raw as Record<string, unknown> | null)?.items as unknown[]) ?? ((raw as Record<string, unknown> | null)?.records as unknown[]) ?? []
      return arr.map((r) => {
        const x = r as Record<string, unknown>
        return {
          timestamp: String(x.timestamp ?? x.time ?? ''),
          reputation: Number(x.reputation ?? x.reputation_score ?? 0),
          author: x.author ? String(x.author) : undefined,
        }
      })
    },
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
      <div className="flex h-full items-center justify-center p-4 text-[11px] uppercase tracking-wider text-muted-foreground">
        {t('tiAnalyst.canvas.failed')}
      </div>
    )
  }

  const a = d.attributes

  return (
    <div className="h-full space-y-4 overflow-y-auto px-3 py-4 sm:px-4">
      <div className="flex flex-wrap items-center gap-2">
        <span
          className={
            'rounded-md border border-border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground'
          }
        >
          {a.type}
        </span>
        <span className="min-w-0 break-words font-mono text-sm text-foreground">{a.value}</span>
        <button
          type="button"
          onClick={() => onAskAgent(t('tiAnalyst.entity.askAgent', { type: a.type, value: a.value, id: a.id }))}
          className="ms-auto flex items-center gap-1.5 rounded-md border border-primary/30 px-2.5 py-1 text-[11px] uppercase tracking-wider text-primary transition-colors hover:bg-primary/10"
        >
          <Sparkles className="h-3 w-3" />
          {t('tiAnalyst.entity.askAgentBtn')}
        </button>
      </div>

      <EntityVerdict attributes={a} />

      {d.metadata && <MetadataCard metadata={d.metadata} />}
      {d.extendedMetadata && <ExtendedMetadataCard entries={d.extendedMetadata} />}

      <section className="space-y-2">
        <h3 className="text-[11px] uppercase tracking-wider text-muted-foreground">
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
          <div className="rounded-md border border-dashed border-border px-3 py-6 text-center font-mono text-xs text-muted-foreground">
            {t('tiAnalyst.canvas.noRelations')}
          </div>
        )}
      </section>

      {history.data && history.data.length > 0 && (
        <section className="space-y-2">
          <h3 className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t('tiAnalyst.entity.history')}
          </h3>
          {history.data.map((record, i) => (
            <HistoryRecordCard key={i} record={record} previous={history.data[i + 1]} />
          ))}
        </section>
      )}

      {d.geolocations.length > 0 && (
        <section className="space-y-2">
          <h3 className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t('tiAnalyst.entity.geolocations')}
          </h3>
          {d.geolocations.map((geo, i) => (
            <GeoCard key={`${geo.lat}-${geo.lng}-${i}`} latitude={geo.lat} longitude={geo.lng} radius={geo.radius} />
          ))}
        </section>
      )}

      {d.latest_associations.length > 0 && (
        <section className="space-y-1">
          <h3 className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t('tiAnalyst.entity.associations')}
          </h3>
          {d.latest_associations.map((assoc) => (
            <button
              key={assoc.id}
              type="button"
              onClick={() => onPivot({ id: assoc.id, type: assoc.type, value: assoc.value })}
              className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-start transition-colors hover:bg-muted/40"
            >
              <span className="shrink-0 rounded-md border border-border px-1.5 py-0.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground">
                {assoc.type}
              </span>
              <span className="min-w-0 flex-1 truncate font-mono text-xs text-foreground/90">{assoc.value}</span>
              <span dir="ltr" className="shrink-0 font-mono text-[10px] text-muted-foreground">
                {assoc.reputation > 0 ? `+${assoc.reputation}` : assoc.reputation}
              </span>
            </button>
          ))}
        </section>
      )}
    </div>
  )
}
