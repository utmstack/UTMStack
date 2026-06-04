import {ComplianceStatusExtendedEnum} from '../enums/compliance-status.enum';
import {ComplianceControlBaseType} from './compliance-control-base.type';

export class ComplianceControlLatestEvaluationType extends ComplianceControlBaseType {
  lastEvaluationStatus?: ComplianceStatusExtendedEnum;
  lastEvaluationTimestamp?: string;
}
