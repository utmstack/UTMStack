import {ComplianceEvaluationRuleEnum} from '../enums/compliance-evaluation-rule.enum';

export class ComplianceQueryBaseType {
  id?: number;
  queryName?: string;
  queryDescription?: string;
  evaluationRule?: ComplianceEvaluationRuleEnum;
}
