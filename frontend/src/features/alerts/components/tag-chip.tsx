import { Tag as TagIcon } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { AlertTag } from '../types/alert.types'

/* A tag chip painted with the tag's own colour (from the tag catalog). The alert
 * only stores tag names, so we look up the colour by name. */
export function TagChip({ name, catalog, size = 'sm' }: { name: string; catalog: AlertTag[]; size?: 'sm' | 'xs' }) {
  const color = catalog.find((t) => t.tagName === name)?.tagColor
  const style = color ? { backgroundColor: `${color}26`, color, borderColor: `${color}59` } : undefined
  return (
    <span
      className={cn(
        'inline-flex max-w-[140px] shrink-0 items-center gap-1 whitespace-nowrap rounded-md border font-medium leading-none',
        size === 'xs' ? 'px-1.5 py-0.5 text-[10px]' : 'px-2 py-0.5 text-[11px]',
        !color && 'border-border bg-muted text-muted-foreground'
      )}
      style={style}
      title={name}
    >
      <TagIcon size={size === 'xs' ? 9 : 10} className="shrink-0" />
      <span className="truncate">{name}</span>
    </span>
  )
}
