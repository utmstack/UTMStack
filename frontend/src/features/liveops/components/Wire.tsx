import { curvePath, prog, type Pt } from './geometry'

export function Wire({ a, b, tone, t, bend = 0, opacity = 0.45, solid, appear = 0 }: { a: Pt; b: Pt; tone: string; t: number; bend?: number; opacity?: number; solid?: boolean; appear?: number }) {
  const o = prog(t, appear, appear + 600)
  if (o <= 0) return null
  return (
    <path
      d={curvePath(a, b, bend)}
      fill="none"
      stroke={tone}
      strokeOpacity={opacity * o}
      strokeWidth={solid ? 2.2 : 1.6}
      strokeDasharray={solid ? undefined : '5 7'}
      strokeDashoffset={solid ? undefined : -t / 28}
    />
  )
}
