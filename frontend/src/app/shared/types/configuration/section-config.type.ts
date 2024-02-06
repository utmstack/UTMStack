export class SectionConfigType {
  description?: string;
  id?: 0;
  section?: string;
}

export enum ApplicationConfigSectionEnum {
  SMS = 1,
  Email = 2,
  TwoFactorAuthentication = 3,
  Alerts = 4,
  DateSettings = 5
}
