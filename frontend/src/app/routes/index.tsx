import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute } from '@/features/auth'
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
import { RegexPatternsPage } from '@/features/regex-patterns/pages/RegexPatternsPage'
import { DataProcessingPage } from '@/features/data-processing/pages/DataProcessingPage'
import { TeamPage } from '@/features/team/pages/TeamPage'
import { ConnectionKeyPage } from '@/features/settings/pages/ConnectionKeyPage'
import { DataRetentionPage } from '@/features/settings/pages/DataRetentionPage'
import { IndicesPage } from '@/features/settings/pages/IndicesPage'
import { BrandingPage } from '@/features/branding'
import { IdentityProvidersPage } from '@/features/settings/pages/IdentityProvidersPage'
import { EmailConfigurationPage } from '@/features/settings/pages/EmailConfigurationPage'
import { SocAiSettingsPage } from '@/features/settings/pages/SocAiSettingsPage'
import { DateFormatPage } from '@/features/settings/pages/DateFormatPage'
import { LanguagePage } from '@/features/settings/pages/LanguagePage'
import { AboutPage } from '@/features/settings/pages/AboutPage'
import { LicensePage } from '@/features/billing'
import { NotificationsPage } from '@/features/notifications'
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
        <Route path="regex-patterns" element={<RegexPatternsPage />} />
        <Route path="data-processing" element={<DataProcessingPage />} />
        {/* Federation serves its own flat team page from the admin chrome (top-level). */}
        {!IS_FEDERATION && <Route path="team" element={<TeamPage />} />}
        {/* Settings — drill-down sub-pages */}
        <Route path="settings" element={<Navigate to="/settings/license" replace />} />
        <Route path="settings/license" element={<LicensePage />} />
        <Route path="settings/notifications" element={<NotificationsPage />} />
        {/* Legacy redirect — the workspaces module was removed */}
        <Route path="settings/workspaces" element={<Navigate to="/settings/license" replace />} />
        <Route path="settings/connection-key" element={<ConnectionKeyPage />} />
        <Route path="settings/data-retention" element={<DataRetentionPage />} />
        <Route path="settings/indices" element={<IndicesPage />} />
        {/* Legacy redirects */}
        <Route path="settings/index-management" element={<Navigate to="/settings/indices" replace />} />
        <Route path="settings/index-patterns" element={<Navigate to="/settings/indices" replace />} />
        <Route path="settings/branding" element={<BrandingPage />} />
        <Route path="settings/theme" element={<Navigate to="/settings/branding" replace />} />
        {/* API keys are per-user (a key authenticates AS its owner), so they live
            embedded in the profile page, not system settings. */}
        <Route path="settings/api-keys" element={<Navigate to="/profile" replace />} />
        <Route path="settings/identity-providers" element={<IdentityProvidersPage />} />
        {/* Federation serves email config from its own admin chrome (top-level route). */}
        {!IS_FEDERATION && <Route path="settings/email" element={<EmailConfigurationPage />} />}
        <Route path="settings/soc-ai" element={<SocAiSettingsPage />} />
        <Route path="settings/date-format" element={<DateFormatPage />} />
        <Route path="settings/language" element={<LanguagePage />} />
        <Route path="settings/audit-logs" element={<AuditPage />} />
        <Route path="settings/about" element={<AboutPage />} />
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
