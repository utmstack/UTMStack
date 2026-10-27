import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone, Counter } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// Data is staged on the host, then flows out to a cloud/external store in
// steady chunks; the counter shows bytes leaving, the verdict pill names the tool.
export function Exfil(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { bytes: number; tool: string; dest: string }
  const src = party(scene, 'source')!
  const dst = party(scene, 'destination')!
  const S: Pt = { x: 820, y: 270 }
  const D: Pt = { x: 180, y: 250 }
  const start = 2600
  const stop = before(7600, props.blockedAt)
  const chunks = stream(t, start, stop, 800, 1500)
  const sent = Math.round(f.bytes * prog(t, start, stop))
  return (
    <g>
      <Zone x={36} y={70} w={300} h={360} label={tr('liveOps.scCloud')} tone={tone.cloud} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={D} tone={tone.internal} t={t} />
      {chunks.map((p) => (
        <Packet key={p.i} a={S} b={D} u={p.u} tone={tone.alert} r={4.5} label={p.u < 0.18 ? tr('liveOps.scChunk') : undefined} />
      ))}
      {t > start && t < stop + 400 && <Pill x={500} y={190} text={`${f.tool} → ${f.dest}`} tone={tone.internal} small />}
      <Burst at={D} t={t} start={stop - 200} tone={tone.alert} size={70} dur={1000} />
      <Counter x={500} y={455} label={tr('liveOps.scExfilBytes')} value={sent > 1048576 ? `${(sent / 1048576).toFixed(1)} MB` : `${sent} KB`} tone={tone.alert} opacity={prog(t, 2800, 3400)} />
      <MachineNode party={src} at={S} t={t} state={t > start ? 'compromised' : 'active'} selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={dst} at={D} t={t} state={t > stop ? 'flagged' : 'idle'} selected={props.selectedId === dst.id} onSelect={sel(props, dst)} />
      <AlertBell at={{ x: D.x + 60, y: D.y - 118 }} t={t} start={8800} />
    </g>
  )
}
