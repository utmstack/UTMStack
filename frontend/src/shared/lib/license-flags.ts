// Module-level mirror of the MSSP license flag so non-React code (URL/host
// helpers used inside plain functions) can read it synchronously. Updated by
// BillingProvider on every license refresh.

let mssp = false

export function isMssp(): boolean {
  return mssp
}

export function setMsspFlag(value: boolean): void {
  mssp = value
}
