import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

// An attacker emails a user; the envelope travels across the zone, the user
// opens it, the attachment lands on the machine, the machine goes red.
export function Phishing(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { subject: string; file: string }
  const attacker = party(scene, 'attacker')!
  const user = party(scene, 'user')!
  const A: Pt = { x: 150, y: 250 }
  const U: Pt = { x: 680, y: 250 }
  const env = stream(t, 800, 3000, 2200, 1400)
  const opened = t > 2400
  const bad = t > 4600
  return (
    <g>
      <Zone x={36} y={70} w={268} h={360} label="Internet" tone={tone.attacker} />
      <Zone x={420} y={70} w={544} h={360} label={tr('liveOps.netInternal')} tone={tone.internal} />
      <Wire a={A} b={U} tone={opened ? tone.highlight : tone.attacker} t={t} />
      {env.map((p) => (
        <Packet key={p.i} a={A} b={U} u={p.u} tone={tone.attacker} r={5} label={p.u < 0.2 ? '✉ ' + f.subject.slice(0, 22) : undefined} bend={10} />
      ))}
      {opened && <Burst at={U} t={t} start={2400} tone={tone.highlight} size={60} dur={800} />}
      {t > 3400 && (
        <g opacity={prog(t, 3400, 4000)} transform={`translate(${U.x},${U.y + 86})`}>
          <rect x={-95} y={-18} width={190} height={36} rx={8} fill={bad ? tone.alertWash : tone.surfaceAlt} stroke={bad ? tone.alert : tone.line} strokeWidth={bad ? 2 : 1.2} filter={bad ? glow : undefined} />
          <text y={5} textAnchor="middle" fill={bad ? tone.alertText : tone.textSoft} fontSize={12} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
            {f.file}
          </text>
        </g>
      )}
      {bad && <Pill x={U.x} y={U.y - 84} text={tr('liveOps.scAttachment')} tone={tone.alert} small />}
      <MachineNode party={attacker} at={A} t={t} state="active" selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      <MachineNode party={user} at={U} t={t} state={bad ? 'compromised' : opened ? 'targeted' : 'idle'} selected={props.selectedId === user.id} onSelect={sel(props, user)} />
      <AlertBell at={{ x: U.x + 70, y: U.y - 118 }} t={t} start={7800} />
    </g>
  )
}
