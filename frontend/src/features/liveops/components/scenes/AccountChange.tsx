import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// An object box (account / group / policy) changes state: the "before" chip
// fades, the "after" chip fills in, a burst marks the change, and the DC/host
// is flagged.
export function AccountChange(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { what: string; before: string; after: string }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const O: Pt = { x: 640, y: 250 }
  const changed = t > 3400
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scObject')} tone={tone.muted} />
      <g opacity={ease(prog(t, 400, 900))} transform={`translate(${O.x},${O.y})`}>
        <rect x={-170} y={-90} width={340} height={180} rx={12} fill={tone.surface} stroke={changed ? tone.alert : tone.line} strokeWidth={changed ? 2 : 1.4} filter={changed ? glow : undefined} />
        <text y={-64} textAnchor="middle" fill={tone.textMuted} fontSize={11} letterSpacing="0.1em">
          {f.what.toUpperCase()}
        </text>
        <text x={-150} y={-20} fill={changed ? tone.textFaint : tone.textSoft} fontSize={12} fontFamily="JetBrains Mono, monospace" opacity={changed ? 0.5 : 1}>
          {f.before.length > 24 ? f.before.slice(0, 23) + '…' : f.before}
        </text>
        {changed && <line x1={-150} y1={-25} x2={150} y2={-25} stroke={tone.alert} strokeWidth={2} opacity={prog(t, 3400, 3900)} />}
        <text x={-150} y={28} fill={tone.alertText} fontSize={12} fontFamily="JetBrains Mono, monospace" fontWeight={600} opacity={prog(t, 3400, 4000)}>
          {f.after.length > 24 ? f.after.slice(0, 23) + '…' : f.after}
        </text>
        <text y={64} textAnchor="middle" fill={tone.textMuted} fontSize={10.5}>
          {tr('liveOps.scChangedBy')}
        </text>
      </g>
      {changed && <Burst at={O} t={t} start={3400} tone={tone.alert} size={90} dur={1000} />}
      {changed && t > 5200 && <Pill x={O.x} y={O.y + 120} text={tr('liveOps.scHighRisk')} tone={tone.alert} opacity={prog(t, 5200, 5800)} />}
      <MachineNode party={host} at={H0} t={t} state={changed ? 'flagged' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8000} />
    </g>
  )
}
