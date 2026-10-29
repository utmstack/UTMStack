import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone, Counter } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// A signature fires on the perimeter device: attack packets hit the firewall,
// it recognizes the pattern, the stream is cut, the counter tallies hits.
export function PerimeterDetection(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { sig: string; blocked: boolean; hits: number }
  const attacker = party(scene, 'attacker')!
  const fw = party(scene, 'perimeter')!
  const A: Pt = { x: 150, y: 250 }
  const F: Pt = { x: 640, y: 250 }
  const attackEnd = before(f.blocked ? 6200 : 7800, props.blockedAt)
  const attack = stream(t, 400, attackEnd, 260, 700)
  const detected = t > 3000
  const hits = Math.round(f.hits * prog(t, 400, attackEnd))
  return (
    <g>
      <Zone x={36} y={70} w={268} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={430} y={70} w={534} h={360} label={tr('liveOps.netPerimeter')} tone={tone.perimeter} />
      <Wire a={A} b={F} tone={tone.attacker} t={t} />
      {attack.map((p) => (
        <Packet key={p.i} a={A} b={F} u={p.u} tone={tone.attacker} r={3.6} bend={((p.i % 5) - 2) * 10} />
      ))}
      {detected && <Burst at={F} t={t} start={3000} tone={tone.alert} size={60} dur={900} />}
      {detected && <Pill x={F.x} y={F.y - 78} text={f.sig} tone={tone.alert} small mono opacity={prog(t, 3000, 3500)} />}
      {f.blocked && t > 6200 && <Pill x={(A.x + F.x) / 2} y={A.y + 48} text={tr('liveOps.scDroppedAtFw')} tone={tone.ok} small opacity={prog(t, 6200, 6700)} />}
      {f.blocked && t > 6200 && <Pill x={(A.x + F.x) / 2} y={A.y - 48} text={tr('liveOps.scTrafficStops')} tone={tone.ok} small opacity={prog(t, 6800, 7300)} />}
      <Counter x={F.x} y={455} label={tr('liveOps.scSigHits')} value={hits.toLocaleString()} tone={tone.alert} opacity={prog(t, 600, 1000)} />
      <MachineNode party={attacker} at={A} t={t} state={f.blocked && t > 6200 ? 'blocked' : 'active'} selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      <MachineNode party={fw} at={F} t={t} state={detected ? 'targeted' : 'active'} selected={props.selectedId === fw.id} onSelect={sel(props, fw)} />
      <AlertBell at={{ x: F.x + 70, y: F.y - 118 }} t={t} start={before(8000, (props.blockedAt ?? 99999))} />
    </g>
  )
}
