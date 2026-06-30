import { useEffect, useRef } from 'react'
import { cn } from '@/shared/lib/utils'
import type { Suggestion } from '../services/autocomplete-trie.service'

interface Props {
  items: Suggestion[]
  activeIndex: number
  position: { x: number; y: number }
  onPick: (item: Suggestion) => void
  onHover: (index: number) => void
}

const ROW_HEIGHT = 28
const MAX_VISIBLE = 8

export function SqlAutocompleteDropdown({
  items,
  activeIndex,
  position,
  onPick,
  onHover,
}: Props) {
  const listRef = useRef<HTMLUListElement>(null)

  useEffect(() => {
    const el = listRef.current?.children[activeIndex] as HTMLElement | undefined
    el?.scrollIntoView({ block: 'nearest' })
  }, [activeIndex])

  if (items.length === 0) return null

  return (
    <ul
      ref={listRef}
      role="listbox"
      className="fixed z-50 min-w-[220px] overflow-y-auto rounded-md border border-border bg-popover py-1 text-xs shadow-lg"
      style={{
        left: position.x,
        top: position.y,
        maxHeight: ROW_HEIGHT * MAX_VISIBLE,
      }}
      onMouseDown={(e) => e.preventDefault()}
    >
      {items.map((item, i) => (
        <li
          key={`${item.tag}:${item.word}`}
          role="option"
          aria-selected={i === activeIndex}
          className={cn(
            'flex cursor-pointer items-center justify-between gap-3 px-2.5 py-1 font-mono',
            i === activeIndex
              ? 'bg-accent text-accent-foreground'
              : 'text-foreground hover:bg-accent/60',
          )}
          onMouseEnter={() => onHover(i)}
          onClick={() => onPick(item)}
        >
          <span className="truncate">{item.word}</span>
          <span
            className={cn(
              'shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wider',
              item.tag === 'sql' && 'bg-violet-500/15 text-violet-600 dark:text-violet-300',
              item.tag === 'field' && 'bg-sky-500/15 text-sky-600 dark:text-sky-300',
              item.tag === 'index' && 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300',
            )}
          >
            {item.tag === 'sql' ? 'SQL' : item.tag === 'index' ? 'index' : 'field'}
          </span>
        </li>
      ))}
    </ul>
  )
}
