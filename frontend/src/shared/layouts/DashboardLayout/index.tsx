import { Outlet } from 'react-router-dom'
import { SocAiPanel, SocAiProvider, useSocAiConfigured } from '@/features/soc-ai'
import { AdminSetupGate } from '@/features/onboarding/AdminSetupGate'
import { Sidebar } from './Sidebar'
import { SupportSessionBanner } from './SupportSessionBanner'
import { Topbar } from './Topbar'

export function DashboardLayout() {
  // The chat panel is only useful once a provider is configured in
  // Settings → SOC-AI. The Topbar's Ask button gates on the same thing.
  const socAiReady = useSocAiConfigured()

  return (
    <SocAiProvider>
      <div className="flex h-screen w-screen flex-col overflow-hidden">
        <SupportSessionBanner />
        <Topbar />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          <main className="min-w-0 flex-1 overflow-y-auto bg-muted/30">
            <Outlet />
          </main>
        </div>
        {/* SOC-AI assistant: shown only when a model provider is configured. */}
        {socAiReady && <SocAiPanel />}
        {/* First-run: force the admin to set a real email before using the app. */}
        <AdminSetupGate />
      </div>
    </SocAiProvider>
  )
}
