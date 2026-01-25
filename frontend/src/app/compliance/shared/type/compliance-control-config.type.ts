import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {ComplianceStrategyEnum} from '../enums/compliance-strategy.enum';
import {UtmComplianceQueryConfigType} from './compliance-query-config.type';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';


export class ComplianceControlConfigType {
  id?: number;
  section?: ComplianceStandardSectionType;
  standardSectionId?: number;
  controlName?: string;
  controlSolution?: string;
  controlRemediation?: string;
  controlStrategy?: ComplianceStrategyEnum;
  queriesConfigs?: UtmComplianceQueryConfigType[];

  columns?: UtmFieldType[];
  selected?: boolean;
  status?: string;
}
