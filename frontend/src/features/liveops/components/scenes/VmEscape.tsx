import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A guest VM inside a hypervisor zone: an escape arc breaks out of the guest
// box into the host, crossing the boundary — the host is then flagged.
export function VmEscape(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { guest: string }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 820, y: 260 }
  // The guest box inside the left zone.
  const G: Pt = { x: 300, y: 250 }
  const guestOut = prog(t, 3600, 5600)
  const out = t > 5600
  return (
    <g>
      <Zone x={40} y={70} w={520} h={380} label={tr('liveOps.scHypervisor')} tone={tone.internal} />
      <Zone x={640} y={70} w={324} h={360} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      {/* the guest VM box */}
      <g opacity={ease(prog(t, 300, 900))} transform={`translate(${G.x - 150},${G.y - 80})`}>
        <rect width={300} height={160} rx={12} fill={tone.surface} stroke={out ? tone.alert : tone.perimeter} strokeWidth={out ? 2.4 : 1.6} filter={out ? glow : undefined} />
        <text x={18} y={30} fill={out ? tone.alertText : tone.textSoft} fontSize={13} fontWeight={600}>
          {f.guest}
        </text>
        <text x={18} y={52} fill={tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
          {tr('liveOps.scGuestVm')}
        </text>
        <text x={18} y={90} fill={out ? tone.alertSoft : tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
          {out ? tr('liveOps.scEscaped') : tr('liveOps.scConfined')}
        </text>
      </g>
      {/* the escape arc breaking the guest boundary toward the host */}
      {guestOut > 0 && (
        <path
          d={`M${G.x + 150},${G.y} Q${(G.x + 150 + H0.x) / 2},${G.y - 120} ${H0.x - 40},${H0.y}`}
          fill="none"
          stroke={tone.highlight}
          strokeWidth={2.4}
          strokeDasharray="900"
          strokeDashoffset={900 * (1 - guestOut)}
        />
      )}
      {guestOut > 0 && guestOut < 1 && (
        <circle cx={G.x + 150 + (H0.x - G.x - 150) * guestOut} cy={G.y - 110 * Math.sin(guestOut * Math.PI) + 0} r={8} fill={tone.highlight} filter={glow} />
      )}
      {out && <Burst at={{ x: H0.x - 40, y: H0.y }} t={t} start={5600} tone={tone.alert} size={80} dur={1000} />}
      {out && <Pill x={560} y={150} text={tr('liveOps.scBoundary')} tone={tone.highlight} opacity={prog(t, 5600, 6200)} />}
      <MachineNode party={host} at={H0} t={t} state={out ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8200} />
    </g>
  )
}
