import { Outlet } from 'react-router-dom'
import { SocAiFloating, SocAiPanel, SocAiProvider } from '@/features/soc-ai'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'

export function DashboardLayout() {
  return (
    <SocAiProvider>
      <div className="flex h-screen w-screen flex-col overflow-hidden">
        <Topbar />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          <main className="min-w-0 flex-1 overflow-y-auto bg-muted/30">
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
