import {ComplianceQueryEvaluationType} from './compliance-query-evaluation.type';

export class ComplianceIndexPatternQueryGroupEvaluationType {
  indexPatternId?: number;
  indexPatternName?: string;
  queries?: ComplianceQueryEvaluationType[];
}

