import {UtmLicenceEnum} from './enum/utm-licence.enum';

export class UtmLicenseType {
  details: {
    activation: Date,
    customer_email: string,
    customer_name: string,
    days: 0,
    key: string,
    status: UtmLicenceEnum
  };
  expire: Date;
  valid: boolean;
}
