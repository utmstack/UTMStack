import { Outlet } from 'react-router-dom'
import { SocAiPanel, SocAiProvider, useSocAi, useSocAiConfigured } from '@/features/soc-ai'
import { AdminSetupGate } from '@/features/onboarding/AdminSetupGate'
import { cn } from '@/shared/lib/utils'
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
          <Main />
        </div>
        {/* SOC-AI assistant: shown only when a model provider is configured. */}
        {socAiReady && <SocAiPanel />}
        {/* First-run: force the admin to set a real email before using the app. */}
        <AdminSetupGate />
      </div>
    </SocAiProvider>
  )
}

// Reads the panel's own open/expanded state — has to live inside SocAiProvider
// — and reserves exactly the width SocAiPanel occupies (same fixed panel,
// unchanged) so it pushes the page aside instead of covering it. The widths
// below must stay in sync with SocAiPanel's own w-[...] classes.
function Main() {
  const { open, expanded } = useSocAi()
  return (
    <main
      className={cn(
        'min-w-0 flex-1 overflow-y-auto bg-muted/30 transition-[margin-right] duration-200',
        open && (expanded ? 'mr-[min(720px,95vw)]' : 'mr-[min(420px,95vw)]')
      )}
    >
      <Outlet />
    </main>
  )
}
