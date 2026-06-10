// Trailing result suffix (redundant with the status pill) — stripped before labeling.
const STATUS_SUFFIX = /\.(success|fail|failure|attempt)$/

// Curated labels for the common / awkward actions. Anything not here falls back
// to a generic title-cased phrase, so new actions still render reasonably.
const ACTION_LABELS: Record<string, string> = {
  'auth.login': 'Signed in',
  'auth.logout': 'Signed out',
  'auth.change_password': 'Changed password',
  'auth.me.update': 'Updated profile',
  'auth.reset_password.init': 'Requested password reset',
  'auth.reset_password.finish': 'Completed password reset',
  'auth.saml': 'SSO sign-in',
  'apikey.create': 'Created API key',
  'apikey.delete': 'Deleted API key',
  'license.upload': 'Uploaded license',
  'config.updated': 'Updated configuration',
  'config.mail.checked': 'Tested mail settings',
  'branding.updated': 'Updated branding',
  'branding.asset.uploaded': 'Uploaded branding asset',
  'notification.delete': 'Deleted notification',
  'notification.read_all': 'Marked all notifications read',
  'incident.create': 'Created incident',
  'incident.status.change': 'Changed incident status',
  'incident.alert.add': 'Added alert to incident',
}

/** Turn a raw action code ("notification.delete.success") into a human label. */
export function humanizeAction(action: string): string {
  const base = action.replace(STATUS_SUFFIX, '')
  const mapped = ACTION_LABELS[base]
  if (mapped) return mapped
  const text = base.replace(/_/g, ' ').replace(/\./g, ' ').trim()
  return text.charAt(0).toUpperCase() + text.slice(1)
}
