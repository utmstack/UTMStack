import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { LayoutDashboard, Loader2, Lock, Plus, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useVisualizations } from '@/features/dashboard/hooks/useVisualizations'
import { getChartIcon } from '@/features/dashboard/components/chart-icon'
import { parseBuilderConfig } from '@/features/dashboard/utils/builder-config'
import { parseLayout } from '@/features/dashboard/utils/layout'
import { GRID_COLS } from '@/features/dashboard/constants'
import type { ChartTypeId, Dashboard, Visualization } from '@/features/dashboard/types'

export function DashboardGallery({
  dashboards,
  loading,
  search,
  onSearchChange,
  onSelect,
  onCreate,
}: {
  dashboards: Dashboard[]
  loading: boolean
  search: string
  onSearchChange: (value: string) => void
  onSelect: (id: number) => void
  onCreate: () => void
}) {
  const { t } = useTranslation()

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-base font-semibold">{t('dashboards.list.title')}</h1>
        <Button size="sm" onClick={onCreate}>
          <Plus size={14} className="mr-1" />
          {t('dashboards.list.create')}
        </Button>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={t('dashboards.list.searchPlaceholder') ?? ''}
            className="h-9 pl-9"
          />
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto">
        {loading && (
          <div className="flex items-center justify-center gap-2 py-10 text-xs text-muted-foreground">
            <Loader2 size={14} className="animate-spin" />
            {t('dashboards.list.loading')}
          </div>
        )}
        {!loading && dashboards.length === 0 && (
          <div className="py-10 text-center text-xs text-muted-foreground">
            {t('dashboards.list.empty')}
          </div>
        )}
        {!loading && dashboards.length > 0 && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {dashboards.map((d) => (
              <DashboardCard key={d.id} dashboard={d} onSelect={onSelect} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Editing/deleting a dashboard only happens from inside it (DashboardPreviewHeader)
// — the gallery is browse-only, so a card is just its preview + metadata, no actions.
function DashboardCard({ dashboard: d, onSelect }: { dashboard: Dashboard; onSelect: (id: number) => void }) {
  const modified = d.modifiedDate ? new Date(d.modifiedDate).toLocaleString() : '—'
  // Just enough to draw the mini layout skeleton — no query execution, so this
  // stays cheap even with many dashboards on screen at once.
  const visualizations = useVisualizations({ dashboardId: d.id, page: 0, size: 50 })

  return (
    <div
      onClick={() => onSelect(d.id)}
      className="flex cursor-pointer flex-col overflow-hidden rounded-lg border border-border bg-card transition-colors hover:border-primary/40 hover:shadow-sm"
    >
      <MiniLayoutPreview visualizations={visualizations.data?.data ?? []} loading={visualizations.isLoading} />

      <div className="border-t border-border px-3 py-2.5">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-sm font-medium text-foreground" title={d.name}>
            {d.name}
          </span>
          {d.systemOwner && <Lock size={11} className="shrink-0 text-muted-foreground/60" />}
        </div>
        {d.description && (
          <p className="mt-0.5 truncate text-xs text-muted-foreground" title={d.description}>
            {d.description}
          </p>
        )}
        <p className="mt-1 text-[10px] text-muted-foreground/70">{modified}</p>
      </div>
    </div>
  )
}

/** Wireframe-style mini preview: a rectangle per widget, positioned/sized from
 * its stored grid layout and scaled to fit the card — no data, no queries. */
function MiniLayoutPreview({ visualizations, loading }: { visualizations: Visualization[]; loading: boolean }) {
  const { t } = useTranslation()
  const items = useMemo(
    () =>
      visualizations.map((v) => ({
        id: v.id,
        chartType: parseBuilderConfig(v.config).builder?.chartType,
        ...parseLayout(v.layout),
      })),
    [visualizations]
  )

  const maxX = Math.max(GRID_COLS, ...items.map((i) => i.x + i.w))
  const maxY = Math.max(1, ...items.map((i) => i.y + i.h))

  return (
    <div className="relative h-44 w-full shrink-0 overflow-hidden bg-muted/30">
      {loading ? (
        <div className="absolute inset-0 animate-pulse bg-muted/50" />
      ) : items.length === 0 ? (
        <div className="flex h-full w-full items-center justify-center gap-1.5 px-4 text-center text-[11px] text-muted-foreground/60">
          <LayoutDashboard size={14} className="shrink-0" />
          {t('dashboards.grid.empty')}
        </div>
      ) : (
        items.map((it) => (
          <div
            key={it.id}
            style={{
              left: `calc(${(it.x / maxX) * 100}% + 2px)`,
              top: `calc(${(it.y / maxY) * 100}% + 2px)`,
              width: `calc(${(it.w / maxX) * 100}% - 4px)`,
              height: `calc(${(it.h / maxY) * 100}% - 4px)`,
            }}
            className="absolute overflow-hidden rounded-[3px] border border-primary/20 bg-primary/10 text-primary"
          >
            <MiniChartMock chartType={it.chartType} seed={it.id} />
          </div>
        ))
      )}
    </div>
  )
}

// Deterministic pseudo-random sequence (same widget always draws the same
// mock) — a real PRNG would be overkill for a decorative sketch.
function seededValues(seed: number, count: number, min: number, max: number): number[] {
  let s = seed || 1
  const out: number[] = []
  for (let i = 0; i < count; i++) {
    s = (s * 9301 + 49297) % 233280
    out.push(min + (s / 233280) * (max - min))
  }
  return out
}

/** A tiny sketch of what the chart type looks like — no query, no real data,
 * just enough shape to tell a bar card from a pie card at a glance. */
function MiniChartMock({ chartType, seed }: { chartType: ChartTypeId | undefined; seed: number }) {
  switch (chartType) {
    case 'bar': {
      const heights = seededValues(seed, 4, 22, 50)
      return (
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" className="h-full w-full">
          {heights.map((h, i) => (
            <rect key={i} x={8 + i * 24} y={56 - h} width="14" height={h} rx="1.5" fill="currentColor" opacity="0.55" />
          ))}
        </svg>
      )
    }
    case 'line':
    case 'area': {
      const ys = seededValues(seed, 5, 8, 44)
      const pts = ys.map((y, i) => `${6 + i * 22},${56 - y}`).join(' ')
      return (
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" className="h-full w-full">
          {chartType === 'area' && (
            <polygon points={`6,56 ${pts} 94,56`} fill="currentColor" opacity="0.15" />
          )}
          <polyline
            points={pts}
            fill="none"
            stroke="currentColor"
            strokeWidth="2.5"
            opacity="0.6"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      )
    }
    case 'pie': {
      const splits = seededValues(seed, 3, 15, 40)
      const total = splits.reduce((a, b) => a + b, 0)
      const r = 22
      const c = 30
      let acc = 0
      return (
        <svg viewBox="0 0 60 60" className="h-full w-full">
          {splits.map((v, i) => {
            const frac = v / total
            const dash = frac * 2 * Math.PI * r
            const rotate = (acc / total) * 360 - 90
            acc += v
            return (
              <circle
                key={i}
                cx={c}
                cy={c}
                r={r}
                fill="none"
                stroke="currentColor"
                strokeWidth="13"
                strokeDasharray={`${dash} ${2 * Math.PI * r}`}
                opacity={0.3 + i * 0.2}
                transform={`rotate(${rotate} ${c} ${c})`}
              />
            )
          })}
        </svg>
      )
    }
    case 'metric': {
      const n = Math.round(seededValues(seed, 1, 12, 980)[0])
      return (
        <div className="flex h-full w-full items-center justify-center">
          <span className="text-base font-bold opacity-50">{n.toLocaleString()}</span>
        </div>
      )
    }
    case 'table': {
      const widths = seededValues(seed, 4, 45, 92)
      return (
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" className="h-full w-full">
          {widths.map((w, i) => (
            <rect key={i} x="5" y={7 + i * 13} width={w} height="6" rx="1.5" fill="currentColor" opacity={i === 0 ? 0.45 : 0.22} />
          ))}
        </svg>
      )
    }
    case 'list': {
      const widths = seededValues(seed, 4, 55, 90)
      return (
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" className="h-full w-full">
          {widths.map((w, i) => (
            <rect key={i} x="5" y={8 + i * 13} width={w} height="4" rx="1.5" fill="currentColor" opacity="0.3" />
          ))}
        </svg>
      )
    }
    case 'text': {
      const widths = seededValues(seed, 3, 50, 95)
      return (
        <svg viewBox="0 0 100 60" preserveAspectRatio="none" className="h-full w-full">
          {widths.map((w, i) => (
            <rect key={i} x="5" y={10 + i * 15} width={w} height="4" rx="1.5" fill="currentColor" opacity="0.3" />
          ))}
        </svg>
      )
    }
    case 'region_map': {
      const pts = seededValues(seed, 3, 12, 88)
      return (
        <svg viewBox="0 0 100 60" className="h-full w-full">
          <rect x="2" y="2" width="96" height="56" rx="4" fill="currentColor" opacity="0.06" />
          {pts.map((x, i) => (
            <circle key={i} cx={x} cy={16 + ((i * 17) % 30)} r="3" fill="currentColor" opacity="0.6" />
          ))}
        </svg>
      )
    }
    default: {
      const Icon = getChartIcon(chartType)
      return (
        <div className="flex h-full w-full items-center justify-center">
          <Icon size={12} className="opacity-60" />
        </div>
      )
    }
  }
}
