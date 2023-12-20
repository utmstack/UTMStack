export interface IOpenvasAggregateColumnModel {
  name?: string;
  stat?: string;
  type?: string;
  column?: string;
  dataType?: string;
}

export class OpenvasAggregateColumnModel implements IOpenvasAggregateColumnModel {
  constructor(public name?: string,
              public stat?: string,
              public type?: string,
              public column?: string,
              publicdataType?: string) {
  }
}
