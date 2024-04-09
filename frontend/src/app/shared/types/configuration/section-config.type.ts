export class SectionConfigType {
  description?: string;
  id?: 0;
  section?: string;
  shortName?: string;
}

export enum ApplicationConfigSectionEnum {
  SMS = 1,
  EMAIL = 2,
  TFA = 3,
  ALERTS = 4,
  DATE_SETTINGS = 5
}
