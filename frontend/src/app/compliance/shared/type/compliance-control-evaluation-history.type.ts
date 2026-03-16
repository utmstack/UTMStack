import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
import {ComplianceQueryEvaluationGroup} from './compliance-query-evaluation.group';
import {ComplianceQueryEvaluationType} from './compliance-query-evaluation.type';

export class ComplianceControlEvaluationHistoryType {
  controlId?: number;
  controlName?: string;
  status?: ComplianceStatusExtendedEnum;
  timestamp?: string;
  queryEvaluations?: ComplianceQueryEvaluationType[] | ComplianceQueryEvaluationGroup[];
}
