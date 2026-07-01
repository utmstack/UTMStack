import { Radio } from 'lucide-react'
import { MouseEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'

/* Cyan-tinted chip showing the echo count for an alert row. Mirrors the
 * <TagChip> visual shape (rounded ring + icon + label). When count is 0 the
 * button is still rendered but visually disabled — clicks become no-ops. */
export function EchoesChip({
  count,
  expanded,
  onClick,
}: {
  count: number
  expanded: boolean
  onClick: () => void
}) {
  const { t } = useTranslation()
  const enabled = count > 0

  const handleClick = (e: MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation()
    if (enabled) onClick()
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      disabled={!enabled}
      title={
        enabled
          ? t('alerts.row.echoesTooltip', { count })
          : t('alerts.row.echoesEmpty')
      }
      aria-pressed={expanded}
      className={cn(
        'inline-flex items-center gap-1 whitespace-nowrap rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none ring-1 ring-inset transition-colors',
        enabled
          ? 'bg-cyan-500/10 text-cyan-600 ring-cyan-500/30 hover:bg-cyan-500/20 dark:text-cyan-300'
          : 'cursor-not-allowed bg-muted/40 text-muted-foreground/60 ring-border/40',
        enabled && expanded && 'bg-cyan-500/25 ring-cyan-500/50',
      )}
    >
      <Radio size={10} className="shrink-0" />
      <span className="tabular-nums">{count}</span>
    </button>
  )
}
