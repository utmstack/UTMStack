import { IS_FEDERATION } from '@/shared/config/mode'
import { authHttpService } from '@/features/auth/services/auth-http.service'
import { federationAuthService } from '@/features/federation/services/federation-auth.service'
import type {
  TfaDisableRequest,
  TfaEnrollmentRequest,
  TfaEnrollmentResponse,
} from '@/features/auth/types/auth.types'

/**
 * 2FA enrollment/teardown, routed to whichever backend owns the account: the FS
 * (federation mode, TOTP only) or the instance. The instance takes the three
 * stages on one endpoint; the FS still has one call per stage, so this adapts
 * them onto the same surface and TwoFactorBody drives both identically.
 */
export interface TfaService {
  enroll(input: TfaEnrollmentRequest): Promise<TfaEnrollmentResponse>
  tfaDisable(input: TfaDisableRequest): Promise<void>
}

async function federationEnroll(input: TfaEnrollmentRequest): Promise<TfaEnrollmentResponse> {
  if (input.type !== 'totp') {
    throw new Error('only an authenticator app is supported in federation mode')
  }
  switch (input.stage) {
    case 'INIT':
      return { stage: input.stage, init: await federationAuthService.tfaInit() }
    case 'VERIFY':
      await federationAuthService.tfaVerify(input.code ?? '')
      return { stage: input.stage, verified: true }
    case 'COMPLETE':
      await federationAuthService.tfaComplete()
      return { stage: input.stage, enabled: true }
  }
}

export const tfaService: TfaService = IS_FEDERATION
  ? {
      enroll: federationEnroll,
      tfaDisable: ({ password }) => federationAuthService.tfaDisable(password),
    }
  : {
      enroll: authHttpService.tfaEnrollment,
      tfaDisable: authHttpService.tfaDisable,
    }
