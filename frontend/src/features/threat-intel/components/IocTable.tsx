import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { MoreHorizontal } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Pagination } from '@/shared/components/ui/pagination'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { searchItemValue, type EntitySummary } from '../domain/threat-intel.types'
import { SkeletonRows } from './SkeletonRows'
import { REPUTATION_STYLE, reputationLabelKey, reputationTone, typeMeta } from './utils/severity-style'
import { absTimestamp } from './utils/time-format'

interface IocTableProps {
  iocs: EntitySummary[]
  onOpen: (id: string) => void
  isLoading?: boolean
  page: number
  pageSize: number
  totalItems: number
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  hasMore?: boolean
  onLoadMore?: () => void
}

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'

function buildIocColumns(t: TFunction): ColumnDef<EntitySummary>[] {
  return [
    {
      id: 'reputationBar',
      header: () => null,
      size: 28,
      minSize: 28,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <span className={cn('block h-3 w-1 rounded-full', REPUTATION_STYLE[reputationTone(row.original.reputation)].bar)} />
      ),
    },
    {
      id: 'type',
      header: t('threatIntel.iocs.table.type'),
      size: 110,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const meta = typeMeta(row.original.type)
        const Icon = meta.icon
        return (
          <span className="flex items-center gap-1.5">
            <Icon size={11} className={cn('shrink-0', meta.tone)} />
            <span className="truncate font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
              {t(meta.labelKey)}
            </span>
          </span>
        )
      },
    },
    {
      id: 'indicator',
      header: t('threatIntel.iocs.table.indicator'),
      size: 360,
      minSize: 120,
      meta: {
        headerClassName: TH,
        cellClassName: `${TD} font-mono`,
        cellProps: (ioc) => ({ title: searchItemValue(ioc) }),
      },
      cell: ({ row }) => <span className="block truncate">{searchItemValue(row.original)}</span>,
    },
    {
      id: 'reputation',
      header: t('threatIntel.iocs.table.reputation'),
      size: 130,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const rep = REPUTATION_STYLE[reputationTone(row.original.reputation)]
        return (
          <span className={cn('text-[11px] font-medium', rep.tone)}>{t(reputationLabelKey(row.original.reputation))}</span>
        )
      },
    },
    {
      id: 'tags',
      header: t('threatIntel.iocs.table.tags'),
      size: 240,
      minSize: 80,
      meta: {
        headerClassName: TH,
        cellClassName: TD,
        cellProps: (ioc) => ({ title: ioc.tags?.join(', ') }),
      },
      cell: ({ row }) => {
        const tags = row.original.tags ?? []
        return (
          <span className="flex items-center gap-1">
            {tags.slice(0, 3).map((tag) => (
              <span key={tag} className="truncate rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
                {tag}
              </span>
            ))}
            {tags.length > 3 && <span className="shrink-0 text-[10px] text-muted-foreground">+{tags.length - 3}</span>}
          </span>
        )
      },
    },
    {
      id: 'lastSeen',
      header: t('threatIntel.iocs.table.lastSeen'),
      size: 190,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: `${TD} font-mono text-[11px] text-muted-foreground` },
      cell: ({ row }) => (row.original.lastSeen ? absTimestamp(row.original.lastSeen) : '—'),
    },
    {
      id: 'actions',
      header: () => null,
      size: 52,
      minSize: 52,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: () => (
        <div className="flex justify-end opacity-0 group-hover:opacity-100">
          <button
            onClick={(e) => e.stopPropagation()}
            className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground"
          >
            <MoreHorizontal size={14} />
          </button>
        </div>
      ),
    },
  ]
}

export function IocTable({
  iocs,
  onOpen,
  isLoading,
  page,
  pageSize,
  totalItems,
  onPageChange,
  onPageSizeChange,
  hasMore,
  onLoadMore,
}: IocTableProps) {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement | null>(null)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    const node = sentinelRef.current
    const root = scrollRef.current
    if (!node || !root || !hasMore || !onLoadMore) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && !isLoading) onLoadMore()
      },
      { root, rootMargin: '200px' }
    )
    io.observe(node)
    return () => io.disconnect()
  }, [hasMore, onLoadMore, isLoading])

  return (
    <>
      <div className="flex h-[60dvh] flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div ref={scrollRef} className="flex-1 overflow-x-auto overflow-y-auto">
          <ResizableDataTable
            columns={buildIocColumns(t)}
            data={iocs}
            flexColumnId="indicator"
            storageKey="threat-intel-ioc-table-sizing"
            getRowId={(ioc) => ioc.id}
            onRowClick={(ioc) => onOpen(ioc.id)}
            rowClassName={() => 'text-xs last:border-b-0'}
            loading={isLoading && iocs.length === 0}
            loadingContent={<SkeletonRows />}
            emptyContent={
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('threatIntel.iocs.empty')}</div>
            }
          />
          <div ref={sentinelRef} />
          {hasMore && (
            <div className="px-6 py-4 text-center text-[11px] text-muted-foreground">
              {isLoading ? t('threatIntel.iocs.loadingMore') : t('threatIntel.iocs.loadedProgress', { loaded: iocs.length, total: totalItems })}
            </div>
          )}
          {!hasMore && iocs.length > 0 && (
            <div className="px-6 py-4 text-center text-[11px] text-muted-foreground">
              {t('threatIntel.iocs.endOfResults', { loaded: iocs.length, total: totalItems })}
            </div>
          )}
        </div>
      </div>
      <Pagination
        page={page}
        pageSize={pageSize}
        total={totalItems}
        loading={isLoading}
        onPageChange={onPageChange}
        onPageSizeChange={onPageSizeChange}
      />
    </>
  )
}
