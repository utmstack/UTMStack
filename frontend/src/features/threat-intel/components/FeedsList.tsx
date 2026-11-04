import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { Download } from 'lucide-react'
import type { ColumnDef } from '@tanstack/react-table'
import { cn } from '@/shared/lib/utils'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { toast } from 'sonner'
import type { ThreatFeed } from '../domain/threat-intel.types'
import { useTiFeeds } from '../hooks/use-ti-feeds'
import { SkeletonRows } from './SkeletonRows'
import { feedAccuracyMeta, feedTypeTone } from './utils/severity-style'
import { downloadChecksum, downloadFeed, feedFormatLabel } from '../services/feed-download'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'

function buildFeedColumns(t: TFunction): ColumnDef<ThreatFeed>[] {
  return [
    {
      id: 'accuracyDot',
      header: () => null,
      size: 36,
      minSize: 36,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const acc = feedAccuracyMeta(row.original.accuracy)
        return <span className={cn('block h-2 w-2 rounded-full', acc.dot)} title={t(acc.labelKey)} />
      },
    },
    {
      id: 'name',
      header: t('threatIntel.feeds.table.name'),
      size: 420,
      minSize: 120,
      meta: { headerClassName: TH, cellClassName: `${TD} font-medium`, cellProps: (feed) => ({ title: feed.name }) },
      cell: ({ row }) => <span className="block truncate">{row.original.name}</span>,
    },
    {
      id: 'type',
      header: t('threatIntel.feeds.table.type'),
      size: 160,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <span className={cn('text-[11px] capitalize', feedTypeTone(row.original.type))}>{row.original.type}</span>
      ),
    },
    {
      id: 'accuracy',
      header: t('threatIntel.feeds.table.accuracy'),
      size: 140,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: `${TD} text-[11px] text-muted-foreground` },
      cell: ({ row }) => t(feedAccuracyMeta(row.original.accuracy).labelKey),
    },
    {
      id: 'format',
      header: t('threatIntel.feeds.table.format'),
      size: 120,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: `${TD} text-[11px] text-muted-foreground/70` },
      cell: ({ row }) => feedFormatLabel(row.original.type),
    },
    {
      id: 'download',
      header: () => null,
      size: 60,
      minSize: 48,
      enableResizing: false,
      meta: { headerClassName: 'whitespace-nowrap px-3 py-2.5 text-right align-middle font-medium', cellClassName: 'whitespace-nowrap px-3 py-2.5 text-right align-middle' },
      cell: ({ row }) => (
        <button
          type="button"
          title={t('threatIntel.feeds.download')}
          aria-label={t('threatIntel.feeds.download')}
          onClick={() => {
            void downloadFeed(row.original).catch((e) =>
              toast.error(e instanceof Error ? e.message : String(e)),
            )
          }}
          className="rounded-p text-muted-foreground transition-colors hover:text-primary"
        >
          <Download className="h-4 w-4" />
        </button>
      ),
    },
  ]
}

export function FeedsList() {
  const { t } = useTranslation()
  const { data, isLoading } = useTiFeeds()

  if (data?.kind === 'not-configured') return null

  const feeds = data?.kind === 'ok' ? data.value : []

  if (!isLoading && feeds.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
        {t('threatIntel.feeds.empty')}
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
          {t('threatIntel.feeds.table.header')}
        </span>
        <button
          type="button"
          onClick={() => {
            void downloadChecksum().catch((e) =>
              toast.error(e instanceof Error ? e.message : String(e)),
            )
          }}
          className="inline-flex items-center gap-1.5 rounded-sm border border-border/50 px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider text-muted-foreground transition-colors hover:bg-muted/50 hover:text-foreground"
        >
          <Download className="h-3 w-3" />
          {t('threatIntel.feeds.checksum')}
        </button>
      </div>
      <div className="max-h-[70dvh] overflow-x-auto overflow-y-auto rounded-xl border border-border bg-card">
        <ResizableDataTable
          columns={buildFeedColumns(t)}
          data={feeds}
          flexColumnId="name"
          storageKey="threat-intel-feeds-table-sizing"
          getRowId={(feed) => `${feed.type}-${feed.name}`}
          rowClassName={() => 'text-xs last:border-b-0'}
          loading={isLoading}
          loadingContent={<SkeletonRows />}
        />
      </div>
    </div>
  )
}
