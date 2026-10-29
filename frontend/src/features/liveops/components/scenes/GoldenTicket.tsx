import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone, Counter } from '..'
import { stream, ease, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function GoldenTicket(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { domain: string; account: string; forged: number; internal: boolean }
  const src = party(scene, 'source')!
  const dc = party(scene, 'target')!
  // The domain controller signs every ticket: it stays inside the network even
  // when the one forging them is an outside attacker.
  const A: Pt = { x: 160, y: 270 }
  const D: Pt = { x: 820, y: 270 }
  const forged = stream(t, 1200, 6200, 950, 1250)
  const used = stream(t, 6800, 1e12, 1150, 1300)
  const detected = t > 6200
  const count = Math.min(f.forged, Math.round(f.forged * ease((t - 1200) / 5000)))
  return (
    <g>
      <Zone x={36} y={70} w={290} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={500} y={70} w={464} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={A} b={D} tone={detected ? tone.alert : tone.internal} t={t} solid={detected} />
      {forged.map((p) => (
        <Packet key={`f${p.i}`} a={A} b={D} u={p.u} tone={tone.highlight} r={5} label={p.u < 0.55 ? tr('liveOps.scForgedTicket', { account: f.account }) : undefined} />
      ))}
      {used.map((p) => (
        <Packet key={`u${p.i}`} a={A} b={D} u={p.u} tone={tone.alert} r={4.5} label={p.u < 0.5 ? tr('liveOps.scTicketUsed', { account: f.account }) : undefined} />
      ))}
      {detected && <Burst at={D} t={t} start={6200} tone={tone.alert} size={64} dur={1100} />}
      {detected && <Pill x={D.x} y={D.y - 70} text={tr('liveOps.scTgtLifetime')} tone={tone.alert} small />}
      {detected && <Pill x={D.x} y={D.y + 104} text={tr('liveOps.scKrbtgtKey')} tone={tone.highlight} small />}
      <Counter x={500} y={452} label={tr('liveOps.scForgedTgts')} value={Math.max(0, count).toLocaleString()} tone={tone.highlight} opacity={ease((t - 1200) / 1500)} />
      <MachineNode party={src} at={A} t={t} state={t > 1000 ? 'active' : 'idle'} selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={dc} at={D} t={t} state={detected ? 'flagged' : t > 1000 ? 'targeted' : 'idle'} selected={props.selectedId === dc.id} onSelect={sel(props, dc)} />
      <AlertBell at={{ x: D.x + 62, y: D.y - 118 }} t={t} start={9000} labelAbove />
    </g>
  )
}
