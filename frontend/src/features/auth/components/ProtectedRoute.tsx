import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../services/auth.context'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return <div className="flex h-screen items-center justify-center">Loading…</div>
  }

  if (!isAuthenticated) {
    return <Navigate to="/auth/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}
