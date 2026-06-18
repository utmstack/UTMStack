import { Topbar } from '@/shared/layouts/DashboardLayout/Topbar'
import { ProfilePage } from '@/features/profile/pages/ProfilePage'

/**
 * Federation account page: the exact same ProfilePage as the instance app, just
 * rendered standalone (it normally lives inside DashboardLayout) so it's reachable
 * without an instance selected. The ProfilePage hides instance-only sections
 * (2FA, sessions, API keys) in federation mode and routes avatar/profile/password
 * actions to the FS via useAuth().
 */
export function FederationProfilePage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Topbar />
      <ProfilePage />
    </div>
  )
}
