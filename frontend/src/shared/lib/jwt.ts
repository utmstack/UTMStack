/**
 * Minimal client-side JWT payload decode. Used ONLY to read the user's roles /
 * permissions for UX gating (hide/disable). NOT a security boundary — the backend
 * enforces every permission server-side (RequirePermission). The signature is not
 * verified here.
 */

interface AccessTokenClaims {
  roles?: string[]
  permissions?: string[]
  /** The instance operator, as the backend computes it: the default tenant plus
   * the admin role. A tenant's own administrator also holds ROLE_ADMIN, so roles
   * alone cannot tell them apart. */
  platform?: boolean
  /** Which tenant this session is signed into (`tid`). */
  tenantId?: string
}

function base64UrlDecode(segment: string): string {
  const b64 = segment.replace(/-/g, '+').replace(/_/g, '/')
  const pad = b64.length % 4 === 0 ? '' : '='.repeat(4 - (b64.length % 4))
  return atob(b64 + pad)
}

export function decodeAccessToken(token: string | undefined | null): AccessTokenClaims {
  if (!token) return {}
  const parts = token.split('.')
  if (parts.length !== 3) return {}
  try {
    const json = base64UrlDecode(parts[1])
    const claims = JSON.parse(json) as AccessTokenClaims & { tid?: string }
    return {
      roles: Array.isArray(claims.roles) ? claims.roles : [],
      permissions: Array.isArray(claims.permissions) ? claims.permissions : [],
      platform: claims.platform === true,
      tenantId: typeof claims.tid === 'string' ? claims.tid : undefined,
    }
  } catch {
    return {}
  }
}
