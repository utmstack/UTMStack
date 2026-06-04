import {IncidentOriginTypeEnum} from '../../../shared/enums/incident-response/incident-origin-type.enum';
import {IncidentResponseStatusEnum} from '../../../shared/enums/incident-response/incident-response-status.enum';
import {IncidentActionType} from './incident-action.type';

export class IncidentJobType {
  id?: number;
  action?: IncidentActionType;
  actionId?: number;
  params?: string;
  agent?: string;
  status?: IncidentResponseStatusEnum;
  jobResult?: string;
  originId?: number | string;
  originType?: IncidentOriginTypeEnum;
  createdDate?: Date;
  createdUser?: string;
  modifiedDate?: Date;
  modifiedUser?: string;
}
