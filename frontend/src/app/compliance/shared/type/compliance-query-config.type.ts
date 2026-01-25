import {ComplianceEvaluationRuleEnum} from '../enums/compliance-evaluation-rule.enum';

export class UtmComplianceQueryConfigType {
  id?: number;
  queryName?: string;
  queryDescription?: string;
  sqlQuery?: string;
  evaluationRule?: ComplianceEvaluationRuleEnum;
  indexPatternId?: number;
  controlConfigId?: number;
}
