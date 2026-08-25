import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Building2, ChevronDown, Plus } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useAuth } from '@/features/auth'
import { setSupportTenant, useSupportTenant } from '@/shared/lib/current-tenant'
import {
  canReadTenant,
  tenantsHttpService,
} from '../services/tenants-http.service'
import type { Tenant } from '../types/tenant.types'
import { CreateTenantDialog } from './CreateTenantDialog'

/**
 * Topbar switcher for the tenant the current session is reading against.
 *
 * Rendered only for admins. The tenants list is exempt from the support-tenant
 * header, so it works even mid-session; on the endpoint responding 403 (a role
 * that says admin but not to this endpoint) the switcher hides itself.
 *
 * Entering a tenant is a hard navigation, same reason as the tenant cards: the
 * react-query caches, branding and notification feed all belong to whoever we
 * were before, and a soft navigation would leave them on screen next to the
 * other tenant's data.
 */
export function TenantSwitcher() {
  const { t } = useTranslation()
  const { isAdmin, tenantId: ownId } = useAuth()
  const support = useSupportTenant()
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [failed, setFailed] = useState(false)
  const [open, setOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const load = useCallback(async () => {
    if (!isAdmin) return
    try {
      const list = await tenantsHttpService.list({ size: 200 })
      setTenants(
        list.filter(
          (x) => x.id !== ownId && canReadTenant(x) && x.status !== 'TERMINATED'
        )
      )
      setFailed(false)
    } catch {
      setFailed(true)
    }
  }, [ownId, isAdmin])

  useEffect(() => {
    void load()
  }, [load])

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  if (!isAdmin || failed) return null

  // Reload in place so the target tenant's data replaces ours, but bounce off
  // /tenants first: entering a tenant strips access to that page, so staying
  // there would just 403 on the next load.
  const reloadAfterSwitch = () => {
    if (window.location.pathname.startsWith('/tenants')) {
      window.location.assign('/home')
    } else {
      window.location.reload()
    }
  }

  const enterSelf = () => {
    setOpen(false)
    setSupportTenant(null)
    reloadAfterSwitch()
  }

  const enter = (tenant: Tenant) => {
    setOpen(false)
    setSupportTenant({
      id: tenant.id,
      name: tenant.name,
      access: tenant.supportAccess === 'FULL' ? 'FULL' : 'READ',
      domain: tenant.domain,
    })
    reloadAfterSwitch()
  }

  const selfLabel = t('tenants.switcher.self', { defaultValue: 'Default tenant' })
  const current = support?.name ?? selfLabel

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        aria-label={t('tenants.switcher.aria', { defaultValue: 'Switch tenant' })}
        className={cn(
          'flex h-9 items-center gap-1.5 rounded-md border border-border bg-muted/40 px-2.5 text-[12px] transition-colors',
          open
            ? 'bg-muted text-foreground'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
        )}
      >
        <Building2 size={13} strokeWidth={1.75} />
        <span className="max-w-[9rem] truncate">{current}</span>
        <ChevronDown
          size={12}
          className={cn('transition-transform duration-150', open ? 'rotate-180' : 'rotate-0')}
        />
      </button>
      {open && (
        <div className="absolute right-0 top-full z-50 mt-1 w-64 overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-lg">
          <div className="max-h-72 overflow-y-auto py-1">
            <button
              onClick={enterSelf}
              className={cn(
                'block w-full truncate px-3 py-1.5 text-left text-sm hover:bg-muted',
                !support && 'font-semibold text-primary'
              )}
            >
              {selfLabel}
            </button>
            {tenants.map((tn) => (
              <button
                key={tn.id}
                onClick={() => enter(tn)}
                title={tn.name}
                className={cn(
                  'block w-full truncate px-3 py-1.5 text-left text-sm hover:bg-muted',
                  support?.id === tn.id && 'font-semibold text-primary'
                )}
              >
                {tn.name}
              </button>
            ))}
          </div>
          <button
            onClick={() => {
              setOpen(false)
              setCreating(true)
            }}
            className="flex w-full items-center gap-2 border-t border-border px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <Plus size={13} strokeWidth={1.75} />
            {t('tenants.switcher.create', { defaultValue: 'Create tenant' })}
          </button>
        </div>
      )}
      {creating && (
        <CreateTenantDialog
          onClose={() => setCreating(false)}
          onCreated={() => {
            setCreating(false)
            void load()
          }}
        />
      )}
    </div>
  )
}
