import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function Outbound(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { port: number; blocked: boolean; kind: string }
  const src = party(scene, 'source')!
  const fw = party(scene, 'perimeter')
  const dst = party(scene, 'destination')!
  const S: Pt = { x: 840, y: 270 }
  const D: Pt = { x: 155, y: 230 }
  // With no firewall in the scene the connection is drawn straight source → destination.
  const P: Pt = { x: 500, y: 250 }
  const hopA = S
  const hopB = fw ? P : D
  const midX = fw ? (P.x + D.x) / 2 : (S.x + D.x) / 2
  const dns = prog(t, 400, 2400)
  const conn = stream(t, 2600, 5400, 700, 1600)
  const dropped = f.blocked || (props.blockedAt !== undefined && t >= props.blockedAt)
  const later = stream(t, 8400, 1e12, dropped ? 1700 : 1500, dropped ? 900 : 1600)
  const matched = t > 5600
  return (
    <g>
      <Zone x={36} y={70} w={290} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={650} y={70} w={fw ? 314 : 214} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={fw ? P : D} tone={matched ? tone.attacker : tone.internal} t={t} />
      {fw && <Wire a={P} b={D} tone={matched ? tone.attacker : tone.muted} t={t} />}
      {dns > 0 && dns < 1 && <Packet a={S} b={fw ? P : D} u={dns} tone={tone.textSoft} r={4} label={tr('liveOps.scDns', { host: dst.name.length > 26 ? dst.name.slice(0, 25) + '…' : dst.name })} />}
      {conn.map((p) => {
        if (!fw) return <Packet key={`c${p.i}`} a={hopA} b={hopB} u={p.u} tone={tone.internal} />
        return p.u < 0.5 ? (
          <Packet key={`c${p.i}`} a={S} b={P} u={p.u * 2} tone={tone.internal} />
        ) : (
          <Packet key={`c${p.i}`} a={P} b={D} u={(p.u - 0.5) * 2} tone={tone.internal} bend={20} />
        )
      })}
      {t > 2600 && t < 5600 && <Pill x={midX} y={175} text={`TCP ${f.port}`} tone={tone.internal} mono small />}
      {later.map((p) => {
        if (!fw) return dropped ? <Packet key={`l${p.i}`} a={hopA} b={hopB} u={p.u} tone={tone.internal} /> : <Packet key={`l${p.i}`} a={hopA} b={hopB} u={p.u} tone={tone.alert} label={p.u < 0.2 ? tr('liveOps.scBeacon') : undefined} bend={20} />
        return dropped ? (
          <Packet key={`l${p.i}`} a={S} b={P} u={p.u} tone={tone.internal} />
        ) : p.u < 0.5 ? (
          <Packet key={`l${p.i}`} a={S} b={P} u={p.u * 2} tone={tone.alert} label={p.u < 0.2 ? tr('liveOps.scBeacon') : undefined} />
        ) : (
          <Packet key={`l${p.i}`} a={P} b={D} u={(p.u - 0.5) * 2} tone={tone.alert} bend={20} />
        )
      })}
      {matched && <Burst at={D} t={t} start={5600} tone={tone.alert} size={70} dur={1100} />}
      {matched && <Pill x={D.x} y={D.y - 72} text={tr('liveOps.scThreatIntel')} tone={tone.alert} small />}
      {matched && <Pill x={D.x} y={D.y + 104} text={f.kind} tone={tone.attacker} small />}
      {fw && dropped && t > 8400 && <Pill x={P.x} y={P.y - 66} text={tr('liveOps.scConnDropped')} tone={tone.ok} small />}
      <MachineNode party={dst} at={D} t={t} state={matched ? 'flagged' : 'idle'} selected={props.selectedId === dst.id} onSelect={sel(props, dst)} />
      {fw && <MachineNode party={fw} at={P} t={t} state={dropped && t > 8400 ? 'blocked' : 'idle'} selected={props.selectedId === fw.id} onSelect={sel(props, fw)} />}
      <MachineNode party={src} at={S} t={t} state={!dropped && t > 8400 ? 'compromised' : 'active'} selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <AlertBell at={{ x: S.x + 64, y: S.y - 118 }} t={t} start={8600} />
    </g>
  )
}
