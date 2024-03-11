import {SectionConfigType} from './section-config.type';

export class SectionConfigParamType {
  confParamDatatype?: ConfigDataTypeEnum;
  confParamDescription?: string;
  confParamLarge?: string;
  confParamRequired?: true;
  confParamShort?: string;
  confParamValue?: any;
  id?: number;
  modificationTime?: Date;
  modificationUser?: string;
  sectionId?: number;
  section?: SectionConfigType;
  confParamRestartRequired: boolean;
  confParamOption: string;
  confParamRegexp?: string;
}

export enum ConfigDataTypeEnum {
  Text = 'text',
  Tel = 'tel',
  Password = 'password',
  Email = 'email',
  EmailList = 'email_list',
  Number = 'number',
  Bool = 'bool',
  List = 'list',
  TimezoneList = 'timezone_list',
  DateFormatList = 'dateformat_list',
  Radio = 'radio'
}
