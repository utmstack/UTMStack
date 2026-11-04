import { useCallback, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

export const DEFAULT_CANVAS_FRACTION = 0.46
export const MIN_CONVERSATION_PX = 360
export const MIN_CANVAS_PX = 320
export const SPLIT_HANDLE_PX = 6

export function clampCanvasFraction(fraction: number, containerWidth: number): number {
  if (!Number.isFinite(fraction) || !Number.isFinite(containerWidth) || containerWidth <= 0) {
    return DEFAULT_CANVAS_FRACTION
  }
  const min = MIN_CANVAS_PX / containerWidth
  const max = (containerWidth - MIN_CONVERSATION_PX - SPLIT_HANDLE_PX) / containerWidth
  if (min > max) return 0.5
  return Math.min(max, Math.max(min, fraction))
}

export function fractionFromDrag(
  startFraction: number,
  deltaX: number,
  containerWidth: number,
): number {
  if (containerWidth <= 0) return startFraction
  const delta = -deltaX / containerWidth
  return clampCanvasFraction(startFraction + delta, containerWidth)
}

export function useSplit() {
  const [fraction, setFraction] = useState<number>(() => {
    try {
      const raw = localStorage.getItem('utmstack.ti-analyst.split')
      if (!raw) return DEFAULT_CANVAS_FRACTION
      const parsed = Number.parseFloat(raw)
      return Number.isFinite(parsed) && parsed > 0.05 && parsed < 0.95 ? parsed : DEFAULT_CANVAS_FRACTION
    } catch {
      return DEFAULT_CANVAS_FRACTION
    }
  })

  const setClamped = useCallback((next: number, containerWidth: number) => {
    setFraction(clampCanvasFraction(next, containerWidth))
  }, [])

  const reset = useCallback(() => setFraction(DEFAULT_CANVAS_FRACTION), [])

  return { fraction, setFraction: setClamped, reset }
}

export function SplitHandle({
  fraction,
  containerWidth,
  onChange,
  onReset,
}: {
  fraction: number
  containerWidth: number
  onChange(fraction: number): void
  onReset(): void
}) {
  const { t } = useTranslation()
  const origin = useRef<{ x: number; fraction: number } | null>(null)
  const [dragging, setDragging] = useState(false)

  const beginDrag = useCallback(
    (e: React.PointerEvent) => {
      if (e.button !== 0) return
      try {
        e.currentTarget.setPointerCapture(e.pointerId)
      } catch {
      }
      origin.current = { x: e.clientX, fraction }
      setDragging(true)
    },
    [fraction],
  )

  const onPointerMove = useCallback(
    (e: React.PointerEvent) => {
      const from = origin.current
      if (!from) return
      onChange(fractionFromDrag(from.fraction, e.clientX - from.x, containerWidth))
    },
    [onChange, containerWidth],
  )

  const endDrag = useCallback((e: React.PointerEvent) => {
    origin.current = null
    if (e.currentTarget.hasPointerCapture(e.pointerId)) {
      e.currentTarget.releasePointerCapture(e.pointerId)
    }
    setDragging(false)
  }, [])

  const onKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const step = (e.shiftKey ? 80 : 20)
      const moves: Record<string, number | undefined> = {
        ArrowLeft: fractionFromDrag(fraction, -step, containerWidth),
        ArrowRight: fractionFromDrag(fraction, step, containerWidth),
        Home: fractionFromDrag(fraction, containerWidth, containerWidth),
        End: fractionFromDrag(fraction, -containerWidth, containerWidth),
      }
      const next = moves[e.key]
      if (next === undefined) return
      e.preventDefault()
      onChange(next)
    },
    [fraction, containerWidth, onChange],
  )

  const percent = Math.round(fraction * 100)

  return (
    <div
      role="separator"
      tabIndex={0}
      aria-orientation="vertical"
      aria-label={t('tiAnalyst.console.resizeResults')}
      aria-valuenow={percent}
      aria-valuemin={0}
      aria-valuemax={100}
      onPointerDown={beginDrag}
      onPointerMove={onPointerMove}
      onPointerUp={endDrag}
      onPointerCancel={endDrag}
      onKeyDown={onKeyDown}
      onDoubleClick={onReset}
      title={t('tiAnalyst.console.resizeHint')}
      className="group relative w-1.5 shrink-0 cursor-col-resize touch-none focus:outline-none focus-visible:ring-1 focus-visible:ring-primary"
    >
      <span
        aria-hidden
        className={cn(
          'absolute inset-y-0 start-1/2 w-px -translate-x-1/2 transition-colors',
          dragging ? 'bg-primary' : 'bg-border/60 group-hover:bg-primary/60',
        )}
      />
    </div>
  )
}
