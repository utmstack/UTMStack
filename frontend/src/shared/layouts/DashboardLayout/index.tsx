import { Outlet } from 'react-router-dom'
import { DatasourceLimitBanner } from '@/features/datasources/components/DatasourceLimitBanner'
import { SocAiFloating, SocAiPanel, SocAiProvider } from '@/features/soc-ai'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'

export function DashboardLayout() {
  return (
    <SocAiProvider>
      <div className="flex h-screen w-screen flex-col overflow-hidden">
        <Topbar />
        <DatasourceLimitBanner />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          {/* pb-16 keeps page content clear of the fixed SOC-AI floating pill at
              the bottom-center, which would otherwise overlap end-of-page content. */}
          <main className="min-w-0 flex-1 overflow-y-auto bg-muted/30 pb-16">
            <Outlet />
          </main>
        </div>
        {/* SOC-AI assistant: a floating composer + a slide-in chat panel (mock). */}
        <SocAiFloating />
        <SocAiPanel />
      </div>
    </SocAiProvider>
  )
}
