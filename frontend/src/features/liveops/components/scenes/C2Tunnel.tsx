import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone, Counter } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// The infected host sends data disguised as normal protocol traffic (DNS /
// ICMP / SSH tunnel): the pill on the wire names the disguise, a counter shows
// bytes smuggled out, the far end receives it.
export function C2Tunnel(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { proto: string; bytes: number }
  const src = party(scene, 'source')!
  const dst = party(scene, 'destination')!
  const S: Pt = { x: 820, y: 270 }
  const D: Pt = { x: 180, y: 250 }
  const start = 1600
  const stop = before(6800, props.blockedAt)
  const pkts = stream(t, start, stop, 900, 1700)
  const sent = Math.round(f.bytes * prog(t, start, stop))
  return (
    <g>
      <Zone x={36} y={70} w={300} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={D} tone={tone.alert} t={t} />
      {pkts.map((p) => (
        <Packet key={p.i} a={S} b={D} u={p.u} tone={tone.alert} r={4} label={p.u < 0.18 ? f.proto : undefined} bend={16} />
      ))}
      {t > start && t < stop + 500 && <Pill x={500} y={186} text={tr('liveOps.scTunnel', { proto: f.proto })} tone={tone.alert} mono small />}
      <Burst at={D} t={t} start={stop - 100} tone={tone.alert} size={60} dur={900} />
      <Counter x={500} y={455} label={tr('liveOps.scTunneled')} value={sent > 1048576 ? `${(sent / 1048576).toFixed(1)} MB` : `${sent} KB`} tone={tone.alert} opacity={prog(t, start + 200, start + 800)} />
      <MachineNode party={src} at={S} t={t} state="compromised" selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={dst} at={D} t={t} state={t > stop ? 'flagged' : 'idle'} selected={props.selectedId === dst.id} onSelect={sel(props, dst)} />
      <AlertBell at={{ x: D.x + 60, y: D.y - 118 }} t={t} start={8200} />
    </g>
  )
}
