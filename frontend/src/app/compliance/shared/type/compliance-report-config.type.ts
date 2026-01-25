import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {ComplianceStrategyEnum} from '../enums/compliance-strategy.enum';
import {UtmComplianceQueryConfigType} from './compliance-query-config.type';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';


export class ComplianceReportConfigType {
  id?: number;
  section?: ComplianceStandardSectionType;
  standardSectionId?: number;
  configReportName?: string;
  configSolution?: string;
  configRemediation?: string;
  configStrategy?: ComplianceStrategyEnum;
  queriesConfigs?: UtmComplianceQueryConfigType[];

  columns?: UtmFieldType[];
  selected?: boolean;
  status?: string;
}
