import { easeOut, prog, type Pt } from './geometry'

// An expanding ring where something hits.
export function Burst({ at, t, start, tone, size = 46, dur = 700 }: { at: Pt; t: number; start: number; tone: string; size?: number; dur?: number }) {
  const u = prog(t, start, start + dur)
  if (u <= 0 || u >= 1) return null
  const k = easeOut(u)
  return (
    <g>
      <circle cx={at.x} cy={at.y} r={8 + size * 0.45 * k} fill={tone} fillOpacity={0.18 * (1 - u)} />
      <circle cx={at.x} cy={at.y} r={10 + size * k} fill="none" stroke={tone} strokeWidth={2.5 * (1 - u * 0.6)} strokeOpacity={1 - u} />
    </g>
  )
}
