import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, AlertBell, Zone } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function CloudChange(props: SceneProps) {
  const { scene, t } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { what: string; before: string; after: string; risk: string }
  const actor = party(scene, 'actor')!
  const cloud = party(scene, 'cloud')!
  const A: Pt = { x: 150, y: 260 }
  const C: Pt = { x: 450, y: 260 }
  const flip = prog(t, 6000, 6600)
  const card = prog(t, 3000, 3500)
  const req = stream(t, 600, 5800, 1300, 1000)
  return (
    <g>
      <Zone x={36} y={70} w={250} h={380} label={actor.role === 'internal' ? 'Internal' : 'Internet'} tone={actor.role === 'internal' ? tone.internal : tone.attacker} />
      <Zone x={330} y={70} w={634} h={380} label={cloud.name} tone={tone.cloud} />
      <Wire a={A} b={C} tone={tone.cloud} t={t} />
      {req.map((p) => (
        <Packet key={p.i} a={A} b={C} u={p.u} tone={tone.cloud} />
      ))}
      {card > 0 && (
        <g opacity={card} transform={`translate(600,${160 - 10 * (1 - card)})`}>
          <rect width={330} height={200} rx={14} fill={tone.surface} stroke={flip > 0.5 ? tone.alert : tone.cloud} strokeWidth={flip > 0.5 ? 2.2 : 1.4} filter={flip > 0.5 ? glow : undefined} />
          <text x={20} y={34} fill={tone.textMuted} fontSize={11} letterSpacing="0.1em">
            {f.what.toUpperCase().slice(0, 40)}
          </text>
          <text x={20} y={78} fill={flip > 0.5 ? tone.textDim : tone.textSoft} fontSize={16} fontWeight={600} textDecoration={flip > 0.5 ? 'line-through' : undefined}>
            {f.before}
          </text>
          {flip > 0 && (
            <text x={20} y={118} fill={tone.alertSoft} fontSize={17} fontWeight={700} opacity={flip}>
              → {f.after.length > 30 ? `${f.after.slice(0, 29)}…` : f.after}
            </text>
          )}
          {flip >= 1 && (
            <foreignObject x={20} y={134} width={292} height={60}>
              <div style={{ color: tone.textSoft, fontSize: 12, lineHeight: 1.4 }}>{f.risk}</div>
            </foreignObject>
          )}
        </g>
      )}
      <MachineNode party={actor} at={A} t={t} state="active" selected={props.selectedId === actor.id} onSelect={sel(props, actor)} />
      <MachineNode party={cloud} at={C} t={t} state={flip > 0.5 ? 'flagged' : 'targeted'} selected={props.selectedId === cloud.id} onSelect={sel(props, cloud)} />
      <AlertBell at={{ x: 900, y: 122 }} t={t} start={9000} labelAbove />
    </g>
  )
}
