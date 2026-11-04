import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, ease, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// A remote host serves a tool; the target downloads it in chunks, the file
// lands in a writable folder, then it is executed — the machine turns red.
export function ToolDownload(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { tool: string; folder: string }
  const src = party(scene, 'attacker')!
  const tgt = party(scene, 'target')!
  const S: Pt = { x: 170, y: 250 }
  const T: Pt = { x: 810, y: 260 }
  const start = 1200
  const stop = before(4600, props.blockedAt)
  const chunks = stream(t, start, stop, 600, 1300)
  const landed = t > stop + 300
  const exec = t > stop + 1800
  return (
    <g>
      <Zone x={36} y={70} w={268} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={S} b={T} tone={exec ? tone.alert : tone.attacker} t={t} />
      {chunks.map((p) => (
        <Packet key={p.i} a={S} b={T} u={p.u} tone={tone.attacker} r={4} label={p.u < 0.18 ? tr('liveOps.scDownload') : undefined} />
      ))}
      {landed && (
        <g opacity={ease(prog(t, stop + 300, stop + 900))} transform={`translate(${T.x},${T.y + 86})`}>
          <rect x={-85} y={-18} width={170} height={36} rx={8} fill={exec ? tone.alertWash : tone.surfaceAlt} stroke={exec ? tone.alert : tone.line} strokeWidth={exec ? 2 : 1.2} filter={exec ? glow : undefined} />
          <text y={5} textAnchor="middle" fill={exec ? tone.alertText : tone.textSoft} fontSize={11.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
            {f.folder}/{f.tool}
          </text>
        </g>
      )}
      {exec && <Burst at={T} t={t} start={stop + 1800} tone={tone.alert} size={70} dur={900} />}
      {exec && <Pill x={T.x} y={T.y - 84} text={`${f.tool} ${tr('liveOps.scExecuted')}`} tone={tone.alert} small opacity={prog(t, stop + 1800, stop + 2400)} />}
      <MachineNode party={src} at={S} t={t} state="active" selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={tgt} at={T} t={t} state={exec ? 'compromised' : landed ? 'targeted' : 'idle'} selected={props.selectedId === tgt.id} onSelect={sel(props, tgt)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 118 }} t={t} start={8400} />
    </g>
  )
}
