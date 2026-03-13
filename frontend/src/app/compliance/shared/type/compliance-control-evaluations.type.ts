import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
import {ComplianceIndexPatternQueryGroupEvaluationType} from './compliance-index-pattern-query-group-evaluation.type';
import {ComplianceQueryEvaluationType} from './compliance-query-evaluation.type';

export class ComplianceControlEvaluationsType {
  // TODO: ELENA registro historico de evaluaciones
  controlId?: number;
  controlName?: string;
  status?: ComplianceStatusExtendedEnum;
  timestamp?: string;
  queryEvaluations?: ComplianceQueryEvaluationType[] | ComplianceIndexPatternQueryGroupEvaluationType[];
  groupedQueryEvaluations?: ComplianceIndexPatternQueryGroupEvaluationType[]; // TODO: ELENA quitar
}
