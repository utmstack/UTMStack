import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute, useAuth } from '@/features/auth'
import { TenantsPage } from '@/features/tenants/pages/TenantsPage'
import { LoginPage } from '@/features/auth/pages/LoginPage'
import { IS_FEDERATION } from '@/shared/config/mode'
import { FederationGate, InstancePickerPage } from '@/features/federation'
import { FederationAdminLayout } from '@/features/federation/components/FederationAdminLayout'
import { FederationTeamPage } from '@/features/federation/pages/FederationTeamPage'
import { HomePage } from '@/features/home/pages/HomePage'
import {
  DashboardPage,
  NewDashboardPage,
  NewVisualizationPage,
  EditVisualizationPage,
} from '@/features/dashboard'
import { ProfilePage } from '@/features/profile/pages/ProfilePage'
import { AuditPage } from '@/features/audit/pages/AuditPage'
import { AlertsPage } from '@/features/alerts/pages/AlertsPage'
import { TaggingRulesPage } from '@/features/alerts/pages/TaggingRulesPage'
import { IncidentsPage } from '@/features/incidents/pages/IncidentsPage'
import { AdversariesPage } from '@/features/adversaries/pages/AdversariesPage'
import { FlowsPage } from '@/features/soar/pages/FlowsPage'
import { ExecutionHistoryPage } from '@/features/soar/pages/ExecutionHistoryPage'
import { InteractiveConsolePage } from '@/features/soar/pages/InteractiveConsolePage'
import { LogExplorerPage } from '@/features/log-explorer/pages/LogExplorerPage'
import { UserAuditorPage } from '@/features/user-auditor/pages/UserAuditorPage'
import { ThreatIntelPage } from '@/features/threat-intel/pages/ThreatIntelPage'
import { CompliancePage } from '@/features/compliance/pages/CompliancePage'
import { FrameworkReportPage } from '@/features/compliance/pages/FrameworkReportPage'
import { DataSourcesPage } from '@/features/datasources/pages/DataSourcesPage'
import { IntegrationsPage } from '@/features/integrations/pages/IntegrationsPage'
import { AlertingRulesPage } from '@/features/alerting-rules/pages/AlertingRulesPage'
import { ParsingFiltersPage } from '@/features/parsing-filters/pages/ParsingFiltersPage'
import { DataProcessingPage } from '@/features/data-processing/pages/DataProcessingPage'
import { TeamPage } from '@/features/team/pages/TeamPage'
import { ConnectionKeyPage } from '@/features/settings/pages/ConnectionKeyPage'
import { DataRetentionPage } from '@/features/settings/pages/DataRetentionPage'
import { BrandingPage } from '@/features/branding'
import { IdentityProvidersPage } from '@/features/settings/pages/IdentityProvidersPage'
import { EmailConfigurationPage } from '@/features/settings/pages/EmailConfigurationPage'
import { SocAiSettingsPage } from '@/features/settings/pages/SocAiSettingsPage'
import { DateFormatPage } from '@/features/settings/pages/DateFormatPage'
import { LanguagePage } from '@/features/settings/pages/LanguagePage'
import { AboutPage } from '@/features/settings/pages/AboutPage'
import { SupportAccessPage } from '@/features/settings/pages/SupportAccessPage'
import { LicensePage, useBilling } from '@/features/billing'
import { NotificationsPage } from '@/features/notifications'
import { ApiKeysPage } from '@/features/api-keys'
import { DashboardLayout } from '@/shared/layouts/DashboardLayout'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/auth/login" element={<LoginPage />} />

      {/* Federation: account pages that don't need an instance (authed only). */}
      {IS_FEDERATION && (
        <Route
          path="/instances"
          element={
            <ProtectedRoute>
              <InstancePickerPage />
            </ProtectedRoute>
          }
        />
      )}
      {/* Federation account/admin pages: FS-owned, reachable without an instance,
          so they sit outside the FederationGate under their own chrome. */}
      {IS_FEDERATION && (
        <Route
          element={
            <ProtectedRoute>
              <FederationAdminLayout />
            </ProtectedRoute>
          }
        >
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/team" element={<FederationTeamPage />} />
          <Route path="/settings/email" element={<EmailConfigurationPage />} />
        </Route>
      )}

      <Route
        path="/"
        element={
          <ProtectedRoute>
            {IS_FEDERATION ? (
              <FederationGate>
                <DashboardLayout />
              </FederationGate>
            ) : (
              <DashboardLayout />
            )}
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/home" replace />} />
        <Route path="home" element={<HomePage />} />
        {/* Tenants belong to whoever runs the instance, not to a customer. The
            route is gated as well as hidden: a URL typed by hand must not reach
            it either. */}
        <Route
          path="tenants"
          element={
            <PlatformOnly requiresMSSP>
              <TenantsPage />
            </PlatformOnly>
          }
        />
        {/* Dashboards */}
        <Route path="dashboards" element={<Navigate to="/dashboards/list" replace />} />
        <Route path="dashboards/list" element={<DashboardPage />} />
        <Route path="dashboards/new" element={<NewDashboardPage />} />
        <Route path="dashboards/:dashboardId/visualizations/new" element={<NewVisualizationPage />} />
        <Route path="dashboards/:dashboardId/visualizations/:id" element={<EditVisualizationPage />} />

        {/* Threat Management */}
        <Route path="threat-management" element={<Navigate to="/threat-management/alerts" replace />} />
        <Route path="threat-management/alerts" element={<AlertsPage />} />
        <Route path="threat-management/alerts/tagging-rules" element={<TaggingRulesPage />} />
        <Route path="threat-management/alerts/:id" element={<AlertsPage />} />
        <Route path="threat-management/incidents" element={<IncidentsPage />} />
        <Route path="threat-management/adversaries" element={<AdversariesPage />} />

        {/* SOAR */}
        <Route path="soar" element={<Navigate to="/soar/flows" replace />} />
        <Route path="soar/flows" element={<FlowsPage />} />
        <Route path="soar/execution-history" element={<ExecutionHistoryPage />} />
        <Route path="soar/interactive-console" element={<InteractiveConsolePage />} />
        {/* Legacy redirects */}
        <Route path="soar/console" element={<Navigate to="/soar/interactive-console" replace />} />
        <Route path="soar/audit" element={<Navigate to="/settings/audit-logs" replace />} />

        <Route path="log-explorer" element={<LogExplorerPage />} />
        <Route path="user-auditor" element={<UserAuditorPage />} />
        <Route path="threat-intelligence" element={<ThreatIntelPage />} />

        {/* Compliance */}
        <Route path="compliance" element={<CompliancePage />} />
        <Route path="compliance/frameworks/:key" element={<FrameworkReportPage />} />
        {/* Legacy redirects */}
        <Route path="compliance/new" element={<Navigate to="/compliance" replace />} />
        <Route path="compliance/schedule" element={<Navigate to="/compliance" replace />} />
        <Route path="compliance/manage" element={<Navigate to="/compliance" replace />} />

        {/* Configure */}
        <Route path="datasources" element={<DataSourcesPage />} />
        <Route path="integrations" element={<IntegrationsPage />} />
        <Route path="alerting-rules" element={<AlertingRulesPage />} />
        <Route path="pipelines" element={<ParsingFiltersPage />} />
        <Route path="data-processing" element={<DataProcessingPage />} />
        {/* Federation serves its own flat team page from the admin chrome (top-level). */}
        {!IS_FEDERATION && <Route path="team" element={<TeamPage />} />}
        {/* Settings — drill-down sub-pages */}
        <Route path="settings" element={<SettingsIndex />} />
        {/* The licence, the retention policy and the build are properties of the
            instance, not of a customer on it. A tenant is shown neither. */}
        <Route
          path="settings/license"
          element={
            <PlatformOnly>
              <LicensePage />
            </PlatformOnly>
          }
        />
        <Route path="settings/notifications" element={<NotificationsPage />} />
        {/* Legacy redirect — the workspaces module was removed */}
        <Route path="settings/workspaces" element={<Navigate to="/settings" replace />} />
        <Route path="settings/connection-key" element={<ConnectionKeyPage />} />
        <Route
          path="settings/data-retention"
          element={
            <PlatformOnly>
              <DataRetentionPage />
            </PlatformOnly>
          }
        />
        {/* The mirror image of /tenants: what a customer grants, not what the
            operator takes. The platform's own tenant has nobody to grant it to,
            and the backend refuses it there. */}
        <Route
          path="settings/support-access"
          element={
            <TenantOnly>
              <SupportAccessPage />
            </TenantOnly>
          }
        />
        <Route path="settings/branding" element={<BrandingPage />} />
        <Route path="settings/theme" element={<Navigate to="/settings/branding" replace />} />
        {/* API keys are per-user (a key authenticates AS its owner). Not available in federation. */}
        {!IS_FEDERATION && <Route path="settings/api-keys" element={<ApiKeysPage />} />}
        <Route path="settings/identity-providers" element={<IdentityProvidersPage />} />
        {/* Federation serves email config from its own admin chrome (top-level route). */}
        {!IS_FEDERATION && <Route path="settings/email" element={<EmailConfigurationPage />} />}
        <Route path="settings/soc-ai" element={<SocAiSettingsPage />} />
        <Route path="settings/date-format" element={<DateFormatPage />} />
        <Route path="settings/language" element={<LanguagePage />} />
        <Route path="settings/audit-logs" element={<AuditPage />} />
        <Route
          path="settings/about"
          element={
            <PlatformOnly>
              <AboutPage />
            </PlatformOnly>
          }
        />
        {/* Instance-user profile (normal mode). In federation the account lives at
            the top-level /profile (FederationProfilePage), reachable without an instance. */}
        {!IS_FEDERATION && <Route path="profile" element={<ProfilePage />} />}
        <Route path="audit" element={<Navigate to="/settings/audit-logs" replace />} />

        {/* Legacy redirects */}
        <Route path="users" element={<Navigate to="/team" replace />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

/**
 * Renders its children only for the instance operator. Anyone else is sent
 * home, which is also what a customer's administrator gets if they guess the
 * URL — the backend refuses the calls behind it either way.
 *
 * `requiresMSSP` adds the multitenancy licence on top, for pages that are about
 * having tenants at all. Without it there is only ever one tenant whose
 * administrator *is* the platform identity, so the platform check alone would
 * let them through to a page the API answers 403 to.
 */
function PlatformOnly({
  children,
  requiresMSSP,
}: {
  children: React.ReactNode
  requiresMSSP?: boolean
}) {
  const { isPlatformAdmin, isLoading } = useAuth()
  const mssp = useMSSP()
  if (isLoading || (requiresMSSP && mssp === 'pending')) return null
  if (!isPlatformAdmin || (requiresMSSP && !mssp)) return <Navigate to="/home" replace />
  return <>{children}</>
}

/**
 * The complement of PlatformOnly: pages that only mean something to a customer
 * tenant. The instance operator is sent home — they have nobody to grant access
 * to, and the backend answers 403 there anyway.
 */
function TenantOnly({ children }: { children: React.ReactNode }) {
  const { isPlatformAdmin, isAdmin, isLoading } = useAuth()
  const mssp = useMSSP()
  if (isLoading || mssp === 'pending') return null
  if (isPlatformAdmin || !isAdmin || !mssp) return <Navigate to="/home" replace />
  return <>{children}</>
}

/**
 * Whether this install is licensed for multitenancy, with an explicit "not yet
 * known" so a guard holds instead of bouncing a page refresh that arrives
 * before the licence does. An unreadable licence counts as no.
 */
function useMSSP(): boolean | 'pending' {
  const { license, error } = useBilling()
  if (license === null && !error) return 'pending'
  return license?.mssp === true
}

/**
 * Where /settings lands. The licence page is the operator's home for this area
 * but is hidden from tenants, so sending everyone there would bounce a customer
 * administrator straight back out.
 */
function SettingsIndex() {
  const { isPlatformAdmin, isLoading } = useAuth()
  if (isLoading) return null
  return <Navigate to={isPlatformAdmin ? '/settings/license' : '/settings/notifications'} replace />
}
