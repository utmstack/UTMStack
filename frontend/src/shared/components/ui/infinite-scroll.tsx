import { useEffect, useRef } from 'react'
import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

/**
 * Sentinel that calls `onReach` when it scrolls into view — the building block for
 * infinite-scroll lists. Works with any scroll parent (window or an overflow
 * container) via IntersectionObserver. Place it right after the rendered rows.
 *
 * The page keeps `page`, `total` and `loading`; on reach it bumps the page and
 * appends the next batch. Latest callback/state are read through refs so the
 * observer binds once and never misses or double-fires.
 */
export function InfiniteScrollSentinel({
  onReach,
  hasMore,
  loading,
  endLabel,
}: {
  onReach: () => void
  hasMore: boolean
  loading: boolean
  /** Optional "all N loaded" caption shown when there's nothing left. */
  endLabel?: string
}) {
  const { t } = useTranslation()
  const ref = useRef<HTMLDivElement | null>(null)
  const cb = useRef(onReach)
  cb.current = onReach
  const state = useRef({ hasMore, loading })
  state.current = { hasMore, loading }

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && state.current.hasMore && !state.current.loading) cb.current()
      },
      { rootMargin: '400px' },
    )
    io.observe(el)
    return () => io.disconnect()
  }, [])

  return (
    <div ref={ref} className="flex items-center justify-center py-4 text-xs text-muted-foreground">
      {loading ? (
        <span className="inline-flex items-center gap-2">
          <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('common.loadingMore')}
        </span>
      ) : hasMore ? (
        ''
      ) : (
        endLabel && <span className="text-muted-foreground/70">{endLabel}</span>
      )}
    </div>
  )
}
