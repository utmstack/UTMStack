import { Ban, Cloud, Globe, Laptop, Monitor, Radar, Server, ShieldCheck, Skull, Terminal, User, Boxes, BrickWall, Network } from 'lucide-react'
import type { NodeKind, Party } from '@/features/liveops/model'
import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { backOut, ease, easeOut, prog, type Pt } from './geometry'

const MONO = 'JetBrains Mono, monospace'

function halo(bg: string, width = 4) {
  return { stroke: bg, strokeWidth: width, strokeLinejoin: 'round' as const, paintOrder: 'stroke' as const }
}

function subLines(sub: string): string[] {
  const clip = (x: string, n: number) => (x.length > n ? `${x.slice(0, n - 1)}…` : x)
  if (sub.length <= 44 || !sub.includes(' · ')) return [clip(sub, 50)]
  const [first, ...rest] = sub.split(' · ')
  return [clip(first, 44), clip(rest.join(' · '), 44)]
}

const ICON: Record<NodeKind, typeof Globe> = {
  globe: Globe,
  firewall: BrickWall,
  server: Server,
  dc: Network,
  workstation: Monitor,
  laptop: Laptop,
  web: Server,
  k8s: Boxes,
  cloud: Cloud,
  skull: Skull,
  user: User,
  sensor: Radar,
}

export type NodeState = 'idle' | 'active' | 'targeted' | 'compromised' | 'blocked' | 'flagged'

interface NodeProps {
  party: Party
  at: Pt
  t: number
  state?: NodeState
  /** When the machine fades in (ms). */
  appear?: number
  selected?: boolean
  onSelect?: (p: Party) => void
  r?: number
}

interface NodeProps {
  party: Party
  at: Pt
  t: number
  state?: NodeState
  /** When the machine fades in (ms). */
  appear?: number
  selected?: boolean
  onSelect?: (p: Party) => void
  r?: number
}

export function MachineNode({ party, at, t, state = 'idle', appear = 0, selected, onSelect, r = 36 }: NodeProps) {
  const p = usePalette()
  const glow = useGlow()
  // An outside machine that has been blocked stays red but fades: it is cut off, not safe.
  const cutOff = state === 'blocked' && (party.role === 'attacker' || party.role === 'external')
  const tone = state === 'compromised' || state === 'flagged' ? p.alert : state === 'blocked' && !cutOff ? p.ok : p[party.role]
  const Icon = ICON[party.kind] ?? Server
  const grow = prog(t, appear, appear + 650)
  const o = ease(Math.min(1, grow * 1.6))
  const pulsing = state === 'targeted' || state === 'compromised' || state === 'flagged'
  // Two rings a half-beat apart, slowing as they spread, so the pulse reads as a steady signal.
  const rings = pulsing ? [(t % 1600) / 1600, ((t + 800) % 1600) / 1600] : []
  const clickable = !!onSelect
  const subs = party.sub ? subLines(party.sub) : []
  const userY = r + 36 + subs.length * 15
  return (
    <g
      transform={`translate(${at.x},${at.y}) scale(${0.82 + 0.18 * backOut(grow)})`}
      opacity={o * (cutOff ? 0.6 : 1)}
      style={{ cursor: clickable ? 'pointer' : 'default' }}
      onClick={clickable ? () => onSelect(party) : undefined}
      data-party={party.id}
    >
      {rings.map((u, i) => (
        <circle key={i} r={r + 4 + easeOut(u) * 30} fill="none" stroke={tone} strokeOpacity={0.5 * (1 - u) * (1 - u)} strokeWidth={2} />
      ))}
      {selected && <circle r={r + 9} fill="none" stroke={p.text} strokeOpacity={0.7} strokeDasharray="4 5" strokeWidth={1.5} />}
      <circle r={r} fill={p.surface} stroke={tone} strokeWidth={2.2} filter={glow} />
      <circle r={r - 7} fill={tone} fillOpacity={0.09} />
      <Icon x={-14} y={-14} width={28} height={28} color={tone} strokeWidth={1.7} />
      {party.flag && (
        <text x={r - 6} y={-r + 10} fontSize={18} textAnchor="middle">
          {party.flag}
        </text>
      )}
      {party.agentId && (
        <g transform={`translate(${r - 8},${r - 8})`}>
          <circle r={10} fill={p.surfaceAlt} stroke={p.textMuted} strokeOpacity={0.6} />
          <Terminal x={-6} y={-6} width={12} height={12} color={p.textSoft} strokeWidth={2} />
        </g>
      )}
      {state === 'blocked' && (
        <g transform={`translate(${-r + 8},${-r + 8})`}>
          <circle r={11} fill={cutOff ? p.alertWash : p.okWash} stroke={cutOff ? p.alert : p.ok} />
          {cutOff ? <Ban x={-7} y={-7} width={14} height={14} color={p.alert} strokeWidth={2.2} /> : <ShieldCheck x={-7} y={-7} width={14} height={14} color={p.ok} strokeWidth={2.2} />}
        </g>
      )}
      <text y={r + 20} textAnchor="middle" fill={p.text} fontSize={13} fontWeight={600} {...halo(p.bg)}>
        {party.name.length > 44 ? `${party.name.slice(0, 43)}…` : party.name}
      </text>
      {subs.map((line, i) => (
        <text key={i} y={r + 36 + i * 15} textAnchor="middle" fill={p.textMuted} fontSize={11} fontFamily={MONO} {...halo(p.bg)}>
          {line}
        </text>
      ))}
      {party.user && party.role === 'internal' && (
        <g transform={`translate(0,${userY})`}>
          <rect x={-78} y={-11} width={156} height={20} rx={10} fill={p.surfaceAlt} stroke={p.line} />
          <User x={-72} y={-7} width={13} height={13} color={p.textMuted} />
          <text x={6} y={4} textAnchor="middle" fill={p.textSoft} fontSize={11}>
            {party.user.length > 24 ? `${party.user.slice(0, 23)}…` : party.user}
          </text>
        </g>
      )}
    </g>
  )
}
