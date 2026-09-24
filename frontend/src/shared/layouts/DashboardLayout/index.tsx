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
      <div className="flex h-screen w-screen overflow-hidden">
        {/* Everything but the assistant. The assistant is a column beside it, so
            opening the chat pushes the whole screen aside rather than covering it. */}
        <div className="flex min-w-0 flex-1 flex-col">
          <SupportSessionBanner />
          <Topbar />
          {/* `contain: layout` makes this the containing block of the
              `position: fixed` overlays rendered below (drawers, modals): they
              fill this area instead of the viewport, so an open drawer and the
              assistant sit side by side. The topbar stays outside so its
              buttons remain reachable with a drawer open. */}
          <div className="flex min-h-0 flex-1 [contain:layout]">
            <Sidebar />
            <main className="min-w-0 flex-1 overflow-y-auto bg-muted/30">
              <Outlet />
            </main>
          </div>
        </div>
        {/* SOC-AI assistant: shown only when a model provider is configured. */}
        {socAiReady && <SocAiPanel />}
        {/* First-run: force the admin to set a real email before using the app. */}
        <AdminSetupGate />
      </div>
    </SocAiProvider>
  )
}
