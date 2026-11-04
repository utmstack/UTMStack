import { usePalette } from '@/features/liveops/hooks/usePalette'

const MONO = 'JetBrains Mono, monospace'

export function Pill({ x, y, text, tone, small, mono, opacity = 1, anchor = 'middle' }: { x: number; y: number; text: string; tone: string; small?: boolean; mono?: boolean; opacity?: number; anchor?: 'middle' | 'start' | 'end' }) {
  const p = usePalette()
  if (opacity <= 0) return null
  const fs = small ? 10.5 : 12
  const w = Math.max(28, text.length * (mono ? fs * 0.62 : fs * 0.56) + 18)
  const h = small ? 19 : 24
  const left = anchor === 'middle' ? x - w / 2 : anchor === 'start' ? x : x - w
  return (
    <g opacity={opacity} className="lo-pop">
      <rect x={left} y={y - h / 2} width={w} height={h} rx={h / 2} fill={p.surfaceAlt} stroke={tone} strokeOpacity={0.75} />
      <text x={left + w / 2} y={y + fs * 0.36} textAnchor="middle" fill={tone} fontSize={fs} fontWeight={600} fontFamily={mono ? MONO : undefined}>
        {text}
      </text>
    </g>
  )
}
