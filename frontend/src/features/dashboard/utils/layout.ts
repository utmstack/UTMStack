import { GRID_COLS } from '@/features/dashboard/constants'
import type { DashboardVisualization, GridLayoutItem } from '@/features/dashboard/types'

// Widgets live on a 12-column react-grid-layout: free x/y position and w/h size,
// all in grid units. Stored layout JSON is `{ x, y, w, h }`.
//
// Backward compatibility: older dashboards stored `{ order, w, h }` (a 1–3 column
// span + a 1/2 height preset, with no x/y). Those are migrated on read — the span
// is scaled up to the 12-col grid and the items are shelf-packed by their order.

export interface StoredLayout {
  x: number
  y: number
  w: number
  h: number
}

const DEFAULT_W = 4
const DEFAULT_H = 8

function clampInt(v: unknown, min: number, max: number, fallback: number): number {
  const n = Math.round(Number(v))
  return Number.isFinite(n) ? Math.min(max, Math.max(min, n)) : fallback
}

interface ParsedLayout extends StoredLayout {
  /** true when the stored layout already carries an explicit x/y position. */
  positioned: boolean
  order: number
}

function parseStored(raw: string | null | undefined, fallbackOrder: number): ParsedLayout {
  if (raw) {
    try {
      const p = JSON.parse(raw) as Record<string, unknown>
      // Current format: explicit x/y/w/h.
      if (
        typeof p.x === 'number' &&
        typeof p.y === 'number' &&
        typeof p.w === 'number' &&
        typeof p.h === 'number'
      ) {
        return {
          x: clampInt(p.x, 0, GRID_COLS - 1, 0),
          y: clampInt(p.y, 0, 100000, 0),
          w: clampInt(p.w, 1, GRID_COLS, DEFAULT_W),
          h: clampInt(p.h, 1, 100000, DEFAULT_H),
          positioned: true,
          order: typeof p.order === 'number' ? p.order : fallbackOrder,
        }
      }
      // Legacy format: { order, w (1–3), h (1–2) } → scale onto the 12-col grid.
      const w = clampInt((Number(p.w) || 1) * (GRID_COLS / 3), 1, GRID_COLS, DEFAULT_W)
      const h = Number(p.h) === 2 ? 10 : 6
      return {
        x: 0,
        y: 0,
        w,
        h,
        positioned: false,
        order: typeof p.order === 'number' ? p.order : fallbackOrder,
      }
    } catch {
      /* fall through to default */
    }
  }
  return { x: 0, y: 0, w: DEFAULT_W, h: DEFAULT_H, positioned: false, order: fallbackOrder }
}

export function parseLayout(raw: string | null | undefined, fallbackOrder = 0): StoredLayout {
  const { x, y, w, h } = parseStored(raw, fallbackOrder)
  return { x, y, w, h }
}

export function serializeLayout(l: { x: number; y: number; w: number; h: number }): string {
  return JSON.stringify({
    x: clampInt(l.x, 0, GRID_COLS - 1, 0),
    y: clampInt(l.y, 0, 100000, 0),
    w: clampInt(l.w, 1, GRID_COLS, DEFAULT_W),
    h: clampInt(l.h, 1, 100000, DEFAULT_H),
  })
}

function collides(a: GridLayoutItem, b: GridLayoutItem): boolean {
  return a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h
}

// Vertical compaction: pull every item up to the lowest non-colliding row. Keeps
// the layout gap-free and, crucially, matches react-grid-layout's own vertical
// compaction so we don't get a spurious "dirty" change the moment edit starts.
function compactVertical(items: GridLayoutItem[]): GridLayoutItem[] {
  const sorted = [...items].sort((a, b) => a.y - b.y || a.x - b.x)
  const placed: GridLayoutItem[] = []
  for (const it of sorted) {
    const candidate: GridLayoutItem = { ...it, y: 0 }
    for (;;) {
      const hit = placed.find((p) => collides(candidate, p))
      if (!hit) break
      candidate.y = hit.y + hit.h
    }
    placed.push(candidate)
  }
  return placed
}

// Shelf-pack loose (un-positioned / legacy) items left-to-right, wrapping at the
// column count, starting below any already-positioned rows.
function pack(loose: { i: string; w: number; h: number }[], startY: number): GridLayoutItem[] {
  let x = 0
  let y = startY
  let rowH = 0
  return loose.map(({ i, w, h }) => {
    if (x + w > GRID_COLS) {
      x = 0
      y += rowH
      rowH = 0
    }
    const placed: GridLayoutItem = { i, x, y, w, h }
    x += w
    rowH = Math.max(rowH, h)
    return placed
  })
}

/** Build react-grid-layout items from stored rows, migrating legacy layouts. */
export function toGridItems(rows: DashboardVisualization[]): GridLayoutItem[] {
  const parsed = rows.map((dv, idx) => ({ dv, p: parseStored(dv.layout, idx) }))

  const positioned: GridLayoutItem[] = parsed
    .filter((r) => r.p.positioned)
    .map((r) => ({ i: String(r.dv.id), x: r.p.x, y: r.p.y, w: r.p.w, h: r.p.h }))

  const loose = parsed
    .filter((r) => !r.p.positioned)
    .sort((a, b) => a.p.order - b.p.order || a.dv.id - b.dv.id)
    .map((r) => ({ i: String(r.dv.id), w: r.p.w, h: r.p.h }))

  const startY = positioned.reduce((max, it) => Math.max(max, it.y + it.h), 0)
  return compactVertical([...positioned, ...pack(loose, startY)])
}

/** Lowest free row below all items — where a freshly added widget should land. */
export function nextRow(items: GridLayoutItem[]): number {
  return items.reduce((max, it) => Math.max(max, it.y + it.h), 0)
}
