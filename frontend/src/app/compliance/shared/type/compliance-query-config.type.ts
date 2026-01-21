import {ComplianceEvaluationRuleEnum} from '../enums/compliance-evaluation-rule.enum';

export class UtmComplianceQueryConfigType {
  id?: number;
  name?: string;
  queryDescription?: string;
  sqlQuery?: string;
  evaluationRule?: ComplianceEvaluationRuleEnum;
  indexPatternId?: number;
  controlConfigId?: number;
}
