import {
  SectionConfigParamType
} from '../../../shared/types/configuration/section-config-param.type';
import {SectionConfigType} from '../../../shared/types/configuration/section-config.type';
import {ConfigDataTypeEnum} from '../../../shared/types/configuration/section-config-param.type';

// SAFETY: mirrors backend DB row for sectionId 3; fields checked in app-config-sections
const TFA_SECTION = {
  id: 3,
  section: 'Two Factor Authentication',
  description: 'Two-Factor Authentication adds an extra layer of security by requiring two forms of verification. Choose email or authenticator app. Once enabled, you\'ll confirm your identity with your selected method at each login.',
  shortName: 'TFA'
} as unknown as SectionConfigType;

export const FEDERATION_TFA_CONFIG_SECTION: SectionConfigType = TFA_SECTION;

export const FEDERATION_TFA_CONFIG_PARAMS: SectionConfigParamType[] = ([
  {
    id: 17,
    sectionId: 3,
    confParamShort: 'utmstack.tfa.enable',
    confParamLarge: 'Enable Two Factor Authentication',
    confParamDescription: null,
    confParamValue: 'false',
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Bool,
    confParamRestartRequired: false,
    confParamOption: null,
    section: TFA_SECTION
  },
  {
    id: 10000,
    sectionId: 3,
    confParamShort: 'utmstack.tfa.method',
    confParamLarge: 'Verification method',
    confParamDescription: 'Select the verification method used for two-factor authentication.',
    confParamValue: null,
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Radio,
    confParamRestartRequired: false,
    confParamOption: 'EMAIL,TOTP',
    section: TFA_SECTION
  }
] as unknown) as SectionConfigParamType[];
