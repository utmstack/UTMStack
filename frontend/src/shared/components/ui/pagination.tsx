import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from './button'

const DEFAULT_PAGE_SIZES = [10, 20, 50, 100]

/**
 * Classic page-size + page-of-N + prev/next pager, matching the Notifications page.
 * `page` is 0-based. Renders nothing when there is nothing to page through.
 */
export function Pagination({
  page,
  pageSize,
  total,
  loading,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = DEFAULT_PAGE_SIZES,
  align = 'between',
}: {
  page: number
  pageSize: number
  total: number
  loading?: boolean
  onPageChange: (page: number) => void
  onPageSizeChange: (size: number) => void
  pageSizeOptions?: number[]
  /** 'between' spreads controls across the row; 'right' groups them all to the right. */
  align?: 'between' | 'right'
}) {
  const { t } = useTranslation()
  if (total <= 0) return null
  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div
      className={cn(
        'mt-4 flex flex-wrap items-center gap-3',
        align === 'right' ? 'justify-end' : 'justify-between'
      )}
    >
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <span>{t('pagination.perPage')}</span>
        <select
          value={pageSize}
          onChange={(e) => onPageSizeChange(Number(e.target.value))}
          className="cursor-pointer rounded-md border border-border bg-card px-1.5 py-1 text-xs text-foreground outline-none focus:ring-1 focus:ring-ring"
        >
          {pageSizeOptions.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        <span className="ml-1">{t('pagination.pageOf', { page: page + 1, total: totalPages })}</span>
      </div>
      {totalPages > 1 && (
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 0 || loading}
            onClick={() => onPageChange(Math.max(0, page - 1))}
          >
            <ChevronLeft size={14} className="mr-1" />
            {t('pagination.previous')}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={page >= totalPages - 1 || loading}
            onClick={() => onPageChange(Math.min(totalPages - 1, page + 1))}
          >
            {t('pagination.next')}
            <ChevronRight size={14} className="ml-1" />
          </Button>
        </div>
      )}
    </div>
  )
}
