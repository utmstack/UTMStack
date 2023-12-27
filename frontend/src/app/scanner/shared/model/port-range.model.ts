export interface IPortRange {
  comment?: string;
  end?: string;
  exclude?: string;
  id?: string;
  portListId?: string;
  start?: string;
  type?: string;
  uuid?: string;
}

export class PortRangeModel implements IPortRange {
  constructor(
    public comment?: string,
    public end?: string,
    public exclude?: string,
    public  id?: string,
    public portListId?: string,
    public  start?: string,
    public type?: string,
    public uuid?: string
  ) {
    this.id = id ? id : null;
    this.comment = comment ? comment : null;
    this.exclude = exclude ? exclude : null;
    this.portListId = portListId ? portListId : null;
    this.start = start ? start : null;
    this.type = type ? type : null;
    this.uuid = uuid ? uuid : null;
  }
}
