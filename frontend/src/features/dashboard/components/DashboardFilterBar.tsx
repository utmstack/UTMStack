import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Filter, Loader2, Pencil, Trash2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { usePropertyValues } from '@/features/dashboard/hooks/usePropertyValues'
import { DashboardFilterChipEditor } from '@/features/dashboard/components/DashboardFilterChipEditor'
import type { DashboardFilterChip } from '@/features/dashboard/types'

export type ChipValueMap = Record<string, string | string[]>

const POPOVER_WIDTH = 720
const POPOVER_GAP = 8
const VIEWPORT_MARGIN = 16

/**
 * Horizontal chip bar above the dashboard grid. In edit mode each chip exposes
 * pencil + trash icons (open the config popover / open a confirm dialog), plus
 * an "Add filter" trigger that spawns an empty chip form. Outside edit mode the
 * chips work as value pickers (unchanged).
 */
export function DashboardFilterBar({
  chips,
  values,
  onChange,
  editable = false,
  onSaveChips,
  savingChips = false,
}: {
  chips: DashboardFilterChip[]
  values: ChipValueMap
  onChange: (next: ChipValueMap) => void
  editable?: boolean
  onSaveChips?: (next: DashboardFilterChip[]) => void
  savingChips?: boolean
}) {
  const { t } = useTranslation()
  // Anchor + mode for the single-chip editor popover. `chip=null` → add mode,
  // `chip=<chip>` → edit mode (form pre-filled).
  const [popover, setPopover] = useState<
    | { anchor: HTMLElement; chip: DashboardFilterChip | null }
    | null
  >(null)
  const [pendingDelete, setPendingDelete] = useState<DashboardFilterChip | null>(null)

  if (chips.length === 0 && !editable) return null

  const setValue = (id: string, next: string | string[] | null) => {
    if (next === null || (Array.isArray(next) && next.length === 0) || next === '') {
      const { [id]: _, ...rest } = values
      onChange(rest)
      return
    }
    onChange({ ...values, [id]: next })
  }

  const submitChip = (chip: DashboardFilterChip) => {
    if (!popover) return
    const next = popover.chip
      ? chips.map((c) => (c.id === chip.id ? chip : c))
      : [...chips, chip]
    onSaveChips?.(next)
    setPopover(null)
  }

  const confirmRemove = () => {
    if (!pendingDelete) return
    onSaveChips?.(chips.filter((c) => c.id !== pendingDelete.id))
    setPendingDelete(null)
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      {chips.map((chip) =>
        editable ? (
          <EditableChip
            key={chip.id}
            chip={chip}
            onEdit={(anchor) => setPopover({ anchor, chip })}
            onRemove={() => setPendingDelete(chip)}
          />
        ) : (
          <FilterChip
            key={chip.id}
            chip={chip}
            value={values[chip.id]}
            onChange={(next) => setValue(chip.id, next)}
            clearLabel={t('dashboards.filters.clear')}
          />
        )
      )}
      {editable && (
        <AddFilterButton onOpen={(anchor) => setPopover({ anchor, chip: null })} />
      )}
      {savingChips && (
        <span
          className="flex items-center gap-1.5 text-xs text-muted-foreground"
          role="status"
          aria-live="polite"
        >
          <Loader2 size={12} className="animate-spin" />
          {t('dashboards.filters.saving')}
        </span>
      )}

      {popover && (
        <ChipEditorPopover
          anchor={popover.anchor}
          initial={popover.chip}
          busy={savingChips}
          onCancel={() => setPopover(null)}
          onSubmit={submitChip}
        />
      )}

      <ConfirmDialog
        open={pendingDelete != null}
        title={t('dashboards.filters.removeTitle')}
        body={t('dashboards.filters.removeConfirm', { label: pendingDelete?.label ?? '' })}
        confirmLabel={t('dashboards.filters.remove') ?? undefined}
        danger
        busy={savingChips}
        onClose={() => setPendingDelete(null)}
        onConfirm={confirmRemove}
      />
    </div>
  )
}

function AddFilterButton({ onOpen }: { onOpen: (anchor: HTMLElement) => void }) {
  const { t } = useTranslation()
  return (
    <button
      type="button"
      onClick={(e) => onOpen(e.currentTarget)}
      className="flex h-8 items-center gap-1.5 rounded-md border border-dashed border-input px-2 text-xs text-muted-foreground transition-colors hover:border-primary/60 hover:text-foreground"
    >
      <Filter size={12} />
      {t('dashboards.filters.addChip')}
    </button>
  )
}

function EditableChip({
  chip,
  onEdit,
  onRemove,
}: {
  chip: DashboardFilterChip
  onEdit: (anchor: HTMLElement) => void
  onRemove: () => void
}) {
  const { t } = useTranslation()
  const editRef = useRef<HTMLButtonElement>(null)
  return (
    <div className="flex h-8 items-center gap-1 rounded-md border border-input bg-background px-2 text-xs">
      <span className="font-medium text-foreground/80">{chip.label}</span>
      <button
        ref={editRef}
        type="button"
        onClick={() => editRef.current && onEdit(editRef.current)}
        className="ml-1 rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
        aria-label={t('dashboards.filters.editAria') ?? 'Edit filter'}
        title={t('dashboards.filters.editAria') ?? 'Edit filter'}
      >
        <Pencil size={12} />
      </button>
      <button
        type="button"
        onClick={onRemove}
        className="rounded p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
        aria-label={t('dashboards.filters.removeAria') ?? 'Remove filter'}
        title={t('dashboards.filters.removeAria') ?? 'Remove filter'}
      >
        <Trash2 size={12} />
      </button>
    </div>
  )
}

