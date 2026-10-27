// Shared geometry and timing helpers for the Live Operations scenes (pure, no components).

export const W = 1000
export const H = 520

export const clamp01 = (x: number) => (x < 0 ? 0 : x > 1 ? 1 : x)
/** 0 before a, 1 after b, linear between. */
export const prog = (t: number, a: number, b: number) => clamp01((t - a) / (b - a))
export const ease = (x: number) => (x < 0.5 ? 4 * x * x * x : 1 - Math.pow(-2 * x + 2, 3) / 2)
export const lerp = (a: number, b: number, u: number) => a + (b - a) * u
/** Ease out with a small overshoot, for things that pop into place. */
export const backOut = (x: number) => {
  const c1 = 1.70158
  return 1 + (c1 + 1) * Math.pow(x - 1, 3) + c1 * Math.pow(x - 1, 2)
}
export const easeOut = (x: number) => 1 - Math.pow(1 - x, 3)

export interface Pt {
  x: number
  y: number
}

/** A point on the curve from a to b (bend lifts the middle of the curve). */
export function along(a: Pt, b: Pt, u: number, bend = 0): Pt {
  const mx = (a.x + b.x) / 2
  const my = (a.y + b.y) / 2 - bend
  const v = 1 - u
  return { x: v * v * a.x + 2 * v * u * mx + u * u * b.x, y: v * v * a.y + 2 * v * u * my + u * u * b.y }
}

export function curvePath(a: Pt, b: Pt, bend = 0) {
  return `M${a.x},${a.y} Q${(a.x + b.x) / 2},${(a.y + b.y) / 2 - bend} ${b.x},${b.y}`
}

/** Positions of packets launched every `every` ms from `start` to `stop`, each taking `travel` ms. */
export function stream(t: number, start: number, stop: number, every: number, travel: number): { u: number; i: number; age: number }[] {
  const out: { u: number; i: number; age: number }[] = []
  if (t < start) return out
  const first = Math.max(0, Math.floor((t - start - travel) / every))
  const last = Math.floor((Math.min(t, stop) - start) / every)
  for (let i = first; i <= last; i++) {
    const launched = start + i * every
    const age = t - launched
    if (age < 0 || age > travel) continue
    out.push({ u: age / travel, i, age })
  }
  return out
}
