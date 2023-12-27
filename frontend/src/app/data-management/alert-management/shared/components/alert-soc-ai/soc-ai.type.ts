export class SocAiType {
  severity: string;
  category: string;
  alertName: string;
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
