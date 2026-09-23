const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const TEMPLATE_RE = /^(?:\$\([^)]*\)|\$\[[^\]]*\])$/

export interface MailParams {
  to?: string
  cc?: string
}

export type MailRecipientField = 'to' | 'cc'

export interface MailRecipientError {
  field: MailRecipientField
  address: string
}

export function splitMailRecipients(value: string | undefined): string[] {
  return (value ?? '')
    .split(',')
    .map((address) => address.trim())
    .filter(Boolean)
}

export function isValidMailRecipient(address: string): boolean {
  const trimmed = address.trim()
  return EMAIL_RE.test(trimmed) || TEMPLATE_RE.test(trimmed)
}

export function mailRecipientError(params: MailParams): MailRecipientError | undefined {
  for (const field of ['to', 'cc'] as const) {
    const address = splitMailRecipients(params[field]).find((recipient) => !isValidMailRecipient(recipient))
    if (address) return { field, address }
  }
  return undefined
}