import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// The host reaches for a destination that a threat feed already knows: a
// lookup pill asks the feed, the feed answers "bad", the destination lights up.
export function IocMatch(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { ioc: string; kind: string }
  const src = party(scene, 'source')!
  const dst = party(scene, 'destination')!
  const S: Pt = { x: 800, y: 260 }
  const D: Pt = { x: 220, y: 250 }
  const look = prog(t, 800, 2400)
  const matched = t > 2800
  const cut = props.blockedAt !== undefined && t >= props.blockedAt
  return (
    <g>
      <Zone x={36} y={70} w={330} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={D} tone={matched ? tone.alert : tone.internal} t={t} />
      {look > 0 && look < 1 && <Packet a={S} b={D} u={look} tone={tone.textSoft} r={4} label={tr('liveOps.scFeedAsk', { ioc: f.ioc.length > 26 ? f.ioc.slice(0, 25) + '…' : f.ioc })} />}
      {matched && !cut && <Packet a={D} b={S} u={prog(t, 2800, 3900)} tone={tone.textSoft} r={3.5} label={tr('liveOps.scFeedBad')} />}
      {matched && <Burst at={D} t={t} start={before(2800, props.blockedAt)} tone={tone.alert} size={70} dur={1000} />}
      {matched && <Pill x={D.x} y={D.y - 78} text={tr('liveOps.scFeedHit', { kind: f.kind })} tone={tone.alert} small />}
      {cut && <Pill x={S.x - 10} y={S.y - 78} text={tr('liveOps.scBlocked', { name: dst.name })} tone={tone.ok} />}
      <MachineNode party={src} at={S} t={t} state={cut ? 'blocked' : 'active'} selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={dst} at={D} t={t} r={34} state={matched ? 'flagged' : 'idle'} selected={props.selectedId === dst.id} onSelect={sel(props, dst)} />
      <AlertBell at={{ x: S.x + 64, y: S.y - 118 }} t={t} start={before(7800, (props.blockedAt ?? 99999))} />
    </g>
  )
}
