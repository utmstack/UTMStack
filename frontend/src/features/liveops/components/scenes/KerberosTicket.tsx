import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// A host asks the DC for service tickets; the forged/stolen ticket is a
// highlight-coloured token that keeps working, letting the host move between
// services with it.
export function KerberosTicket(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { ticket: string; kind: string; services: string[] }
  const src = party(scene, 'source')!
  const dc = party(scene, 'dc')!
  const S: Pt = { x: 800, y: 260 }
  const C: Pt = { x: 200, y: 240 }
  const ask = stream(t, 800, 3400, 1000, 900)
  const forged = t > 3400
  const hopEnd = before(7600, props.blockedAt)
  const hops = stream(t, 5200, hopEnd, 1000, 1200)
  return (
    <g>
      <Zone x={36} y={70} w={330} h={360} label={tr('liveOps.scDc')} tone={tone.internal} />
      <Zone x={600} y={70} w={364} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={C} tone={forged ? tone.highlight : tone.attacker} t={t} />
      {ask.map((p) => (
        <Packet key={`a${p.i}`} a={S} b={C} u={p.u} tone={tone.attacker} r={3.5} label={p.u < 0.18 ? tr('liveOps.scTktAsk') : undefined} />
      ))}
      {ask[ask.length - 1] && !forged && <Packet a={C} b={S} u={1 - ask[ask.length - 1].u} tone={tone.highlight} r={4.5} />}
      {forged && <Burst at={S} t={t} start={3400} tone={tone.highlight} size={60} dur={900} />}
      {forged && <Pill x={(S.x + C.x) / 2} y={180} text={`${f.ticket} · ${f.kind}`} tone={tone.highlight} mono small opacity={prog(t, 3400, 4000)} />}
      {forged && (
        <g opacity={prog(t, 3600, 4200)} transform={`translate(${S.x + 40},${S.y + 78})`}>
          <rect x={-120} y={-16} width={240} height={32} rx={7} fill={tone.surfaceAlt} stroke={tone.highlight} strokeWidth={1.4} filter={glow} />
          <text y={4} textAnchor="middle" fill={tone.highlight} fontSize={11} fontFamily="JetBrains Mono, monospace">
            {tr('liveOps.scTktUses', { n: f.services.length })}
          </text>
        </g>
      )}
      {hops.map((p) => (
        <Packet key={`h${p.i}`} a={S} b={C} u={p.u} tone={tone.highlight} r={3.5} label={p.u < 0.15 ? f.services[p.i % f.services.length] : undefined} />
      ))}
      <MachineNode party={dc} at={C} t={t} state={forged ? 'flagged' : 'active'} selected={props.selectedId === dc.id} onSelect={sel(props, dc)} />
      <MachineNode party={src} at={S} t={t} state={forged ? 'compromised' : 'active'} selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <AlertBell at={{ x: S.x + 64, y: S.y - 118 }} t={t} start={before(8400, (props.blockedAt ?? 99999))} />
    </g>
  )
}
