import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Burst, AlertBell, Zone, Pill } from '..'
import { ease, prog, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A grid of documents: each encrypts in turn (fills red), the ransom note pops,
// then the shadow copies are struck through one by one and the host ends compromised.
export function Ransomware(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { files: number; extensions: string[]; note: boolean }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const per = 3300 / f.files
  const last = f.files - 1
  const done = 600 + last * per + 700
  // Two columns inside the Files zone (350..710 x 70..450), rows spread 140..410.
  const rows = Math.ceil(f.files / 2)
  const cell = (i: number) => ({ x: 370 + (i % 2) * 175, y: rows === 1 ? 275 : 140 + Math.floor(i / 2) * (270 / (rows - 1)) })
  const shadows = [130, 178, 226]
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={360} h={380} label={tr('liveOps.scFiles')} tone={tone.muted} />
      <Zone x={740} y={70} w={224} h={380} label={tr('liveOps.scBackups')} tone={tone.muted} />
      {Array.from({ length: f.files }, (_, i) => {
        const start = 600 + i * per
        const o = ease(prog(t, start, start + 500))
        if (o <= 0) return null
        const r = cell(i)
        const enc = t > start + 700
        const ink = enc ? tone.alert : tone.textMuted
        return (
          <g key={i} opacity={o}>
            <rect x={r.x} y={r.y - 35} width={130} height={70} rx={8} fill={enc ? tone.alertWash : tone.surface} stroke={ink} strokeWidth={enc ? 2 : 1.2} filter={enc && i === last ? glow : undefined} />
            <text x={r.x + 12} y={r.y + 5} fill={enc ? tone.alertText : tone.textSoft} fontSize={11.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
              {enc ? `report_${i + 1}.enc` : `report_${i + 1}.pdf`}
            </text>
            {enc && i === last && <Burst at={{ x: r.x + 65, y: r.y }} t={t} start={start + 700} tone={tone.alert} size={60} dur={800} />}
          </g>
        )
      })}
      {t > done && f.extensions.length > 0 && <Pill x={530} y={428} text={f.extensions.join('  ')} tone={tone.alert} mono small opacity={prog(t, done, done + 500)} />}
      {shadows.map((y, i) => {
        const o = prog(t, 200 + i * 250, 700 + i * 250)
        if (o <= 0) return null
        const dead = t > 6800 + i * 400
        return (
          <g key={i} opacity={o}>
            <rect x={760} y={y} width={184} height={36} rx={8} fill={dead ? tone.surface : tone.surfaceAlt} stroke={dead ? tone.alert : tone.line} strokeWidth={dead ? 1.6 : 1.2} />
            <text x={774} y={y + 23} fill={dead ? tone.textFaint : tone.textMuted} fontSize={11.5} fontFamily="JetBrains Mono, monospace" textDecoration={dead ? 'line-through' : undefined}>
              {`snapshot-${3 - i} (hourly)`}
            </text>
            {dead && <line x1={760} y1={y + 18} x2={944} y2={y + 18} stroke={tone.alert} strokeWidth={2.4} opacity={prog(t, 6800 + i * 400, 7100 + i * 400)} />}
          </g>
        )
      })}
      {t > 8400 && <Pill x={852} y={330} text={tr('liveOps.scShadows')} tone={tone.ok} small opacity={prog(t, 8400, 8900)} />}
      {f.note && t > 5800 && (
        <g opacity={prog(t, 5800, 6400)} transform="translate(852,395)">
          <rect x={-70} y={-24} width={140} height={48} rx={8} fill={tone.alertWash} stroke={tone.alert} strokeWidth={2} filter={glow} />
          <text y={6} textAnchor="middle" fill={tone.alertText} fontSize={12} fontWeight={700}>
            {tr('liveOps.scRansomNote')}
          </text>
        </g>
      )}
      <MachineNode party={host} at={H0} t={t} state={t > done ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8400} />
    </g>
  )
}
