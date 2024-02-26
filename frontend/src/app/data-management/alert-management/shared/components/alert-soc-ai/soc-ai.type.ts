export class SocAiType {
  severity: string;
  category: string;
  alertName: string;
  status: string;
  activityId: string;
  classification: string;
  reasoning: string[];
  nextSteps: SocAiNextStep[];
}

export class SocAiNextStep {
  step: number;
  action: string;
  details: string;
}

export enum IndexSocAiStatus {
  Completed = 'Completed',
  Processing = 'Processing',
  Error = 'Error',
  NotFound = 'NotFound'
}
