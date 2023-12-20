import {OpenvasSubgroupModel} from './openvas-subgroup.model';

export interface IOpenvasGroupModel {
  value?: string;
  subgroups?: OpenvasSubgroupModel[];
  count?: string;
  cCount?: string;
  stats?: {
    column?: string,
    min?: string,
    max?: string,
    mean?: string,
    sum?: string,
    cSum?: string
  }[];
  text?: string;
}

export class OpenvasGroupModel implements IOpenvasGroupModel {
  constructor(public value?: string,
              public subgroups?: OpenvasSubgroupModel[],
              public count?: string,
              public cCount?: string,
              public stats?: {
                column?: string,
                min?: string,
                max?: string,
                mean?: string,
                sum?: string,
                cSum?: string
              }[],
              public text?: string) {
  }
}

