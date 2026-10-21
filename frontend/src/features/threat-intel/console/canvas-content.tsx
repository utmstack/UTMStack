import { useQuery } from '@tanstack/react-query'
import { MapPin } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { proxyFetch } from './proxy'
import type { CanvasPanel } from './types'
import { EntityInspector } from './entity-inspector'
import { RelationGraph } from './relation-graph'
import { relationsToGraph } from './agent'
import type { RawRelation } from './agent'

export function CanvasContent({
  panel,
  onPivot,
  onAskAgent,
}: {
  panel: CanvasPanel
  onPivot(target: { id: string; type: string; value: string }): void
  onAskAgent(question: string): void
}) {
  switch (panel.kind) {
    case 'entity':
      return <EntityInspector entityId={panel.ref} onPivot={onPivot} onAskAgent={onAskAgent} />
    case 'graph':
      return <GraphPanel entityId={panel.ref} onPivot={onPivot} />
    case 'geo':
      return <GeoPanel entityId={panel.ref} />
    case 'search':
      return <SearchPanel query={panel.ref} onPivot={onPivot} />
  }
}

function GraphPanel({ entityId, onPivot }: { entityId: string; onPivot(target: { id: string; type: string; value: string }): void }) {
  const { t } = useTranslation()
  const relations = useQuery({
    queryKey: ['ti-analyst-relations', entityId],
    queryFn: async () => relationsToGraph(entityId, '', (await proxyFetch<RawRelation[]>(`/entity/${encodeURIComponent(entityId)}/relations`)) ?? []),
    staleTime: 5 * 60 * 1000,
  })

  if (relations.isLoading) return <PanelStatus>{t('tiAnalyst.canvas.loading')}</PanelStatus>
  if (relations.isError || !relations.data) return <PanelStatus>{t('tiAnalyst.canvas.failed')}</PanelStatus>
  if (relations.data.nodes.length <= 1) return <PanelStatus>{t('tiAnalyst.canvas.noRelations')}</PanelStatus>

  return (
    <RelationGraph
      graph={relations.data}
      rootId={entityId}
      height={260}
      onSelect={(node) => onPivot({ id: node.id, type: node.type, value: node.value })}
    />
  )
}

function GeoPanel({ entityId }: { entityId: string }) {
  const { t } = useTranslation()
  const details = useQuery({
    queryKey: ['ti-analyst-details', entityId],
    queryFn: async () => {
      const raw = await proxyFetch<{ geolocations: { latitude: number; longitude: number; accuracy_radius: number }[] }>(`/entity/${encodeURIComponent(entityId)}`)
      return raw.geolocations ?? []
    },
    staleTime: 5 * 60 * 1000,
  })

  if (details.isLoading) return <PanelStatus>{t('tiAnalyst.canvas.loading')}</PanelStatus>
  const locations = details.data ?? []
  if (!locations.length) return <PanelStatus>{t('tiAnalyst.canvas.noGeo')}</PanelStatus>

  return (
    <ul className="space-y-1">
      {locations.map((location, i) => (
        <li key={`${location.latitude}-${location.longitude}-${i}`}>
          <div className="flex items-center gap-2 rounded-md px-2 py-1.5">
            <MapPin className="h-3.5 w-3.5 shrink-0 text-primary/60" />
            <span className="font-mono text-xs text-foreground/90">
              {location.latitude.toFixed(4)}°, {location.longitude.toFixed(4)}°
            </span>
            <span className="font-mono text-[10px] text-muted-foreground">
              {t('tiAnalyst.entity.radius', { radius: location.accuracy_radius })}
            </span>
          </div>
        </li>
      ))}
    </ul>
  )
}

interface SearchItem {
  id: string
  type: string
  attributes: Record<string, string>
  reputation: number
  accuracy: number
  tags?: string[]
  lastSeen?: string
}

interface SearchHit {
  id: string
  type: string
  value: string
}

function SearchPanel({ query, onPivot }: { query: string; onPivot(target: { id: string; type: string; value: string }): void }) {
  const { t } = useTranslation()
  const results = useQuery({
    queryKey: ['ti-analyst-search', query],
    queryFn: async () => {
      const res = await proxyFetch<{ items: number; results: SearchItem[] }>('/ai/search', {
        method: 'POST',
        body: JSON.stringify({ query, page: 0, limit: 25 }),
      })
      return { items: res.items ?? 0, results: (res.results ?? []).map((r) => ({ id: r.id, type: r.type, value: r.attributes[r.type] ?? Object.values(r.attributes)[0] ?? '' })) }
    },
    staleTime: 5 * 60 * 1000,
  })

  if (results.isLoading) return <PanelStatus>{t('tiAnalyst.canvas.loading')}</PanelStatus>
  const items: SearchHit[] | undefined = results.data?.results
  if (!items?.length) return <PanelStatus>{t('tiAnalyst.canvas.noResults')}</PanelStatus>

  return (
    <div className="space-y-1">
      <div className="pb-1 text-[11px] uppercase tracking-wider text-muted-foreground">
        {t('tiAnalyst.canvas.matches', { total: results.data?.items ?? items.length })}
      </div>
      {items.map((item) => (
        <button
          key={item.id}
          type="button"
          onClick={() => onPivot({ id: item.id, type: item.type, value: item.value })}
          className="flex w-full items-center gap-2 rounded-md border border-border px-2 py-1.5 text-start transition-colors hover:border-primary/40"
        >
          <span className="shrink-0 rounded-md border border-border px-1.5 py-0.5 font-mono text-[9px] text-muted-foreground">
            {item.type}
          </span>
          <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-foreground">{item.value}</span>
        </button>
      ))}
    </div>
  )
}

function PanelStatus({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full items-center justify-center text-[11px] uppercase tracking-wider text-muted-foreground">
      {children}
    </div>
  )
}
