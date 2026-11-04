import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// Tools on the host are struck through one by one as they are disabled, with a
// "logs cleared" sweep at the end; the host is flagged, not compromised — the
// machine still works, but nothing is watching it.
export function DefenseEvasion(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { tools: { name: string; kind: string }[]; cleared: boolean }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const step = Math.min(2000, 5200 / f.tools.length)
  const last = f.tools.length - 1
  const rows = f.tools.map((_, i) => ({ x: 400, y: 130 + i * (340 / Math.max(1, f.tools.length + (f.cleared ? 1 : 0))) }))
  const clearedAt = (last + 1) * step + 600
  const done = t > (f.cleared ? clearedAt : last * step) + 900
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scDefenses')} tone={tone.muted} />
      {f.tools.map((tool, i) => {
        const appear = i * step
        const o = prog(t, appear, appear + 500)
        if (o <= 0) return null
        const r = rows[i]
        const off = t > appear + 700
        const w = Math.max(160, tool.name.length * 7.6 + 46)
        return (
          <g key={i} opacity={o}>
            <g transform={`translate(${r.x},${r.y - 18})`}>
              <rect width={w} height={36} rx={9} fill={tone.surface} stroke={off ? tone.alert : tone.textMuted} strokeWidth={off ? 2.2 : 1.4} />
              <text x={14} y={23} fill={off ? tone.textFaint : tone.textSoft} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600} textDecoration={off ? 'line-through' : undefined}>
                {tool.name}
              </text>
              {off && <line x1={0} y1={18} x2={w} y2={18} stroke={tone.alert} strokeWidth={2.4} />}
            </g>
            <text x={r.x + w + 14} y={r.y - 2} fill={off ? tone.alertText : tone.textMuted} fontSize={11} fontFamily="JetBrains Mono, monospace" opacity={o}>
              {off ? tr('liveOps.scDisabled') : tool.kind}
            </text>
            {off && i === 0 && <Burst at={{ x: r.x + 20, y: r.y }} t={t} start={appear + 700} tone={tone.alert} size={50} dur={800} />}
          </g>
        )
      })}
      {f.cleared && t > clearedAt && (
        <g opacity={prog(t, clearedAt, clearedAt + 600)}>
          <rect x={400} y={rows[last].y + 18} width={240} height={36} rx={9} fill={tone.alertWash} stroke={tone.alert} strokeWidth={2} />
          <rect x={400} y={rows[last].y + 18} width={240} height={36} rx={9} fill={tone.alertWash} stroke={tone.alert} strokeWidth={2} />
          <text x={414} y={rows[last].y + 41} fill={tone.alertText} fontSize={12.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
            {tr('liveOps.scLogsCleared')}
          </text>
        </g>
      )}
      <MachineNode party={host} at={H0} t={t} state={done ? 'flagged' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8600} />
    </g>
  )
}
