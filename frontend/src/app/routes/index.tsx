import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute } from '@/features/auth'
import { LoginPage } from '@/features/auth/pages/LoginPage'
import { HomePage } from '@/features/home/pages/HomePage'
import { ProfilePage } from '@/features/profile/pages/ProfilePage'
import { AuditPage } from '@/features/audit/pages/AuditPage'
import { AlertsPage } from '@/features/threats/pages/AlertsPage'
import { IncidentsPage } from '@/features/threats/pages/IncidentsPage'
import { AdversariesPage } from '@/features/threats/pages/AdversariesPage'
import { FlowsPage } from '@/features/soar/pages/FlowsPage'
import { InteractiveConsolePage } from '@/features/soar/pages/InteractiveConsolePage'
import { LogExplorerPage } from '@/features/log-explorer/pages/LogExplorerPage'
import { UserAuditorPage } from '@/features/user-auditor/pages/UserAuditorPage'
import { ThreatIntelPage } from '@/features/threat-intel/pages/ThreatIntelPage'
import { CompliancePage } from '@/features/compliance/pages/CompliancePage'
import { DataSourcesPage } from '@/features/datasources/pages/DataSourcesPage'
import { IntegrationsPage } from '@/features/integrations/pages/IntegrationsPage'
import { AlertingRulesPage } from '@/features/alerting-rules/pages/AlertingRulesPage'
import { TaggingRulesPage } from '@/features/tagging-rules/pages/TaggingRulesPage'
import { DataProcessingPage } from '@/features/data-processing/pages/DataProcessingPage'
import { TeamPage } from '@/features/team/pages/TeamPage'
import { ConnectionKeyPage } from '@/features/settings/pages/ConnectionKeyPage'
import { DataRetentionPage } from '@/features/settings/pages/DataRetentionPage'
import { IndicesPage } from '@/features/settings/pages/IndicesPage'
import { BrandingPage } from '@/features/branding'
import { IdentityProvidersPage } from '@/features/settings/pages/IdentityProvidersPage'
import { EmailConfigurationPage } from '@/features/settings/pages/EmailConfigurationPage'
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

      <Route
        path="/"
        element={
          <ProtectedRoute>
            <DashboardLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/home" replace />} />
        <Route path="home" element={<HomePage />} />

        {/* Threat Management */}
        <Route path="threat-management" element={<Navigate to="/threat-management/alerts" replace />} />
        <Route path="threat-management/alerts" element={<AlertsPage />} />
        <Route path="threat-management/incidents" element={<IncidentsPage />} />
        <Route path="threat-management/adversaries" element={<AdversariesPage />} />

        {/* SOAR */}
        <Route path="soar" element={<Navigate to="/soar/flows" replace />} />
        <Route path="soar/flows" element={<FlowsPage />} />
        <Route path="soar/console" element={<InteractiveConsolePage />} />
        <Route path="soar/audit" element={<Navigate to="/settings/audit-logs" replace />} />

        <Route path="log-explorer" element={<LogExplorerPage />} />
        <Route path="user-auditor" element={<UserAuditorPage />} />
        <Route path="threat-intelligence" element={<ThreatIntelPage />} />

        {/* Compliance */}
        <Route path="compliance" element={<CompliancePage />} />
        {/* Legacy redirects */}
        <Route path="compliance/new" element={<Navigate to="/compliance" replace />} />
        <Route path="compliance/schedule" element={<Navigate to="/compliance" replace />} />
        <Route path="compliance/manage" element={<Navigate to="/compliance" replace />} />

        {/* Configure */}
        <Route path="datasources" element={<DataSourcesPage />} />
        <Route path="integrations" element={<IntegrationsPage />} />
        <Route path="alerting-rules" element={<AlertingRulesPage />} />
        <Route path="tagging-rules" element={<TaggingRulesPage />} />
        <Route path="data-processing" element={<DataProcessingPage />} />
        <Route path="team" element={<TeamPage />} />
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
        <Route path="settings/email" element={<EmailConfigurationPage />} />
        <Route path="settings/date-format" element={<DateFormatPage />} />
        <Route path="settings/language" element={<LanguagePage />} />
        <Route path="settings/audit-logs" element={<AuditPage />} />
        <Route path="settings/about" element={<AboutPage />} />
        <Route path="profile" element={<ProfilePage />} />
        <Route path="audit" element={<Navigate to="/settings/audit-logs" replace />} />

        {/* Legacy redirects */}
        <Route path="users" element={<Navigate to="/team" replace />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
