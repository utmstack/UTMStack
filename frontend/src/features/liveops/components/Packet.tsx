import { useGlow } from '@/features/liveops/hooks/useGlow'
import { along, type Pt } from './geometry'
import { Pill } from './Pill'

export function Packet({ a, b, u, tone, bend = 0, r = 4.5, label }: { a: Pt; b: Pt; u: number; tone: string; bend?: number; r?: number; label?: string }) {
  const glow = useGlow()
  const p = along(a, b, u, bend)
  const mid = along(a, b, Math.max(0, u - 0.03), bend)
  const tail = along(a, b, Math.max(0, u - 0.07), bend)
  // Packets fade in as they leave and out as they arrive, instead of blinking on and off.
  const fade = Math.min(1, u / 0.06, (1 - u) / 0.06)
  return (
    <g opacity={Math.max(0, fade)}>
      <line x1={tail.x} y1={tail.y} x2={mid.x} y2={mid.y} stroke={tone} strokeOpacity={0.22} strokeWidth={r * 0.8} strokeLinecap="round" />
      <line x1={mid.x} y1={mid.y} x2={p.x} y2={p.y} stroke={tone} strokeOpacity={0.5} strokeWidth={r * 0.95} strokeLinecap="round" />
      <circle cx={p.x} cy={p.y} r={r} fill={tone} filter={glow} />
      {label && <Pill x={p.x} y={p.y - 18} text={label} tone={tone} small />}
    </g>
  )
}
