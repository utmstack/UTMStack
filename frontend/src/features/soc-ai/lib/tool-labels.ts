/**
 * Human-readable progress labels for MCP tool names, shown in the chat while
 * the agent is working a step. Chat users should never see a raw dotted tool
 * name like "visualizations.create" — it reads as internal/debug noise.
 */
const TOOL_LABELS: Record<string, string> = {
  // dashboards / visualizations
  'dashboards.create': 'Creating dashboard',
  'dashboards.update': 'Updating dashboard',
  'dashboards.delete': 'Deleting dashboard',
  'dashboards.get': 'Loading dashboard',
  'dashboards.list': 'Listing dashboards',
  'visualizations.create': 'Adding widget',
  'visualizations.update': 'Updating widget',
  'visualizations.delete': 'Removing widget',
  'visualizations.get': 'Loading widget',
  'visualizations.list': 'Listing widgets',

  // alerts
  'alerts.convert_to_incident': 'Converting alert to incident',
  'alerts.count_open': 'Counting open alerts',
  'alerts.search_adversary': 'Looking up adversary info',
  'alerts.update_notes': 'Updating alert notes',
  'alerts.update_status': 'Updating alert status',
  'alerts.update_tags': 'Updating alert tags',
  'alerts.score': 'Scoring alert',
  'alert_tags.create': 'Creating alert tag',
  'alert_tags.update': 'Updating alert tag',
  'alert_tags.delete': 'Deleting alert tag',
  'alert_tags.get': 'Loading alert tag',
  'alert_tags.list': 'Listing alert tags',
  'alert_tag_rules.create': 'Creating tag rule',
  'alert_tag_rules.update': 'Updating tag rule',
  'alert_tag_rules.delete': 'Deleting tag rule',
  'alert_tag_rules.get': 'Loading tag rule',
  'alert_tag_rules.list': 'Listing tag rules',

  // incidents
  'incidents.create': 'Creating incident',
  'incidents.get': 'Loading incident',
  'incidents.list': 'Listing incidents',
  'incidents.add_alerts': 'Adding alerts to incident',
  'incidents.assignees': 'Loading assignees',
  'incidents.change_status': 'Updating incident status',
  'incident_alerts.create': 'Linking alert to incident',
  'incident_alerts.update': 'Updating incident alert',
  'incident_alerts.update_status': 'Updating incident alert status',
  'incident_alerts.delete': 'Removing alert from incident',
  'incident_alerts.list': 'Listing incident alerts',
  'incident_notes.create': 'Adding incident note',
  'incident_notes.list': 'Listing incident notes',
  'incident_history.get': 'Loading incident history',
  'incident_history.list': 'Listing incident history',
  'incident_history.count': 'Counting incident history',

  // log/event store
  'store.datasets': 'Listing datasets',
  'store.dataset.fields': 'Inspecting fields',
  'store.count': 'Counting events',
  'store.search': 'Searching events',
  'store.search_csv': 'Exporting results',
  'store.search_sql': 'Running query',
  'store.property_values': 'Loading field values',
  'loganalyzer.chart_view': 'Building chart',
  'loganalyzer.top_values': 'Finding top values',
  'loganalyzer.query.create': 'Saving query',
  'loganalyzer.query.update': 'Updating query',
  'loganalyzer.query.delete': 'Deleting query',
  'loganalyzer.query.get': 'Loading query',
  'loganalyzer.query.list': 'Listing saved queries',
  'ingestion_stats.timeline': 'Loading ingestion timeline',
  'ingestion_stats.totals': 'Loading ingestion totals',

  // compliance
  'compliance.control.create': 'Creating control',
  'compliance.control.update': 'Updating control',
  'compliance.control.delete': 'Deleting control',
  'compliance.control.get': 'Loading control',
  'compliance.control.list': 'Listing controls',
  'compliance.framework.create': 'Creating framework',
  'compliance.framework.update': 'Updating framework',
  'compliance.framework.delete': 'Deleting framework',
  'compliance.framework.get': 'Loading framework',
  'compliance.framework.list': 'Listing frameworks',
  'compliance.report.evaluate': 'Evaluating compliance',
  'compliance.report.get': 'Loading compliance report',
  'compliance.report.list': 'Listing compliance reports',
  'compliance.schedule.create': 'Scheduling report',
  'compliance.schedule.update': 'Updating schedule',
  'compliance.schedule.delete': 'Deleting schedule',
  'compliance.schedule.get': 'Loading schedule',
  'compliance.schedule.list_by_user': 'Listing schedules',

  // datasources / integrations / correlation
  'datasources.count': 'Counting data sources',
  'datasources.delete': 'Deleting data source',
  'datasources.get': 'Loading data source',
  'datasources.list': 'Listing data sources',
  'datasources.update_labels': 'Updating data source labels',
  'integrations.create': 'Creating integration',
  'integrations.update': 'Updating integration',
  'integrations.delete': 'Deleting integration',
  'integrations.get': 'Loading integration',
  'integrations.get_by_name': 'Loading integration',
  'integrations.list': 'Listing integrations',
  'integrations.data_types': 'Loading data types',
  'integrations.config.save': 'Saving integration config',
  'integrations.config.delete': 'Deleting integration config',
  'integrations.config.list': 'Listing integration configs',
  'correlation_rule.create': 'Creating correlation rule',
  'correlation_rule.update': 'Updating correlation rule',
  'correlation_rule.delete': 'Deleting correlation rule',
  'correlation_rule.get': 'Loading correlation rule',
  'correlation_rule.list': 'Listing correlation rules',
  'correlation_rule.set_active': 'Toggling correlation rule',
  'correlation_rule.search_property_values': 'Searching property values',
  'regex_pattern.get': 'Loading regex pattern',
  'regex_pattern.list': 'Listing regex patterns',
  'pipeline.create': 'Creating pipeline',
  'pipeline.update': 'Updating pipeline',
  'pipeline.delete': 'Deleting pipeline',
  'pipeline.get': 'Loading pipeline',
  'pipeline.list': 'Listing pipelines',
  'pipeline.set_active': 'Toggling pipeline',

  // soar
  'soar.rule.create': 'Creating automation rule',
  'soar.rule.update': 'Updating automation rule',
  'soar.rule.delete': 'Deleting automation rule',
  'soar.rule.get': 'Loading automation rule',
  'soar.rule.list': 'Listing automation rules',
  'soar.rule.set_enabled': 'Toggling automation rule',
  'soar.rule.resolve_filter_values': 'Resolving filter values',
  'soar.variable.create': 'Creating variable',
  'soar.variable.update': 'Updating variable',
  'soar.variable.delete': 'Deleting variable',
  'soar.variable.get': 'Loading variable',
  'soar.variable.list': 'Listing variables',
  'soar.execution.list': 'Listing automation runs',
  'soar.agent.list_by_platform': 'Listing agents',

  // notifications / tenants / socai / adaudit / iam
  'notifications.list': 'Loading notifications',
  'notifications.get': 'Loading notification',
  'notifications.delete': 'Deleting notification',
  'notifications.mark_read': 'Marking notification as read',
  'notifications.mark_all_read': 'Marking all notifications as read',
  'notifications.unread_count': 'Counting unread notifications',
  'notifications.update_status': 'Updating notification',
  'tenants.create': 'Creating tenant',
  'tenants.update': 'Updating tenant',
  'tenants.get': 'Loading tenant',
  'tenants.list': 'Listing tenants',
  'tenants.terminate': 'Terminating tenant',
  'tenants.set_support_access': 'Updating support access',
  'socai.analyze_alert': 'Analyzing alert',
  'adaudit.users.list': 'Listing AD users',
  'iam.user.create': 'Creating user',
  'iam.user.update': 'Updating user',
  'iam.user.get': 'Loading user',
  'iam.user.list': 'Listing users',
  'iam.user.deactivate': 'Deactivating user',
  'iam.user.assign_roles': 'Assigning roles',
  'iam.role.get': 'Loading role',
  'iam.role.list': 'Listing roles',
  'iam.apikey.create': 'Creating API key',
  'iam.apikey.update': 'Updating API key',
  'iam.apikey.delete': 'Deleting API key',
  'iam.apikey.get': 'Loading API key',
  'iam.apikey.list': 'Listing API keys',
  'iam.apikey.generate': 'Generating API key',
  'iam.idp.create': 'Creating identity provider',
  'iam.idp.update': 'Updating identity provider',
  'iam.idp.delete': 'Deleting identity provider',
  'iam.idp.get': 'Loading identity provider',
  'iam.idp.list': 'Listing identity providers',
  'iam.auth.me': 'Loading your profile',
  'iam.auth.update_me': 'Updating your profile',
  'iam.auth.change_password': 'Updating password',
  'iam.auth.list_sessions': 'Listing sessions',
  'iam.auth.revoke_session': 'Revoking session',
  'iam.auth.revoke_other_sessions': 'Revoking other sessions',

  // misc admin
  'audit.get': 'Loading audit entry',
  'audit.list': 'Listing audit log',
  'billing.license': 'Loading license',
  'billing.version': 'Loading version info',
  'branding.get': 'Loading branding',
  'branding.get_public': 'Loading branding',
  'branding.update': 'Updating branding',
  'config.check_mail': 'Testing email settings',
  'config.get': 'Loading settings',
  'config.list': 'Listing settings',
  'config.update': 'Updating settings',
}

