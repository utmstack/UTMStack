import type { DashboardVisualization, GridLayoutItem } from '@/features/dashboard/types'

// Widgets are laid out in a responsive grid ordered by `order`, with a per-widget
// size preset: w = column span (1–3), h = height preset (1 normal, 2 tall).
// Stored layout JSON is `{ order, w, h }`.

export interface StoredLayout {
  order: number
  w: number
  h: number
}

function clampW(v: unknown): number {
  const n = Math.round(Number(v))
  return n >= 1 && n <= 3 ? n : 1
}
function clampH(v: unknown): number {
  const n = Math.round(Number(v))
  return n >= 1 && n <= 2 ? n : 1
}

export function parseLayout(raw: string | null | undefined, fallbackOrder: number): StoredLayout {
  if (!raw) return { order: fallbackOrder, w: 1, h: 1 }
  try {
    const p = JSON.parse(raw) as { order?: unknown; w?: unknown; h?: unknown }
    return {
      order: typeof p.order === 'number' ? p.order : fallbackOrder,
      w: clampW(p.w),
      h: clampH(p.h),
    }
  } catch {
    return { order: fallbackOrder, w: 1, h: 1 }
  }
}

export function serializeLayout(l: { order: number; w: number; h: number }): string {
  return JSON.stringify({ order: l.order, w: clampW(l.w), h: clampH(l.h) })
}

/** Rows sorted by saved order (fallback to id), carrying their size preset in w/h. */
export function toOrderedItems(rows: DashboardVisualization[]): GridLayoutItem[] {
  return rows
    .map((dv, idx) => ({ dv, l: parseLayout(dv.layout, idx) }))
    .sort((a, b) => a.l.order - b.l.order || a.dv.id - b.dv.id)
    .map(({ dv, l }) => ({ i: String(dv.id), x: 0, y: 0, w: l.w, h: l.h }))
}
