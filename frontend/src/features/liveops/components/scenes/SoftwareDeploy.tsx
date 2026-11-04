import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// An admin/attacker host pushes a software package across the network to a
// target host; the target installs it and ends compromised.
export function SoftwareDeploy(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { tool: string; target: string }
  const src = party(scene, 'attacker')!
  const tgt = party(scene, 'target')!
  const S: Pt = { x: 180, y: 260 }
  const T: Pt = { x: 800, y: 260 }
  const start = 2200
  const stop = before(5200, props.blockedAt)
  const pkts = stream(t, start, stop, 900, 1400)
  const installed = t > stop + 400
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: src?.name ?? 'unknow' })} tone={tone.attacker} />
      <Zone x={640} y={70} w={324} h={380} label={tr('liveOps.scHost', { name: tgt?.name ?? 'unknow' })} tone={tone.internal} />
      <Wire a={S} b={T} tone={installed ? tone.alert : tone.attacker} t={t} />
      {pkts.map((p) => (
        <Packet key={p.i} a={S} b={T} u={p.u} tone={tone.attacker} r={5} label={p.u < 0.18 ? f.tool : undefined} />
      ))}
      {t > start && t < stop + 400 && <Pill x={500} y={190} text={`${tr('liveOps.scDeploy')}: ${f.tool}`} tone={tone.attacker} small />}
      {installed && <Burst at={T} t={t} start={stop + 400} tone={tone.alert} size={70} dur={1000} />}
      {installed && t > stop + 1200 && (
        <g opacity={prog(t, stop + 1200, stop + 1800)}>
          <rect x={T.x - 70} y={T.y + 64} width={140} height={36} rx={8} fill={tone.alertWash} stroke={tone.alert} strokeWidth={1.8} filter={glow} />
          <text x={T.x} y={T.y + 87} textAnchor="middle" fill={tone.alertText} fontSize={11.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
            {f.target}
          </text>
        </g>
      )}
      <MachineNode party={src} at={S} t={t} state="active" selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={tgt} at={T} t={t} state={installed ? 'compromised' : t > start ? 'targeted' : 'idle'} selected={props.selectedId === tgt.id} onSelect={sel(props, tgt)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 118 }} t={t} start={8400} />
    </g>
  )
}
