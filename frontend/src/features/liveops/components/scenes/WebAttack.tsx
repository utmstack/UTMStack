import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { prog, along, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'
import { type Payload } from '@/features/liveops/model'

export function WebAttack(props: SceneProps) {
  const { scene, t } = props
  const tone = usePalette()
  const f = scene.facts as { payloads: Payload[]; blocked: boolean }
  const attacker = party(scene, 'attacker')!
  const fw = party(scene, 'perimeter')
  const web = party(scene, 'target')!
  const A: Pt = { x: 140, y: 260 }
  const P: Pt = { x: 480, y: 260 }
  const S: Pt = { x: 830, y: 260 }
  const edge = fw ? P : S // where packets stop: the firewall, or the server when there is none
  const every = 1400
  const sends = f.payloads.map((p, i) => ({ p, i, start: 900 + i * every }))
  const isBlocked = props.blockedAt !== undefined && t >= props.blockedAt
  return (
    <g>
      <Zone x={36} y={70} w={260} h={380} label="Internet" tone={tone.attacker} />
      <Zone x={650} y={70} w={fw ? 314 : 214} h={380} label="Customer portal" tone={tone.internal} />
      {fw ? (
        <>
          <Wire a={A} b={P} tone={tone.attacker} t={t} />
          <Wire a={P} b={S} tone={tone.perimeter} t={t} />
        </>
      ) : (
        <Wire a={A} b={S} tone={tone.attacker} t={t} />
      )}
      {sends.map(({ p, i, start }) => {
        const go = prog(t, start, start + 1100)
        const back = prog(t, start + 1100, start + 1900)
        const lane = 165 + (i % 4) * 30
        const label = p.label.length > 30 ? `${p.label.slice(0, 29)}…` : p.label
        if (go <= 0) return null
        if (f.blocked || (props.blockedAt !== undefined && start >= props.blockedAt)) {
          return (
            <g key={i}>
              {go < 1 && <Packet a={A} b={edge} u={go} tone={tone.attacker} />}
              {go < 1 && <Pill x={along(A, edge, go).x} y={lane} text={label} tone={tone.attacker} mono small />}
              {go >= 1 && back < 1 && <Pill x={edge.x} y={edge.y - 70 - (i % 2) * 26} text={`${p.status} dropped`} tone={tone.ok} small opacity={1 - back} />}
              {go >= 1 && back < 0.2 && <Burst at={edge} t={t} start={start + 1100} tone={tone.ok} size={26} dur={400} />}
            </g>
          )
        }
        const whole = { x: A.x + (S.x - A.x) * go, y: A.y }
        return (
          <g key={i}>
            {go < 1 && <Packet a={A} b={S} u={go} tone={tone.attacker} bend={(i % 3) * 18} />}
            {go < 1 && <Pill x={whole.x} y={lane} text={label} tone={tone.attacker} mono small />}
            {go >= 1 && back < 1 && <Packet a={S} b={A} u={back} tone={p.status === 200 ? tone.alert : tone.warn} r={4} label={String(p.status)} />}
          </g>
        )
      })}
      <MachineNode party={attacker} at={A} t={t} state={isBlocked ? 'blocked' : 'active'} selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      {fw && <MachineNode party={fw} at={P} t={t} state={(f.blocked && t > 2000) || isBlocked ? 'blocked' : 'idle'} selected={props.selectedId === fw.id} onSelect={sel(props, fw)} />}
      <MachineNode party={web} at={S} t={t} state={!f.blocked && t > 6200 ? 'compromised' : t > 2000 ? 'targeted' : 'idle'} selected={props.selectedId === web.id} onSelect={sel(props, web)} />
      <AlertBell at={{ x: S.x + 60, y: S.y - 118 }} t={t} start={9000} />
    </g>
  )
}
