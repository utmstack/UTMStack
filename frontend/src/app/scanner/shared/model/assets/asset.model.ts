import {AssetIdentifierModel} from './asset-identifier.model';
import {HostModel} from './host.model';

export interface IAssetModel {
  uuid?: string;
  name?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  writable?: string;
  inUse?: string;
  identifiers?: AssetIdentifierModel[];
  type?: string;
  host?: HostModel;
}

export class AssetModel implements IAssetModel {

  constructor(public uuid?: string,
              public name?: string,
              public comment?: string,
              public creationTime?: string,
              public  modificationTime?: string,
              public  writable?: string,
              public inUse?: string,
              public identifiers?: AssetIdentifierModel[],
              public type?: string,
              public host?: HostModel) {
  }
}
