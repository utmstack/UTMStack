import { Outlet } from 'react-router-dom'
import { cn } from '@/shared/lib/utils'
import { DatasourceLimitBanner } from '@/features/datasources/components/DatasourceLimitBanner'
import { SocAiFloating, SocAiPanel, SocAiProvider, useSocAiConfigured } from '@/features/soc-ai'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'

export function DashboardLayout() {
  // The SOC-AI assistant (floating composer + chat panel) is only useful once a
  // provider is configured in Settings → SOC-AI. Until then we hide it entirely
  // and drop the bottom padding that reserves space for the floating pill.
  const socAiReady = useSocAiConfigured()

  return (
    <SocAiProvider>
      <div className="flex h-screen w-screen flex-col overflow-hidden">
        <Topbar />
        <DatasourceLimitBanner />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          {/* pb-16 keeps page content clear of the fixed SOC-AI floating pill at
              the bottom-center, which would otherwise overlap end-of-page content. */}
          <main className={cn('min-w-0 flex-1 overflow-y-auto bg-muted/30', socAiReady && 'pb-16')}>
            <Outlet />
          </main>
        </div>
        {/* SOC-AI assistant: shown only when a model provider is configured. */}
        {socAiReady && (
          <>
            <SocAiFloating />
            <SocAiPanel />
          </>
        )}
      </div>
    </SocAiProvider>
  )
}
