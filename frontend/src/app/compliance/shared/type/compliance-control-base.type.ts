import {ComplianceStrategyEnum} from '../enums/compliance-strategy.enum';
import {ComplianceQueryType} from './compliance-query.type';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';

export class ComplianceControlBaseType {
  id?: number;
  section?: ComplianceStandardSectionType;
  standardSectionId?: number;
  controlName?: string;
  controlSolution?: string;
  controlRemediation?: string;
  controlStrategy?: ComplianceStrategyEnum;
  queriesConfigs?: ComplianceQueryType[];
}
