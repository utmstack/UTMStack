export interface IOpenvasSubgroupModel {
  value?: string;
  count?: string;
  cCount?: string;
  stats?: string;
}

export class OpenvasSubgroupModel implements IOpenvasSubgroupModel {
  constructor(
    public value?: string,
    public count?: string,
    public cCount?: string,
    public stats?: string
  ) {
  }
}
