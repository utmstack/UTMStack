import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// One artifact per entry in the chain, placed in a row: an icon box (folder or
// terminal) that fills with the artifact name when it is written, a burst where
// it lands, and the host flagged at the end.
export function Persistence(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { chain: { label: string; kind: 'file' | 'cmd'; detail?: string }[] }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const step = Math.min(2400, 6200 / f.chain.length)
  const last = f.chain.length - 1
  const rows = f.chain.map((_, i) => ({ x: 400 + i * 40, y: 130 + i * (220 / Math.max(1, f.chain.length - 1)) }))
  const done = t > last * step + 900
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scPersist')} tone={tone.muted} />
      {f.chain.map((p, i) => {
        const appear = i * step
        const o = prog(t, appear, appear + 500)
        if (o <= 0) return null
        const r = rows[i]
        const written = t > appear + 600
        const ink = written ? tone.alert : tone.textMuted
        const w = Math.max(150, p.label.length * 7.6 + 46)
        return (
          <g key={i} opacity={o}>
            {i > 0 && <path d={`M${rows[i - 1].x + 18},${rows[i - 1].y + 18} V${r.y} H${r.x}`} fill="none" stroke={ink} strokeOpacity={0.6} strokeWidth={1.8} strokeDasharray="5 6" strokeDashoffset={-t / 30} />}
            <g transform={`translate(${r.x},${r.y - 18})`}>
              <rect width={w} height={36} rx={9} fill={tone.surface} stroke={ink} strokeWidth={written ? 2.2 : 1.4} filter={written && i === last ? glow : undefined} />
              <text x={14} y={23} fill={written ? tone.alertText : tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
                {p.label.length > 34 ? `${p.label.slice(0, 33)}…` : p.label}
              </text>
            </g>
            {written && i === last && <Burst at={r} t={t} start={appear + 600} tone={tone.alert} size={70} dur={900} />}
          </g>
        )
      })}
      {done && <text x={560} y={430} textAnchor="middle" fill={tone.okText} fontSize={12} fontFamily="JetBrains Mono, monospace" opacity={prog(t, last * step + 900, last * step + 1400)}>{tr('liveOps.scReboot')}</text>}
      <MachineNode party={host} at={H0} t={t} state={done ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8600} />
    </g>
  )
}
