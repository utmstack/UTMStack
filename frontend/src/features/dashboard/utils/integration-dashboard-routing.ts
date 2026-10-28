import type { Integration } from '@/features/integrations/types'

export function getIntegrationDashboardSearch(
  integration: Pick<Integration, 'moduleName' | 'dataType' | 'name'>,
): string {
  const identity = `${integration.moduleName ?? ''} ${integration.dataType ?? ''} ${integration.name}`.toLowerCase()

  if (identity.includes('windows') || identity.includes('wineventlog')) return 'Windows'
  if (identity.includes('linux')) return 'Linux'
  if (identity.includes('macos') || identity.includes('mac os')) return 'macOS'

  return integration.name
}
