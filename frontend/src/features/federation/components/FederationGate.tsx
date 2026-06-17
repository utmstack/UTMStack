import { Navigate } from 'react-router-dom'
import { useCurrentInstanceId } from '@/shared/lib/current-instance'

/**
 * Renders the app only once an instance is selected; otherwise sends the user to
 * the instance picker. Used in federation mode around the DashboardLayout.
 */
export function FederationGate({ children }: { children: React.ReactNode }) {
  const instanceId = useCurrentInstanceId()
  if (instanceId == null) {
    return <Navigate to="/instances" replace />
  }
  return <>{children}</>
}