function ChipEditorPopover({
  anchor,
  initial,
  busy,
  onCancel,
  onSubmit,
}: {
  anchor: HTMLElement
  initial: DashboardFilterChip | null
  busy: boolean
  onCancel: () => void
  onSubmit: (chip: DashboardFilterChip) => void
}) {
  const popoverRef = useRef<HTMLDivElement>(null)
  const [coords, setCoords] = useState<{ top: number; left: number; width: number } | null>(
    null
  )

  // Anchor bottom-right of the trigger. Clamp so the popover never spills off
  // the right edge. Portalled to <body> to escape the surrounding <main>'s
  // overflow, which would otherwise clip a popover positioned outside it.
  useLayoutEffect(() => {
    const compute = () => {
      const rect = anchor.getBoundingClientRect()
      const availableWidth = window.innerWidth - VIEWPORT_MARGIN * 2
      const width = Math.min(POPOVER_WIDTH, availableWidth)
      let left = rect.right + POPOVER_GAP
      const top = rect.bottom + POPOVER_GAP
      if (left + width > window.innerWidth - VIEWPORT_MARGIN) {
        left = Math.max(VIEWPORT_MARGIN, window.innerWidth - VIEWPORT_MARGIN - width)
      }
      setCoords({ top, left, width })
    }
    compute()
    window.addEventListener('resize', compute)
    window.addEventListener('scroll', compute, true)
    return () => {
      window.removeEventListener('resize', compute)
      window.removeEventListener('scroll', compute, true)
    }
  }, [anchor])

  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      const target = e.target as Node
      if (anchor.contains(target) || popoverRef.current?.contains(target)) return
      onCancel()
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [anchor, onCancel])

  if (!coords) return null

  return createPortal(
    <div
      ref={popoverRef}
      className="fixed z-50"
      style={{ top: coords.top, left: coords.left, width: coords.width }}
    >
      <DashboardFilterChipEditor
        initial={initial}
        busy={busy}
        onCancel={onCancel}
        onSubmit={onSubmit}
      />
    </div>,
    document.body
  )
}

function FilterChip({
  chip,
  value,
  onChange,
  clearLabel,
}: {
  chip: DashboardFilterChip
  value: string | string[] | undefined
  onChange: (next: string | string[] | null) => void
  clearLabel: string
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const rootRef = useRef<HTMLDivElement | null>(null)

  const query = usePropertyValues(chip.indexPattern, chip.field, open)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const hasValue = Array.isArray(value) ? value.length > 0 : !!value
  const displayValue = Array.isArray(value)
    ? value.length <= 2
      ? value.join(', ')
      : `${value.length} ${t('dashboards.filters.selected')}`
    : (value ?? '')

  const items = (query.data ?? []).filter((v) =>
    chip.searchable && search ? v.toLowerCase().includes(search.toLowerCase()) : true
  )

  const toggle = (v: string) => {
    if (chip.multiple) {
      const cur = Array.isArray(value) ? value : []
      const next = cur.includes(v) ? cur.filter((x) => x !== v) : [...cur, v]
      onChange(next)
      return
    }
    onChange(v)
    setOpen(false)
  }

  return (
    <div ref={rootRef} className="relative">
      <div
        className={cn(
          'flex h-8 items-center gap-2 rounded-md border border-input bg-background pl-2 pr-1 text-xs',
          hasValue && 'border-primary/60'
        )}
      >
        <button
          type="button"
          onClick={() => setOpen((o) => !o)}
          className="flex items-center gap-1.5"
        >
          <span className="font-medium text-foreground/80">{chip.label}</span>
          {hasValue ? (
            <span className="max-w-[16ch] truncate text-primary">{String(displayValue)}</span>
          ) : (
            <span className="text-muted-foreground">
              {chip.placeholder ?? t('dashboards.filters.chooseValue')}
            </span>
          )}
          <ChevronDown size={12} className="text-muted-foreground" />
        </button>
        {hasValue && (
          <button
            type="button"
            onClick={() => onChange(null)}
            aria-label={clearLabel}
            className="ml-1 rounded p-0.5 text-muted-foreground hover:bg-muted"
          >
            <X size={12} />
          </button>
        )}
      </div>

      {open && (
        <div className="absolute left-0 top-9 z-20 w-64 rounded-md border border-border bg-popover p-2 shadow-md">
          {chip.searchable && (
            <input
              autoFocus
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('dashboards.filters.searchPlaceholder')}
              className="mb-2 h-7 w-full rounded border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          )}
          <div className="max-h-56 overflow-auto">
            {query.isLoading && (
              <div className="flex items-center gap-2 px-1 py-1 text-xs text-muted-foreground">
                <Loader2 size={12} className="animate-spin" />
                {t('dashboards.filters.loading')}
              </div>
            )}
            {query.isError && (
              <div className="px-1 py-1 text-xs text-destructive">
                {t('dashboards.filters.loadError')}
              </div>
            )}
            {!query.isLoading && items.length === 0 && (
              <div className="px-1 py-1 text-xs text-muted-foreground">
                {t('dashboards.filters.noValues')}
              </div>
            )}
            {items.map((v) => {
              const selected = Array.isArray(value) ? value.includes(v) : value === v
              return (
                <button
                  key={v}
                  type="button"
                  onClick={() => toggle(v)}
                  className={cn(
                    'flex w-full items-center gap-2 rounded px-1.5 py-1 text-left text-xs hover:bg-muted',
                    selected && 'bg-muted'
                  )}
                >
                  {chip.multiple && (
                    <input
                      type="checkbox"
                      readOnly
                      checked={selected}
                      className="pointer-events-none"
                    />
                  )}
                  <span className="truncate">{v}</span>
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
