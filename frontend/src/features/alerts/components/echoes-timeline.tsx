import { useMemo } from 'react'
import { AlertTriangle, Loader2, Radio } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Pagination } from '@/shared/components/ui/pagination'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { TS, absTime } from '../lib/alert-meta'
import { useAlertEchoes } from '../hooks/use-alert-echoes'

interface TooltipParam {
  data: [string, string, string]
}

/* Inline panel rendered as a sibling of an expanded AlertRow. Styled as a
 * message-style card: the cyan Radio icon (matching the table chip) sits at
 * top-center, the scatter timeline sits below it. */
export function EchoesTimeline({ parentId }: { parentId: string }) {
  const { t } = useTranslation()
  const { echoes, total, page, pageSize, setPage, setPageSize, loading, error } =
    useAlertEchoes(parentId)

  const option = useMemo(
    () => ({
      grid: { left: 12, right: 12, top: 8, bottom: 26, containLabel: false },
      xAxis: { type: 'time', axisLabel: { fontSize: 10 } },
      yAxis: { type: 'category', data: [''], show: false },
      tooltip: {
        trigger: 'item',
        formatter: (p: TooltipParam) =>
          `<b>${p.data[2] || '—'}</b><br/>${absTime(p.data[0])}`,
      },
      series: [
        {
          type: 'scatter',
          symbolSize: 9,
          itemStyle: { color: '#06b6d4', opacity: 0.85 },
          data: echoes.map((e) => [e[TS] ?? '', '', e.name ?? '—']),
        },
      ],
    }),
    [echoes],
  )

  return (
    <div className="border-b border-border/50 bg-muted/20 px-6 py-4">
      <div className="mx-auto max-w-2xl rounded-lg border border-cyan-500/30 bg-cyan-500/5 px-5 py-4 shadow-sm">
        <div className="flex flex-col items-center gap-1">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-cyan-500/15 text-cyan-600 ring-1 ring-cyan-500/40 dark:text-cyan-300">
            <Radio size={16} />
          </div>
          <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
            <span className="font-medium">{t('alerts.echoes.title', { count: total })}</span>
            {loading && <Loader2 size={12} className="animate-spin" />}
          </div>
        </div>

        <div className="mt-3">
          {error ? (
            <div className="flex items-center justify-center gap-2 py-4 text-xs text-amber-600">
              <AlertTriangle size={14} /> {t('alerts.echoes.loadError')}
            </div>
          ) : total === 0 && !loading ? (
            <div className="py-4 text-center text-xs text-muted-foreground">
              {t('alerts.echoes.empty')}
            </div>
          ) : (
            <div className="h-24 w-full">
              <EChartsRenderer option={option} />
            </div>
          )}
        </div>

        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          loading={loading}
          onPageChange={setPage}
          onPageSizeChange={setPageSize}
          pageSizeOptions={[10, 20, 50, 100]}
        />
      </div>
    </div>
  )
}
