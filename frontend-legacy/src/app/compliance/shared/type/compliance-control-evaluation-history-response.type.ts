import {ComplianceControlEvaluationHistoryType} from './compliance-control-evaluation-history.type';

export interface ComplianceControlEvaluationHistoryResponse {
  startDate: string;
  endDate: string;
  evaluations: ComplianceControlEvaluationHistoryType[];
}
