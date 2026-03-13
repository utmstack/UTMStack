import {ComplianceEvaluationRuleEnum} from '../enums/compliance-evaluation-rule.enum';
import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
// TODO: ELENA ver si puedo usar ComplianceQueryConfigType
export class ComplianceQueryEvaluationType {
  queryName?: string;
  queryDescription?: string;
  evaluationRule?: ComplianceEvaluationRuleEnum;
  indexPatternName?: string;
  hits?: number;
  status?: ComplianceStatusExtendedEnum;
  evidence?: any[];
}
