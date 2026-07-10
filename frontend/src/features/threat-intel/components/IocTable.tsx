import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
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

const IOC_COLS = '4px 90px 1fr 130px 1fr 110px 36px'

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
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: IOC_COLS }}
        >
          <div />
          <div>{t('threatIntel.iocs.table.type')}</div>
          <div>{t('threatIntel.iocs.table.indicator')}</div>
          <div className="text-right">{t('threatIntel.iocs.table.reputation')}</div>
          <div>{t('threatIntel.iocs.table.tags')}</div>
          <div>{t('threatIntel.iocs.table.lastSeen')}</div>
          <div />
        </div>
        <div ref={scrollRef} className="flex-1 overflow-y-auto">
          {iocs.map((ioc) => (
            <IocRow key={ioc.id} ioc={ioc} onOpen={() => onOpen(ioc.id)} />
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
