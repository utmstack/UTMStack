import { Bell } from 'lucide-react'
import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { backOut, easeOut, prog, type Pt } from './geometry'
import { Pill } from './Pill'

// The bell that pops when the alert is raised.
export function AlertBell({ at, t, start, label = 'Alert raised', labelAbove = false }: { at: Pt; t: number; start: number; label?: string; labelAbove?: boolean }) {
  const p = usePalette()
  const glow = useGlow(true)
  const u = prog(t, start, start + 550)
  if (u <= 0) return null
  const s = 0.55 + 0.45 * backOut(u)
  const wobble = Math.sin((t - start) / 90) * Math.max(0, 1 - (t - start) / 1200) * 12
  const ripples = [0, 450].map((d) => prog(t, start + 150 + d, start + 1350 + d)).filter((x) => x > 0 && x < 1)
  return (
    <g transform={`translate(${at.x},${at.y})`} opacity={Math.min(1, u * 2)}>
      {ripples.map((x, i) => (
        <circle key={i} r={22 + easeOut(x) * 26} fill="none" stroke={p.alert} strokeWidth={2} strokeOpacity={0.6 * (1 - x)} />
      ))}
      <g transform={`scale(${s})`}>
        <circle r={22} fill={p.alertWash} stroke={p.alert} strokeWidth={2} filter={glow} />
        <g transform={`rotate(${wobble})`}>
          <Bell x={-11} y={-11} width={22} height={22} color={p.alertText} strokeWidth={2.2} />
        </g>
      </g>
      <Pill x={0} y={labelAbove ? -38 : 38} text={label} tone={p.alert} small />
    </g>
  )
}
