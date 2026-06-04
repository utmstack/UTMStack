import {OpenvasColumnInfoModel} from './openvas-column-info.model';
import {OpenvasGroupModel} from './openvas-group.model';

export interface IOpenvasOptionModel {
  dataType?: string;
  dataColumn?: string;
  groupColumn?: string;
  textColumn?: string;
  groups?: OpenvasGroupModel[];
  overall?: string;
  subgroups?: {
    value?: string
  };
  columnInfo?: OpenvasColumnInfoModel;
}

export class OpenvasOptionModel implements IOpenvasOptionModel {
  constructor(public dataType?: string,
              public dataColumn?: string,
              public groupColumn?: string,
              public  textColumn?: string,
              public  groups?: OpenvasGroupModel[],
              public overall?: string,
              public subgroups?: {
                value?: string
              },
              public  columnInfo?: OpenvasColumnInfoModel) {
  }
}
