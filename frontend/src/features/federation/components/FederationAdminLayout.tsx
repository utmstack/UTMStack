import { NavLink, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Topbar } from '@/shared/layouts/DashboardLayout/Topbar'

/**
 * Chrome for FS-level account/admin pages (Profile, Team, Email). These are
 * federation-owned settings — reachable without an instance selected — so they
 * live OUTSIDE the FederationGate, with the real Topbar plus a tab bar to move
 * between them. The pages themselves are the same ones the instance app uses.
 */
export function FederationAdminLayout() {
  const { t } = useTranslation()
  const tabs = [
    { to: '/profile', label: t('topbar.profile.profile') },
    { to: '/team', label: t('team.title') },
    { to: '/settings/email', label: t('emailConfig.title') },
  ]
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Topbar />
      <nav className="border-b border-border bg-card">
        <div className="flex w-full gap-1 px-6">
          {tabs.map((tab) => (
            <NavLink
              key={tab.to}
              to={tab.to}
              end
              className={({ isActive }) =>
                cn(
                  '-mb-px border-b-2 px-3 py-3 text-sm transition-colors',
                  isActive
                    ? 'border-primary font-medium text-foreground'
                    : 'border-transparent text-muted-foreground hover:text-foreground',
                )
              }
            >
              {tab.label}
            </NavLink>
          ))}
        </div>
      </nav>
      <Outlet />
    </div>
  )
}
