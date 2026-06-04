import {TaskModel} from '../../task.model';
import {NtvModel} from './ntv.model';

export interface IAssetDetailModel {
  id?: string;
  name?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  report?: {
    uuid?: string;
  };
  task?: TaskModel;
  host?: {
    value?: string;
    asset?: {
      assetId?: string;
    }
  };
  port?: string;
  nvt?: NtvModel;
  scanNvtVersion?: string;
  threat?: string;
  severity?: string;
  qod?: {
    value?: string;
    type?: string;
  };
  description?: string;
  details?: string;
}

export class AssetDetailModel implements IAssetDetailModel {
  constructor(
    public  id?: string,
    public name?: string,
    public comment?: string,
    public creationTime?: string,
    public modificationTime?: string,
    public report?: {
      uuid?: string;
    },
    public task?: TaskModel,
    public host?: {
      value?: string;
      asset?: {
        assetId?: string;
      }
    },
    public port?: string,
    public nvt?: NtvModel,
    public scanNvtVersion?: string,
    public threat?: string,
    public severity?: string,
    public qod?: {
      value?: string;
      type?: string;
    },
    public description?: string,
    public details?: string
  ) {
  }
}

