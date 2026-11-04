import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// Two process boxes side by side: the payload box shrinks and melts into the
// trusted box, which then runs with the injected code — trusted process, red
// inside, host compromised.
export function ProcessInjection(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { injector: string; victim: string }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const I: Pt = { x: 480, y: 220 }
  const V: Pt = { x: 720, y: 300 }
  const inject = prog(t, 3000, 5200)
  const done = t > 5200
  const w = (label: string) => Math.max(150, label.length * 7.6 + 46)
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scInject')} tone={tone.muted} />
      {/* injector */}
      <g opacity={ease(prog(t, 600, 1100))} transform={`translate(${I.x},${I.y}) scale(${1 - 0.85 * inject})`} style={{ transformOrigin: `${I.x}px ${I.y}px` }}>
        <rect x={-w(f.injector) / 2} y={-18} width={w(f.injector)} height={36} rx={9} fill={tone.surface} stroke={done ? tone.textFaint : tone.textMuted} strokeWidth={1.4} />
        <text y={5} textAnchor="middle" fill={tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
          {f.injector}
        </text>
      </g>
      {/* the injected chunk traveling */}
      {inject > 0 && inject < 1 && (
        <circle cx={I.x + (V.x - I.x) * inject} cy={I.y + (V.y - I.y) * inject} r={7} fill={tone.highlight} filter={glow} />
      )}
      {/* victim: turns red when injection completes */}
      <g opacity={ease(prog(t, 600, 1100))} transform={`translate(${V.x},${V.y})`}>
        <rect x={-w(f.victim) / 2} y={-18} width={w(f.victim)} height={36} rx={9} fill={done ? tone.alertWash : tone.surface} stroke={done ? tone.alert : tone.internal} strokeWidth={done ? 2.4 : 1.6} filter={done ? glow : undefined} />
        <text y={5} textAnchor="middle" fill={done ? tone.alertText : tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
          {f.victim}
        </text>
        {done && (
          <text y={34} textAnchor="middle" fill={tone.alertSoft} fontSize={11} fontFamily="JetBrains Mono, monospace" opacity={prog(t, 5400, 6000)}>
            {tr('liveOps.scInside')}
          </text>
        )}
      </g>
      {done && <Burst at={V} t={t} start={5200} tone={tone.alert} size={80} dur={1000} />}
      {done && t > 6400 && <Pill x={560} y={160} text={`${f.injector} → ${f.victim}`} tone={tone.highlight} mono small opacity={prog(t, 6400, 7000)} />}
      <MachineNode party={host} at={H0} t={t} state={done ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8400} />
    </g>
  )
}
