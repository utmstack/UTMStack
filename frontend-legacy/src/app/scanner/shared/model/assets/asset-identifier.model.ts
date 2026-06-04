export interface IAssetIdentifierModel {
  uuid?: string;
  name?: string;
  value?: string;
  creationTime?: string;
  modificationTime?: string;
  source?: {
    uuid?: string,
    type?: string,
    data?: string,
    deleted?: string,
    name?: string,
  };
}

export class AssetIdentifierModel implements IAssetIdentifierModel {
  constructor(
    public uuid?: string,
    public name?: string,
    public value?: string,
    public creationTime?: string,
    public modificationTime?: string,
    public source?: {
      uuid?: string,
      type?: string,
      data?: string,
      deleted?: string,
      name?: string,
    }) {
  }
}

