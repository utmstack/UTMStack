import { IS_FEDERATION } from '@/shared/config/mode'
import { authHttpService } from '@/features/auth/services/auth-http.service'
import { federationAuthService } from '@/features/federation/services/federation-auth.service'
import type {
  TfaCompleteRequest,
  TfaDisableRequest,
  TfaInitRequest,
  TfaInitResponse,
  TfaMethod,
  TfaRefreshResponse,
  TfaVerifyRequest,
} from '@/features/auth/types/auth.types'

/**
 * 2FA enrollment/teardown, routed to whichever backend owns the account: the FS
 * (federation mode, TOTP only) or the instance. Both expose the same surface so
 * TwoFactorBody drives the flow identically. EMAIL is FS-unsupported.
 */
export interface TfaService {
  tfaInit(input: TfaInitRequest): Promise<TfaInitResponse>
  tfaVerify(input: TfaVerifyRequest): Promise<void>
  tfaComplete(input: TfaCompleteRequest): Promise<void>
  tfaDisable(input: TfaDisableRequest): Promise<void>
  tfaRefreshChallenge(method: TfaMethod): Promise<TfaRefreshResponse>
}

export const tfaService: TfaService = IS_FEDERATION
  ? {
      tfaInit: () => federationAuthService.tfaInit(),
      tfaVerify: ({ code }) => federationAuthService.tfaVerify(code),
      tfaComplete: () => federationAuthService.tfaComplete(),
      tfaDisable: ({ password }) => federationAuthService.tfaDisable(password),
      tfaRefreshChallenge: () =>
        Promise.reject(new Error('email 2FA is not supported in federation mode')),
    }
  : {
      tfaInit: authHttpService.tfaInit,
      tfaVerify: authHttpService.tfaVerify,
      tfaComplete: authHttpService.tfaComplete,
      tfaDisable: authHttpService.tfaDisable,
      tfaRefreshChallenge: authHttpService.tfaRefreshChallenge,
    }
