import { useCallback, useState, type ReactNode } from 'react'
import { useAuth } from '@/features/auth/services/auth.context'
import { useSupportTenant } from '@/shared/lib/current-tenant'
import type { BulkResult, BulkSelector } from '../services/broadcast-http.service'

export interface BroadcastActionConfig {
  title: string
  intro?: ReactNode
  excludeDefaultTenant?: boolean
  run: (selector: BulkSelector) => Promise<BulkResult>
}

export interface BroadcastActionHandle {
  /** True only when this session may see & fire platform-broadcast actions. */
  canBroadcast: boolean
  /** Modal is currently open. */
  open: boolean
  /** Trigger the modal with a specific action config. */
  trigger: (config: BroadcastActionConfig) => void
  /** Close the modal. */
  close: () => void
  /** The config currently mounted in the modal (null when closed). */
  config: BroadcastActionConfig | null
}

/**
 * Gates every platform-broadcast entry point on the same two flags the sidebar
 * uses and holds the trigger state. Callers own the `<BroadcastDialog>` render;
 * this hook only sets `open` and hands back the action config.
 */
export function useBroadcastAction(): BroadcastActionHandle {
  const { isPlatformAdmin } = useAuth()
  const supportTenant = useSupportTenant()
  const canBroadcast = isPlatformAdmin && supportTenant === null

  const [config, setConfig] = useState<BroadcastActionConfig | null>(null)

  const trigger = useCallback(
    (next: BroadcastActionConfig) => {
      if (!canBroadcast) return
      setConfig(next)
    },
    [canBroadcast],
  )

  const close = useCallback(() => setConfig(null), [])

  return {
    canBroadcast,
    open: config !== null,
    trigger,
    close,
    config,
  }
}
