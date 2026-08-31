import type { FlowNode } from '../types/soar.types'

export const TRIGGER_LAYOUT_ID = '__trigger__'

const LAYER_Y_STEP = 200
const NODE_X_STEP = 320
const CANVAS_X_CENTER = 300
const TRIGGER_Y = 0

/** Top-down layered layout: trigger at layer 0, roots at layer 1, and each
 *  downstream node placed one layer below its deepest parent (longest-path
 *  toposort). Ids in the same layer are spread horizontally, centered around
 *  CANVAS_X_CENTER, so trigger → 1st layer → 2nd layer reads top-to-bottom
 *  with clear separation. Nodes unreachable from the trigger fall into their
 *  own trailing layer instead of piling on top of the graph. */
export function computeLayeredLayout(
  roots: string[],
  nodes: Record<string, FlowNode>,
): Record<string, { x: number; y: number }> {
  const parents: Record<string, string[]> = { [TRIGGER_LAYOUT_ID]: [] }
  for (const id of Object.keys(nodes)) parents[id] = []
  for (const r of roots) if (parents[r]) parents[r].push(TRIGGER_LAYOUT_ID)
  for (const [id, n] of Object.entries(nodes)) {
    for (const c of n.onSuccess ?? []) if (parents[c]) parents[c].push(id)
    for (const c of n.onError ?? []) if (parents[c]) parents[c].push(id)
  }

  const layer: Record<string, number> = { [TRIGGER_LAYOUT_ID]: 0 }
  // Iterate to a fixed point: each pass promotes a node to max(parent)+1 if
  // any parent settled deeper. Caps at nodeCount+1 passes so cycles can't spin.
  const cap = Object.keys(nodes).length + 2
  for (let pass = 0; pass < cap; pass++) {
    let changed = false
    for (const id of Object.keys(nodes)) {
      const ps = parents[id]
      let best = -Infinity
      for (const p of ps) if (layer[p] !== undefined && layer[p] > best) best = layer[p]
      if (best === -Infinity) continue
      const next = best + 1
      if ((layer[id] ?? -1) < next) {
        layer[id] = next
        changed = true
      }
    }
    if (!changed) break
  }

  const reachedMax = Math.max(0, ...Object.values(layer))
  const orphanLayer = reachedMax + 1
  for (const id of Object.keys(nodes)) if (layer[id] === undefined) layer[id] = orphanLayer

  const byLayer: Record<number, string[]> = {}
  for (const [id, l] of Object.entries(layer)) (byLayer[l] ??= []).push(id)

  const positions: Record<string, { x: number; y: number }> = {}
  for (const [lStr, ids] of Object.entries(byLayer)) {
    const l = Number(lStr)
    ids.sort()
    const spanW = (ids.length - 1) * NODE_X_STEP
    const startX = CANVAS_X_CENTER - spanW / 2
    ids.forEach((id, i) => {
      positions[id] = { x: startX + i * NODE_X_STEP, y: TRIGGER_Y + l * LAYER_Y_STEP }
    })
  }
  return positions
}
