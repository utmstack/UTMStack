import {PortRangeModel} from './port-range.model';

export interface IPort {
  id?: string;
  uuid?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  writable?: string;
  inUse?: string;
  portCount?: {
    all?: string,
    tcp?: string,
    udp?: string
  };
  portRanges?: PortRangeModel[];
  targets?: {
    uuid?: string,
    name?: string
  }[];
  name?: string;
}

export class PortModel implements IPort {
  constructor(
    public id?: any,
    public name?: string,
    public uuid?: string,
    public comment?: string,
    public  creationTime?: string,
    public  modificationTime?: string,
    public writable?: string,
    public inUse?: string,
    public portCount?: {
      all?: string,
      tcp?: string,
      udp?: string
    },
    public  portRanges?: PortRangeModel[],
    public  targets?: {
      uuid?: string,
      name?: string
    }[],
  ) {
    this.id = id ? id : null;
    this.comment = comment ? comment : null;
    this.creationTime = creationTime ? creationTime : null;
    this.modificationTime = modificationTime ? modificationTime : null;
    this.name = name ? name : null;
    this.portRanges = portRanges ? portRanges : null;
    this.uuid = uuid ? uuid : null;
    this.writable = writable ? writable : null;
    this.inUse = inUse ? inUse : null;
    this.portCount = portCount ? portCount : null;
    this.targets = targets ? targets : null;
  }
}
