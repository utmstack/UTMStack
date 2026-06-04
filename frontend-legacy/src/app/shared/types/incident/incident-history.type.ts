import {IncidentHistoryActionEnum} from '../../enums/incident/incident-history-action.enum';

export class IncidentHistoryType {
  id: number;
  incidentId: number;
  actionDate: Date;
  actionType: IncidentHistoryActionEnum;
  actionCreatedBy: string;
  action: string;
  actionDetail: string;
}
