import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
import {ComplianceStrategyEnum} from '../enums/compliance-strategy.enum';
import {ComplianceQueryConfigType} from './compliance-query-config.type';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';

export class ComplianceControlEvaluationType {
  //TODO: ELENA current evaluation
  controlId?: number;
  standardSectionId?: number;
  section?: ComplianceStandardSectionType;

  controlName?: string;
  controlSolution?: string;
  controlRemediation?: string;
  controlStrategy?: ComplianceStrategyEnum;

  queriesConfigs?: ComplianceQueryConfigType[];

  lastEvaluationStatus?: ComplianceStatusExtendedEnum;
  lastEvaluationTimestamp?: string;
}
