import {IncidentResponseStatusEnum} from '../../../shared/enums/incident-response/incident-response-status.enum';

export const IR_STATUS: { label: string, status: IncidentResponseStatusEnum }[] =
  [
    {label: 'Pending', status: IncidentResponseStatusEnum.PENDING},
    {label: 'Running', status: IncidentResponseStatusEnum.RUNNING},
    {label: 'Executed', status: IncidentResponseStatusEnum.EXECUTED},
    {label: 'Error', status: IncidentResponseStatusEnum.ERROR}
  ];