const VERB_LABELS: Record<string, string> = {
  create: 'Creating',
  update: 'Updating',
  delete: 'Deleting',
  remove: 'Removing',
  get: 'Loading',
  list: 'Listing',
  count: 'Counting',
  search: 'Searching',
  generate: 'Generating',
  evaluate: 'Evaluating',
  bulk_create: 'Creating',
  bulk_update: 'Updating',
  bulk_delete: 'Deleting',
  bulk_set_active: 'Updating',
}

function titleCase(word: string): string {
  return word.charAt(0).toUpperCase() + word.slice(1)
}

function humanizeResource(segments: string[]): string {
  return segments.join(' ').replace(/_/g, ' ')
}

/**
 * Best-effort friendly label for any tool name, including ones not in
 * TOOL_LABELS (e.g. a newly added backend tool). Splits "resource.verb" into
 * a "Verb-ing resource" phrase; falls back to the raw name only if it can't
 * even do that, which should be effectively never given the dotted
 * `resource.verb` convention every MCP tool follows.
 */
export function humanizeToolLabel(tool: string): string {
  const exact = TOOL_LABELS[tool]
  if (exact) return exact

  // Not a dotted "resource.verb" MCP tool id (e.g. the compaction step's
  // already-human placeholder text) — nothing to humanize, pass through.
  if (!tool.includes('.')) return tool

  const segments = tool.split('.')
  const verb = segments[segments.length - 1]
  const resource = humanizeResource(segments.slice(0, -1))
  const verbLabel = VERB_LABELS[verb]
  if (verbLabel) return `${verbLabel} ${resource}`

  return `${titleCase(verb.replace(/_/g, ' '))} ${resource}`
}
