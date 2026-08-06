import { Outlet, useLocation } from 'react-router-dom'
import { cn } from '@/shared/lib/utils'
import { SocAiFloating, SocAiPanel, SocAiProvider, useSocAiConfigured } from '@/features/soc-ai'
import { AdminSetupGate } from '@/features/onboarding/AdminSetupGate'
import { Sidebar } from './Sidebar'
import { SupportSessionBanner } from './SupportSessionBanner'
import { Topbar } from './Topbar'

export function DashboardLayout() {
  // The SOC-AI assistant (floating composer + chat panel) is only useful once a
  // provider is configured in Settings → SOC-AI. Until then we hide it entirely
  // and drop the bottom padding that reserves space for the floating pill.
  const socAiReady = useSocAiConfigured()
  // The Overview page already has its own SOC-AI composer up top (same shared
  // session) — the floating pill down here would just be a redundant second one.
  const { pathname } = useLocation()
  const hideFloatingPill = pathname === '/home'

  return (
    <SocAiProvider>
      <div className="flex h-screen w-screen flex-col overflow-hidden">
        <SupportSessionBanner />
        <Topbar />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          {/* no padding on bottom section son socai pill overlaps page end as desired */}
          <main className={cn('min-w-0 flex-1 overflow-y-auto bg-muted/30')}>
            <Outlet />
          </main>
        </div>
        {/* SOC-AI assistant: shown only when a model provider is configured. */}
        {socAiReady && (
          <>
            {!hideFloatingPill && <SocAiFloating />}
            <SocAiPanel />
          </>
        )}
        {/* First-run: force the admin to set a real email before using the app. */}
        <AdminSetupGate />
      </div>
    </SocAiProvider>
  )
}
