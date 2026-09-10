import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { ResizableGridHeader } from '@/shared/components/ui/resizable-grid-header'
import { useResizableColumns } from '@/shared/hooks/useResizableColumns'
import type { EntitySummary } from '../domain/threat-intel.types'
import { Pagination } from '@/shared/components/ui/pagination'
import { IocRow } from './IocRow'

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

const IOC_COLS = [4, 90, '1fr', 130, '1fr', 110, 36]

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
  const { template: tableCols, startDrag } = useResizableColumns(IOC_COLS, {
    min: 36,
    storageKey: 'threat-intel-ioc-table-columns',
  })

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
          <ResizableGridHeader
            headers={[
              '',
              t('threatIntel.iocs.table.type'),
              t('threatIntel.iocs.table.indicator'),
              t('threatIntel.iocs.table.reputation'),
              t('threatIntel.iocs.table.tags'),
              t('threatIntel.iocs.table.lastSeen'),
              '',
            ]}
            tableCols={tableCols}
            startDrag={startDrag}
            className="sticky top-0 z-10"
            cellClassName="[&:nth-child(4)]:text-right"
          />
          {iocs.map((ioc) => (
            <IocRow key={ioc.id} ioc={ioc} tableCols={tableCols} onOpen={() => onOpen(ioc.id)} />
          ))}
          {!isLoading && iocs.length === 0 && (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('threatIntel.iocs.empty')}</div>
          )}
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
