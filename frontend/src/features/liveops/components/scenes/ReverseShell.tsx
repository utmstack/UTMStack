import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

// A connection opens from the compromised host back to the attacker, and a
// shell prompt appears on the attacker side: they can now type on the host.
export function ReverseShell(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { port: number }
  const attacker = party(scene, 'attacker')!
  const host = party(scene, 'target')!
  const A: Pt = { x: 160, y: 250 }
  const T: Pt = { x: 800, y: 260 }
  const open = prog(t, 1600, 3200)
  const shell = t > 3600
  const typeProg = prog(t, 4400, 7200)
  const cmd = 'whoami; id; cat /etc/passwd'
  return (
    <g>
      <Zone x={36} y={70} w={268} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={T} b={A} tone={shell ? tone.alert : tone.internal} t={t} appear={1600} />
      {open > 0 && open < 1 && <Packet a={T} b={A} u={open} tone={tone.internal} r={4} label={tr('liveOps.scConn', { port: f.port })} />}
      {shell && <Burst at={A} t={t} start={3600} tone={tone.alert} size={70} dur={900} />}
      {shell && (
        <g opacity={prog(t, 3600, 4200)} transform={`translate(${A.x},${A.y - 96})`}>
          <rect x={-105} y={-22} width={210} height={44} rx={7} fill={tone.surfaceAlt} stroke={tone.alert} strokeWidth={1.6} filter={glow} />
          <text y={4} textAnchor="middle" fill={tone.alertText} fontSize={12} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
            {tr('liveOps.scShellOn')}
          </text>
        </g>
      )}
      {shell && (
        <g opacity={prog(t, 4200, 4800)} transform={`translate(${A.x},${A.y + 128})`}>
          <rect x={-120} y={-20} width={240} height={40} rx={7} fill={tone.bg} stroke={tone.line} />
          <text x={-108} y={5} fill={tone.textSoft} fontSize={11.5} fontFamily="JetBrains Mono, monospace">
            {'$ ' + cmd.slice(0, Math.floor(cmd.length * typeProg))}
            {t > 4400 + typeProg * 2800 && <tspan fill={tone.textMuted}>▌</tspan>}
          </text>
        </g>
      )}
      <MachineNode party={attacker} at={A} t={t} state={shell ? 'compromised' : 'active'} selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      <MachineNode party={host} at={T} t={t} state={shell ? 'compromised' : 'targeted'} selected={props.selectedId === host.id} onSelect={sel(props, host)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 118 }} t={t} start={before(8400, (props.blockedAt ?? 99999))} />
    </g>
  )
}
