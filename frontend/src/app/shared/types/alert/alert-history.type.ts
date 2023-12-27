import {AlertHistoryActionEnum} from '../../../data-management/alert-management/shared/enums/alert-history-action.enum';

export class AlertHistoryType {
  alertId: string;
  id: number;
  logAction: AlertHistoryActionEnum;
  logDate: Date;
  logNewValue: string;
  logOldValue: string;
  logMessage: string;
  logUser: string;
}
