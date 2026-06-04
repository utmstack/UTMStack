import {IncidentRuleType} from './incident-rule.type';

export class IraHistoryType {
  id: number;
  ruleId: number;
  createdDate: Date;
  previousState: IncidentRuleType;
}
