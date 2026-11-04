import { usePalette } from '@/features/liveops/hooks/usePalette'

const MONO = 'JetBrains Mono, monospace'

export function Counter({ x, y, label, value, tone, opacity = 1 }: { x: number; y: number; label: string; value: string; tone: string; opacity?: number }) {
  const p = usePalette()
  if (opacity <= 0) return null
  return (
    <g opacity={opacity} transform={`translate(${x},${y})`}>
      <rect x={-92} y={-30} width={184} height={60} rx={12} fill={p.surfaceAlt} stroke={tone} strokeOpacity={0.5} />
      <text y={-9} textAnchor="middle" fill={p.textMuted} fontSize={10.5} letterSpacing="0.08em">
        {label.toUpperCase()}
      </text>
      <text y={18} textAnchor="middle" fill={tone} fontSize={24} fontWeight={700} fontFamily={MONO}>
        {value}
      </text>
    </g>
  )
}
