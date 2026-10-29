import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell, Zone } from '..'
import { stream, prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A credential store on the host (LSASS/SAM/keychain/shadow) is read: a packet
// of "secrets" leaves the store and is stolen; the store is emptied, host flagged.
export function CredDump(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { store: string; count: number }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const S: Pt = { x: 620, y: 230 }
  const out = stream(t, 2400, 5200, 900, 1500)
  const emptied = t > 5200
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scCred')} tone={tone.muted} />
      {/* the credential store box */}
      <g opacity={ease(prog(t, 400, 900))} transform={`translate(${S.x},${S.y})`}>
        <rect x={-90} y={-30} width={180} height={60} rx={10} fill={emptied ? tone.alertWash : tone.surface} stroke={emptied ? tone.alert : tone.highlight} strokeWidth={emptied ? 2.2 : 1.6} filter={emptied ? glow : undefined} />
        <text y={-6} textAnchor="middle" fill={emptied ? tone.alertText : tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
          {f.store}
        </text>
        <text y={14} textAnchor="middle" fill={emptied ? tone.textFaint : tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
          {emptied ? '—' : tr('liveOps.scSecrets', { n: f.count })}
        </text>
      </g>
      {out.map((p) => (
        <Packet key={p.i} a={S} b={H0} u={p.u} tone={tone.highlight} r={4.5} label={p.u < 0.2 ? tr('liveOps.scSecret') : undefined} />
      ))}
      {emptied && <Burst at={S} t={t} start={5200} tone={tone.alert} size={70} dur={900} />}
      {emptied && <Pill x={S.x} y={S.y + 76} text={tr('liveOps.scStolen')} tone={tone.alert} small opacity={prog(t, 5400, 6000)} />}
      <Wire a={S} b={H0} tone={tone.highlight} t={t} appear={2400} />
      <MachineNode party={host} at={H0} t={t} state={emptied ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8400} />
    </g>
  )
}
