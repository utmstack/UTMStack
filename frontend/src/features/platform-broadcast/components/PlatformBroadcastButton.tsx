import { useState, type ReactNode } from 'react'
import { Radio } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/components/ui/tooltip'
import { useAuth } from '@/features/auth/services/auth.context'
import { useBilling } from '@/features/billing'
import { useSupportTenant } from '@/shared/lib/current-tenant'
import type { BulkResult, BulkSelector } from '../services/broadcast-http.service'
import { BroadcastDialog } from './BroadcastDialog'

export interface PlatformBroadcastButtonProps {
  /** Button label — e.g. "Apply to tenants…". */
  label: string
  /** Modal title — e.g. "Broadcast SMTP update". */
  title: string
  /** Optional context row above the tenant picker. */
  intro?: ReactNode
  /**
   * Set true for endpoints that exclude the platform-plane tenant even when
   * `allTenants=true` (SMTP + branding, per backend/bulk.md).
   */
  excludeDefaultTenant?: boolean
  /** Disable the trigger — e.g. because the form is invalid. */
  disabled?: boolean
  /** Fire the bulk endpoint with the selector the user picked. */
  onBroadcast: (selector: BulkSelector) => Promise<BulkResult>
  /** Optional button variant override; defaults to `outline`. */
  variant?: 'default' | 'outline' | 'secondary' | 'destructive' | 'ghost'
  /** Optional size override; defaults to `sm`. */
  size?: 'default' | 'sm' | 'lg'
}

/**
 * Self-gated broadcast trigger. Renders nothing unless the current session is
 * a platform admin outside a support impersonation — mirrors the sidebar's
 * `isPlatformAdmin && supportTenant === null` rule.
 *
 * Callers wire this next to the normal single-tenant save/delete button and
 * hand it the resource payload they already have in local form state; the
 * modal supplies the tenant selector.
 */
export function PlatformBroadcastButton(props: PlatformBroadcastButtonProps) {
  const { isPlatformAdmin } = useAuth()
  const { license } = useBilling()
  const supportTenant = useSupportTenant()
  const [open, setOpen] = useState(false)

  if (!isPlatformAdmin || supportTenant !== null || license?.mssp !== true) return null

  return (
    <>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant={props.variant ?? 'outline'}
            size="icon"
            disabled={props.disabled}
            onClick={() => setOpen(true)}
            aria-label={props.label}
            className={props.size === 'sm' ? 'h-8 w-8' : undefined}
          >
            <Radio size={14} />
          </Button>
        </TooltipTrigger>
        <TooltipContent>{props.label}</TooltipContent>
      </Tooltip>
      <BroadcastDialog
        open={open}
        title={props.title}
        intro={props.intro}
        excludeDefaultTenant={props.excludeDefaultTenant}
        onClose={() => setOpen(false)}
        onConfirm={props.onBroadcast}
      />
    </>
  )
}
