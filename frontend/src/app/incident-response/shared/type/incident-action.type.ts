import {IncidentResponseActionsEnum} from '../../../shared/enums/incident-response/incident-response-actions.enum';

export class IncidentActionType {
  actionCommand?: string;
  actionDescription?: string;
  actionEditable?: boolean;
  actionParams?: string;
  actionType?: IncidentResponseActionsEnum;
  createdDate?: Date;
  createdUser?: string;
  id?: number;
  modifiedDate?: Date;
  modifiedUser?: string;
}
