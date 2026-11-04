import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, AlertBell, Zone } from '..'
import { stream, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function Lateral(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { account: string }
  const from = party(scene, 'source')!
  const to = party(scene, 'target')!
  const F: Pt = { x: 230, y: 285 }
  const T: Pt = { x: 770, y: 285 }
  const names = ['fs-01', 'wks-0211', 'sql-02', 'wks-0457', 'app-03'].filter((n) => n !== from.name && n !== to.name)
  const others: (Pt & { name: string })[] = [
    { x: 500, y: 140, name: names[0] },
    { x: 380, y: 420, name: names[1] },
    { x: 620, y: 420, name: names[2] },
  ]
  const keys = stream(t, 3000, 6400, 900, 1300)
  const owned = t > 6400
  return (
    <g>
      <Zone x={60} y={70} w={880} h={410} label={tr('liveOps.netInternal')} tone={tone.internal} />
      {others.map((o, i) => (
        <g key={i} opacity={0.45}>
          <Wire a={F} b={o} tone={tone.muted} t={t} opacity={0.25} />
          <circle cx={o.x} cy={o.y} r={16} fill={tone.surface} stroke={tone.muted} />
          <rect x={o.x - 7} y={o.y - 6} width={14} height={10} rx={1.5} fill="none" stroke={tone.textMuted} strokeWidth={1.4} />
          <line x1={o.x - 4} y1={o.y + 7} x2={o.x + 4} y2={o.y + 7} stroke={tone.textMuted} strokeWidth={1.4} />
          <text x={o.x} y={o.y + 32} textAnchor="middle" fill={tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
            {o.name}
          </text>
        </g>
      ))}
      <Wire a={F} b={T} tone={owned ? tone.alert : tone.internal} t={t} bend={70} solid={owned} />
      {keys.map((p) => (
        <Packet key={`k${p.i}`} a={F} b={T} u={p.u} tone={tone.highlight} bend={70} r={5} label={p.u < 0.6 ? `🔑 ${f.account}` : undefined} />
      ))}
      {owned && stream(t, 6400, 1e12, 1100, 1400).map((p) => <Packet key={`s${p.i}`} a={F} b={T} u={p.u} tone={tone.alert} bend={70} r={3.5} />)}
      {owned && <Pill x={500} y={214} text={tr('liveOps.scSessionAs', { account: f.account })} tone={tone.alert} />}
      <MachineNode party={from} at={F} t={t} state="compromised" selected={props.selectedId === from.id} onSelect={sel(props, from)} />
      <MachineNode party={to} at={T} t={t} state={owned ? 'compromised' : t > 3000 ? 'targeted' : 'idle'} selected={props.selectedId === to.id} onSelect={sel(props, to)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 112 }} t={t} start={9200} />
    </g>
  )
}
