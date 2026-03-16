import {ComplianceQueryEvaluationType} from './compliance-query-evaluation.type';

export class ComplianceQueryEvaluationGroup {
  indexPatternId?: number;
  indexPatternName?: string;
  queries?: ComplianceQueryEvaluationType[];
}

