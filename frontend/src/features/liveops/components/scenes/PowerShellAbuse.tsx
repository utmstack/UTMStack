import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, ease, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A terminal on the host: a base64 blob streams in, is decoded live into the
// command, and executes — the prompt turns red when it runs.
export function PowerShellAbuse(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { encoded: string; cmd: string; amsi: boolean }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const TL: Pt = { x: 560, y: 230 }
  const blob = prog(t, 1200, 3800)
  const decode = prog(t, 4200, 5200)
  const ran = t > 5600
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scTerminal')} tone={tone.muted} />
      {/* terminal window */}
      <g opacity={ease(prog(t, 300, 900))} transform={`translate(${TL.x - 120},${TL.y - 80})`}>
        <rect width={440} height={180} rx={10} fill={tone.surface} stroke={ran ? tone.alert : tone.line} strokeWidth={ran ? 2.2 : 1.4} filter={ran ? glow : undefined} />
        <rect width={440} height={26} rx={10} fill={tone.surfaceAlt} />
        <circle cx={16} cy={13} r={4} fill={tone.textFaint} />
        <text x={30} y={17} fill={tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
          {host.os === 'windows' ? 'powershell' : 'shell'}
        </text>
        {/* the encoded blob filling */}
        <text x={16} y={56} fill={tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace">
          {`> `}
        </text>
        <text x={30} y={56} fill={tone.highlight} fontSize={11} fontFamily="JetBrains Mono, monospace" opacity={blob}>
          {f.encoded.slice(0, Math.min(Math.floor(f.encoded.length * blob),60))}
        </text>
        {/* decoded command appearing char by char */}
        <text x={16} y={96} fill={ran ? tone.alertText : tone.textSoft} fontSize={12} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
          {`# ${f.cmd.slice(0, Math.floor(f.cmd.length * decode))}`}
        </text>
        {f.amsi && t > 6400 && (
          <text x={16} y={126} fill={tone.warn} fontSize={11} fontFamily="JetBrains Mono, monospace" opacity={prog(t, 6400, 7000)}>
            {tr('liveOps.scAmsiOff')}
          </text>
        )}
        {ran && <text x={16} y={160} fill={tone.textFaint} fontSize={11} fontFamily="JetBrains Mono, monospace" opacity={prog(t, 5600, 6200)}>
          {tr('liveOps.scRunning')}
        </text>}
      </g>
      {ran && <Burst at={{ x: TL.x, y: TL.y + 30 }} t={t} start={5600} tone={tone.alert} size={70} dur={900} />}
      {ran && <Pill x={TL.x + 10} y={TL.y + 110} text={tr('liveOps.scPayload')} tone={tone.alert} small opacity={prog(t, 5800, 6400)} />}
      <MachineNode party={host} at={H0} t={t} state={ran ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8400} />
    </g>
  )
}
