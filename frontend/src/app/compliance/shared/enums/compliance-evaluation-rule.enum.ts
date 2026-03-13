export enum ComplianceEvaluationRuleEnum {
  NO_HITS_ALLOWED = 'NO_HITS_ALLOWED',
  MIN_HITS_REQUIRED = 'MIN_HITS_REQUIRED',
  //THRESHOLD_MAX = 'THRESHOLD_MAX',
  //MATCH_FIELD_VALUE = 'MATCH_FIELD_VALUE'
}

export const ComplianceEvaluationRuleLabels = {
  [ComplianceEvaluationRuleEnum.NO_HITS_ALLOWED]: 'NO HITS ALLOWED',
  [ComplianceEvaluationRuleEnum.MIN_HITS_REQUIRED]: 'MIN HITS REQUIRED'
};
