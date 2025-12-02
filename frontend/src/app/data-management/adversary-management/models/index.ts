import {UtmAlertType} from '../../../shared/types/alert/utm-alert.type';
import {Side} from '../../../shared/types/event/event';

export interface AdversaryAlerts {
  adversary: Side;
  alerts: AlertWithChildren[];
}

export interface AlertWithChildren {
  alert: UtmAlertType;
  children: UtmAlertType[];
}
