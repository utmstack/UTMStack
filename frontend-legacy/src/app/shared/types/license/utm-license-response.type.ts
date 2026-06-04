import {UtmLicenceEnum} from './enum/utm-licence.enum';

export class UtmLicenseResponseType {
  data: {
    createdAt: string,
    expiresAt: string,
    licenseKey: string,
    remainingActivations: number,
    status: UtmLicenceEnum,
    timesActivated: number,
    timesActivatedMax: number,
    updatedAt: string,
    validFor: number
  };
  success: boolean;
}
