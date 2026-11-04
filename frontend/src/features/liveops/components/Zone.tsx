import { usePalette } from '@/features/liveops/hooks/usePalette'

// A labelled area of the network (Internet, perimeter, internal network, cloud).
export function Zone({ x, y, w, h, label, tone }: { x: number; y: number; w: number; h: number; label: string; tone: string }) {
  const p = usePalette()
  return (
    <g>
      <rect x={x} y={y} width={w} height={h} rx={18} fill={tone} fillOpacity={0.035} stroke={tone} strokeOpacity={p.dark ? 0.22 : 0.35} strokeDasharray="6 6" />
      <text x={x + 16} y={y + 24} fill={tone} fillOpacity={p.dark ? 0.75 : 0.9} fontSize={11} fontWeight={600} letterSpacing="0.12em">
        {label.toUpperCase()}
      </text>
    </g>
  )
}
