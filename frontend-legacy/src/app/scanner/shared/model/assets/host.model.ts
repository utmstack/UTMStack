import {HostDetailModel} from './host-detail.model';

export interface IHostModel {
  severity?: {
    value: string
  };
  detail?: HostDetailModel[];
}

export class HostModel implements IHostModel {
  constructor(public severity?: {
                value: string
              },
              public detail?: HostDetailModel[]) {
  }
}
