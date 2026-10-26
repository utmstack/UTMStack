import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone, Counter } from '..'
import { stream, prog, ease, type Pt } from '../geometry'
import { type SceneProps, party, sel, before } from './SceneProps'

export function BruteForce(props: SceneProps) {
  const { scene, t } = props
  const tone = usePalette()
  const f = scene.facts as { attempts: number; accounts: string[]; account: string; success: boolean; cloud: boolean }
  const attacker = party(scene, 'attacker')!
  const perimeter = party(scene, 'perimeter')
  const target = party(scene, 'target') ?? party(scene, 'cloud')!
  const A: Pt = { x: 150, y: 250 }
  const P: Pt = { x: 500, y: 250 }
  const T: Pt = perimeter ? { x: 835, y: 250 } : { x: 780, y: 250 }
  const firstHop = perimeter ? P : T
  const attackEnd = before(f.success ? 7500 : 9600, props.blockedAt)
  const tries = stream(t, 300, attackEnd, 150, 950)
  const tries2 = perimeter ? stream(t, 1250, attackEnd + 950, 150, 600) : []
  const ambient = t > 12000 && props.blockedAt === undefined ? stream(t, 12000, 1e12, 900, 950) : []
  const isBlocked = props.blockedAt !== undefined && t >= props.blockedAt
  const failed = Math.round(f.attempts * ease(prog(t, 300, attackEnd)))
  const account = f.accounts[Math.floor(t / 650) % f.accounts.length]
  const won = f.success && t > 7600 && (props.blockedAt === undefined || props.blockedAt > 7600)
  return (
    <g>
      <Zone x={36} y={70} w={268} h={360} label="Internet" tone={tone.attacker} />
      {f.cloud ? <Zone x={600} y={70} w={364} h={360} label="Microsoft cloud" tone={tone.cloud} /> : <Zone x={640} y={70} w={324} h={360} label="Internal network" tone={tone.internal} />}
      <Wire a={A} b={firstHop} tone={tone.attacker} t={t} />
      {perimeter && <Wire a={P} b={T} tone={tone.perimeter} t={t} />}
      {tries.map((p) => (
        <Packet key={`a${p.i}`} a={A} b={firstHop} u={p.u} tone={tone.attacker} r={3.6} bend={((p.i % 5) - 2) * 14} />
      ))}
      {tries2.map((p) => (
        <Packet key={`b${p.i}`} a={P} b={T} u={p.u} tone={tone.attacker} r={3.2} bend={((p.i % 5) - 2) * 8} />
      ))}
      {ambient.map((p) => (
        <Packet key={`c${p.i}`} a={A} b={firstHop} u={p.u} tone={tone.attacker} r={3} />
      ))}
      {t > 400 && t < attackEnd + 600 && <Pill x={(A.x + firstHop.x) / 2} y={190} text={`user: ${account}`} tone={tone.attacker} mono small />}
      {stream(t, 1250, attackEnd + 1000, 450, 450).map((p) => (
        <Burst key={`x${p.i}`} at={T} t={t} start={1250 + p.i * 450} tone={tone.alert} size={30} dur={450} />
      ))}
      {t > 1300 && t < attackEnd + 800 && <Pill x={T.x} y={T.y - 64} text="✕ wrong password" tone={tone.alert} small opacity={0.5 + 0.5 * Math.abs(Math.sin(t / 220))} />}
      {won && <Packet a={A} b={T} u={prog(t, 7600, 8600)} tone={tone.highlight} r={6} label="✓ signed in" />}
      {won && t > 8600 && <Pill x={T.x} y={T.y - 64} text={`signed in as ${f.account}`} tone={tone.alert} />}
      <Counter x={500} y={455} label="Failed sign-ins" value={failed.toLocaleString()} tone={tone.attacker} opacity={prog(t, 500, 900)} />
      {isBlocked && <Pill x={(A.x + firstHop.x) / 2} y={A.y + 44} text={`🛡 ${attacker.name} blocked`} tone={tone.ok} />}
      <MachineNode party={attacker} at={A} t={t} state={isBlocked ? 'blocked' : 'active'} selected={props.selectedId === attacker.id} onSelect={sel(props, attacker)} />
      {perimeter && <MachineNode party={perimeter} at={P} t={t} state={isBlocked ? 'blocked' : t > 1000 ? 'targeted' : 'idle'} selected={props.selectedId === perimeter.id} onSelect={sel(props, perimeter)} />}
      <MachineNode party={target} at={T} t={t} state={isBlocked && target.kind === 'firewall' ? 'blocked' : won ? 'compromised' : t > 1000 ? 'targeted' : 'idle'} selected={props.selectedId === target.id} onSelect={sel(props, target)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 118 }} t={t} start={10200} labelAbove />
    </g>
  )
}
