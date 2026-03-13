import {ComplianceControlEvaluationsType} from './compliance-control-evaluations.type';

export interface ComplianceControlEvaluationsResponse {
  startDate: string;
  endDate: string;
  evaluations: ComplianceControlEvaluationsType[];
}
