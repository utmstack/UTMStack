import {ComplianceQueryBaseType} from './compliance-query-base.type';

export class ComplianceQueryType extends ComplianceQueryBaseType {
  controlConfigId?: number;
  ruleValue?: number;
  sqlQuery?: string;
  indexPatternId?: number;
}
