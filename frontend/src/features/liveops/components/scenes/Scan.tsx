import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Packet, Burst, AlertBell, Zone, Counter } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function Scan(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { ports: number[]; open: number[] }
  const attacker = party(scene, 'attacker')!
  const target = party(scene, 'target') ?? party(scene, 'perimeter')!
  const A: Pt = { x: 150, y: 260 }
  const T: Pt = { x: 700, y: 260 }
  const R = 170
  const per = 6000 / f.ports.length
  const sweep = ((t / 1600) % 1) * Math.PI * 2
  return (
    <g>
      <Zone x={36} y={70} w={250} h={380} label="Internet" tone={tone.attacker} />
      <circle cx={T.x} cy={T.y} r={R + 22} fill="none" stroke={tone.muted} strokeOpacity={0.25} strokeDasharray="4 6" />
      {t > 400 && t < 9000 && (
        <path d={`M${T.x},${T.y} L${T.x + Math.cos(sweep) * (R + 22)},${T.y + Math.sin(sweep) * (R + 22)}`} stroke={tone.attacker} strokeOpacity={0.5} strokeWidth={2} />
      )}
      {f.ports.map((port, i) => {
        const ang = (i / f.ports.length) * Math.PI * 2 - Math.PI / 2
        const pos = { x: T.x + Math.cos(ang) * R, y: T.y + Math.sin(ang) * R }
        const start = 600 + i * per
        if (props.blockedAt !== undefined && start > props.blockedAt) return (
          <g key={port}>
            <rect x={pos.x - 23} y={pos.y - 11} width={46} height={22} rx={6} fill={tone.surface} stroke={tone.line} />
            <text x={pos.x} y={pos.y + 4} textAnchor="middle" fill={tone.textFaint} fontSize={11} fontFamily="JetBrains Mono, monospace">{port}</text>
          </g>
        )
        const u = prog(t, start, start + 700)
        const done = t > start + 700
        const open = f.open.includes(port)
        const ink = !done ? tone.muted : open && t > 6200 ? tone.ok : open ? tone.highlight : tone.textFaint
        return (
          <g key={port}>
            {u > 0 && u < 1 && <Packet a={A} b={pos} u={u} tone={tone.attacker} r={3} />}
            <rect x={pos.x - 23} y={pos.y - 11} width={46} height={22} rx={6} fill={tone.surface} stroke={ink} strokeWidth={done && open ? 2 : 1} />
            <text x={pos.x} y={pos.y + 4} textAnchor="middle" fill={done ? (open ? tone.okText : tone.textDim) : tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
              {port}
            </text>
            {done && open && <Burst at={pos} t={t} start={start + 700} tone={tone.ok} size={20} dur={600} />}
          </g>
        )
      })}
      <Counter x={161} y={384} label={tr('liveOps.scPortsProbed')} value={`${Math.min(f.ports.length, Math.max(0, Math.floor((t - 600) / per) + 1))}/${f.ports.length}`} tone={tone.attacker} opacity={prog(t, 500, 900)} />
      <MachineNode party={attacker} at={A} t={t} state={props.blockedAt !== undefined && t >= props.blockedAt ? 'blocked' : 'active'} selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      <MachineNode party={target} at={T} t={t} state={t > 600 ? 'targeted' : 'idle'} selected={props.selectedId === target.id} onSelect={sel(props, target)} />
      <AlertBell at={{ x: T.x + 238, y: T.y - 150 }} t={t} start={9000} />
    </g>
  )
}
