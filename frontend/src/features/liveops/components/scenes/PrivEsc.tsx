import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

// A token passes from the user chip to the host, growing from user-sized to
// admin-sized as it is stolen/escalated; the verdict names the technique.
export function PrivEsc(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { method: string; account: string }
  const host = party(scene, 'host')!
  const user = party(scene, 'user')
  const H0: Pt = { x: 720, y: 260 }
  const U: Pt = { x: 200, y: 260 }
  const steal = prog(t, 2600, 5200)
  const done = t > 5200
  const r = 16 + 10 * prog(t, 2600, 5200)
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      {user && (
        <g opacity={prog(t, 0, 600)}>
          <circle cx={U.x} cy={U.y} r={34} fill={tone.surface} stroke={tone.internal} strokeWidth={2.2} />
          <text x={U.x} y={U.y + 56} textAnchor="middle" fill={tone.text} fontSize={13} fontWeight={600}>
            {user.name}
          </text>
          <text x={U.x} y={U.y + 74} textAnchor="middle" fill={tone.textMuted} fontSize={11}>
            {tr('liveOps.scNormalUser')}
          </text>
        </g>
      )}
      {steal > 0 && steal < 1 && (
        <g>
          <circle cx={U.x + (H0.x - U.x) * steal} cy={U.y} r={r} fill={tone.highlight} fillOpacity={0.9} filter={glow} />
        </g>
      )}
      {done && (
        <g opacity={prog(t, 5200, 5800)}>
          <circle cx={H0.x} cy={H0.y} r={r + 6} fill="none" stroke={tone.highlight} strokeWidth={2.5} filter={glow} />
          <Burst at={H0} t={t} start={5200} tone={tone.highlight} size={80} dur={1000} />
        </g>
      )}
      {t > 5600 && <Pill x={(U.x + H0.x) / 2} y={190} text={`${tr('liveOps.scToken')}: ${f.account}`} tone={tone.highlight} mono small opacity={prog(t, 5600, 6200)} />}
      {t > 6800 && <Pill x={H0.x} y={H0.y - 84} text={f.method} tone={tone.highlight} small opacity={prog(t, 6800, 7400)} />}
      <MachineNode party={host} at={H0} t={t} state={done ? 'compromised' : 'targeted'} selected={props.selectedId === host.id} />
      {user && <MachineNode party={user} at={U} t={t} state={done ? 'blocked' : 'active'} selected={props.selectedId === user.id} onSelect={sel(props, user)} r={30} />}
      <AlertBell at={{ x: H0.x + 170, y: H0.y - 118 }} t={t} start={8400} />
    </g>
  )
}
