import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Burst, AlertBell, Zone } from '..'
import { prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A victim machine: the file pill drops in, gets opened, the app spawns a
// child, and the child turns out to be the payload — a short 2-3 step chain.
export function UserExecution(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { file: string; app: string; payload: string }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  // Chain positions inside the right zone: file, app, payload.
  const F: Pt = { x: 430, y: 160 }
  const A: Pt = { x: 590, y: 280 }
  const P: Pt = { x: 760, y: 390 }
  const open = prog(t, 2200, 3000)
  const spawn = prog(t, 4200, 5000)
  const bad = t > 5400
  const pill = (p: Pt, label: string, on: number, ink: string, filled: boolean) => {
    const o = prog(t, on, on + 500)
    if (o <= 0) return null
    const w = Math.max(150, label.length * 7.6 + 46)
    return (
      <g key={label} opacity={ease(o)} transform={`translate(${p.x},${p.y})`}>
        <rect x={-w / 2} y={-18} width={w} height={36} rx={9} fill={filled ? tone.alertWash : tone.surface} stroke={ink} strokeWidth={filled ? 2.2 : 1.4} filter={filled ? glow : undefined} />
        <text y={5} textAnchor="middle" fill={filled ? tone.alertText : tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
          {label}
        </text>
      </g>
    )
  }
  const link = (a: Pt, b: Pt, from: number, ink: string, dash = true) => {
    const o = prog(t, from, from + 500)
    if (o <= 0) return null
    return <path key={`${a.x}${a.y}`} d={`M${a.x},${a.y + 18} C${a.x + 40},${a.y + 60} ${b.x - 40},${b.y - 60} ${b.x},${b.y - 18}`} fill="none" stroke={ink} strokeOpacity={0.6 * o} strokeWidth={1.8} strokeDasharray={dash ? '5 6' : undefined} strokeDashoffset={dash ? -t / 30 : undefined} />
  }
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scExec')} tone={tone.muted} />
      {link(F, A, 4200, bad ? tone.alert : tone.muted)}
      {link(A, P, 5200, tone.alert, false)}
      {pill(F, f.file, 600, open >= 1 ? tone.highlight : tone.textMuted, false)}
      {open > 0 && open < 1 && <Burst at={F} t={t} start={2200} tone={tone.highlight} size={50} dur={700} />}
      {pill(A, f.app, 4200, tone.muted, false)}
      {pill(P, f.payload, 5200, tone.alert, true)}
      {spawn > 0 && spawn < 1 && <circle cx={F.x + (A.x - F.x) * spawn} cy={F.y + (A.y - F.y) * spawn} r={7} fill={tone.highlight} filter={glow} />}
      {bad && <Burst at={P} t={t} start={5400} tone={tone.alert} size={80} dur={1000} />}
      {bad && t > 6200 && (
        <text x={P.x} y={P.y + 46} textAnchor="middle" fill={tone.alertSoft} fontSize={11.5} fontFamily="JetBrains Mono, monospace" opacity={prog(t, 6200, 6800)}>
          {tr('liveOps.scPayload')}
        </text>
      )}
      <MachineNode party={host} at={H0} t={t} state={bad ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8600} />
    </g>
  )
}
