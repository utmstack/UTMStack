import administration from './administration.json'
import alertsIncidents from './alerts-incidents.json'
import apiKeys from './api-keys.json'
import audit from './audit.json'
import auth from './auth.json'
import automation from './automation.json'
import compliance from './compliance.json'
import core from './core.json'
import dataProcessing from './data-processing.json'
import dataRetention from './data-retention.json'
import homeDashboard from './home-dashboard.json'
import identityProviders from './identity-providers.json'
import integrations from './integrations.json'
import liveOps from './live-ops.json'
import logExplorer from './log-explorer.json'
import notifications from './notifications.json'
import profile from './profile.json'
import socAi from './soc-ai.json'
import teamAccess from './team-access.json'
import threatIntelligence from './threat-intelligence.json'
import userAuditor from './user-auditor.json'

const translation = {
  ...administration,
  ...alertsIncidents,
  ...apiKeys,
  ...audit,
  ...auth,
  ...automation,
  ...compliance,
  ...core,
  ...dataProcessing,
  ...dataRetention,
  ...homeDashboard,
  ...identityProviders,
  ...integrations,
  ...liveOps,
  ...logExplorer,
  ...notifications,
  ...profile,
  ...socAi,
  ...teamAccess,
  ...threatIntelligence,
  ...userAuditor,
}

export default translation
