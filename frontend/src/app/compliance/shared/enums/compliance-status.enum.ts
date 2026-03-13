export enum ComplianceStatusEnum {
  COMPLIANT = 'COMPLIANT',
  NON_COMPLIANT = 'NON_COMPLIANT'
}

export enum ComplianceStatusExtendedEnum {
  COMPLIANT = 'COMPLIANT',
  NON_COMPLIANT = 'NON_COMPLIANT',
  NOT_EVALUATED = 'NOT_EVALUATED'
}

export const ComplianceStatusLabels: Record<ComplianceStatusExtendedEnum, string> = {
  [ComplianceStatusExtendedEnum.COMPLIANT]: 'Compliant',
  [ComplianceStatusExtendedEnum.NON_COMPLIANT]: 'Non Compliant',
  [ComplianceStatusExtendedEnum.NOT_EVALUATED]: 'Not Evaluated'
};

export function getComplianceStatusLabel(
  status: string | ComplianceStatusExtendedEnum
): string {
  return ComplianceStatusLabels[status]
    || ComplianceStatusLabels[ComplianceStatusExtendedEnum.NOT_EVALUATED];
}

