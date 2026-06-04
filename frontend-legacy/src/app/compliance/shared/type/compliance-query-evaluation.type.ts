import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
import {ComplianceQueryBaseType} from './compliance-query-base.type';

export class ComplianceQueryEvaluationType extends ComplianceQueryBaseType {
  indexPatternName?: string;
  hits?: number;
  status?: ComplianceStatusExtendedEnum;
  evidence?: any[];
}
