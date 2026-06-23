import {
  ConfigDataTypeEnum,
  SectionConfigParamType
} from '../../../shared/types/configuration/section-config-param.type';
import {SectionConfigType} from '../../../shared/types/configuration/section-config.type';

const EMAIL_SECTION = ({
  id: 2,
  section: 'email',
  shortName: 'EMAIL',
  description: 'Here you can configure all necessary parameters for email notifications'
} as unknown) as SectionConfigType;

export const FEDERATION_EMAIL_CONFIG_SECTION: SectionConfigType = EMAIL_SECTION;

export const FEDERATION_EMAIL_CONFIG_PARAMS: SectionConfigParamType[] = ([
  {
    id: 4,
    sectionId: 2,
    confParamShort: 'utmstack.mail.password',
    confParamLarge: 'Mail Server Password',
    confParamDescription: 'Login password of the SMTP server',
    confParamValue: null,
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Password,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 5,
    sectionId: 2,
    confParamShort: 'utmstack.mail.from',
    confParamLarge: 'Utmstack email address',
    confParamDescription: 'Address from which emails are sent',
    confParamValue: null,
    confParamRegexp: '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$',
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Email,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 6,
    sectionId: 2,
    confParamShort: 'utmstack.mail.baseUrl',
    confParamLarge: 'Utmstack base url',
    confParamDescription: 'Base url of Utmstack',
    confParamValue: 'https://v11dev2.utmstack.com',
    confParamRegexp: '^(?:https?:\\/\\/)?(?:\\d{1,3}(?:\\.\\d{1,3}){3}|(?:[a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,})(?:\\/[^\\s]*)?$',
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Text,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 7,
    sectionId: 2,
    confParamShort: 'utmstack.mail.host',
    confParamLarge: 'Mail Server Host',
    confParamDescription: 'SMTP server host. For instance, `smtp.example.com`.',
    confParamValue: 'example.hostmail.com',
    confParamRegexp: '^(?:[a-zA-Z0-9-]+\\.)+[a-zA-Z]{2,}|(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$',
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Text,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 8,
    sectionId: 2,
    confParamShort: 'utmstack.mail.port',
    confParamLarge: 'Mail Server Port',
    confParamDescription: 'SMTP server port',
    confParamValue: '587',
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Number,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 9,
    sectionId: 2,
    confParamShort: 'utmstack.mail.username',
    confParamLarge: 'Mail Server Username',
    confParamDescription: 'Login user of the SMTP server',
    confParamValue: null,
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Text,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 10,
    sectionId: 2,
    confParamShort: 'utmstack.mail.organization',
    confParamLarge: 'Organization Name',
    confParamDescription: 'This field helps identify the organization name in incident and alert notification emails.',
    confParamValue: '',
    confParamRegexp: null,
    confParamRequired: false,
    confParamDatatype: ConfigDataTypeEnum.Text,
    confParamRestartRequired: false,
    confParamOption: null,
    section: EMAIL_SECTION
  },
  {
    id: 13,
    sectionId: 2,
    confParamShort: 'utmstack.mail.properties.mail.smtp.auth',
    confParamLarge: 'Encryption type',
    confParamDescription: 'Select the encryption type used by the SMTP server',
    confParamValue: 'STARTTLS',
    confParamRegexp: null,
    confParamRequired: true,
    confParamDatatype: ConfigDataTypeEnum.Radio,
    confParamRestartRequired: false,
    confParamOption: 'STARTTLS,SSL/TLS,NONE',
    section: EMAIL_SECTION
  }
] as unknown) as SectionConfigParamType[];
