import type { ReactElement } from 'react'
import type { Party, Scene } from '@/features/liveops/model'

// What every scene receives: the built scene, the current time and how the stage handles
// machine selection. Scenes share the small helpers below.
export interface SceneProps {
  scene: Scene
  t: number
  /** i18n function: scene labels are shown in the visitor's language. */
  tr: (k: string, o?: Record<string, unknown>) => string
  selectedId?: string
  onSelect: (p: Party) => void
  /** Scene time at which the visitor blocked the attacker (traffic stops at the firewall from then on). */
  blockedAt?: number
}

export const before = (stop: number, blockedAt?: number) => (blockedAt === undefined ? stop : Math.min(stop, blockedAt))

export const party = (s: Scene, id: string) => s.parties.find((p) => p.id === id)
export const sel = (s: SceneProps, p?: Party) => (p && (p.agentId || p.role === 'attacker' || p.role === 'external' || p.role === 'perimeter') ? s.onSelect : undefined)

export type { ReactElement }
