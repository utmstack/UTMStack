import {TaskModel} from './task.model';

export interface ITarget {
  id?: string;
  uuid?: string;
  name?: string;
  comment?: string;
  creationTime?: string;
  modificationTime?: string;
  writable?: string;
  inUse?: string;
  hosts?: string;
  excludeHosts?: string;
  maxHosts?: string;
  sshCredential?: {
    uuid?: string,
    name?: string,
    port?: string,
    trash?: string
  };
  smbCredential?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  esxiCredential?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  snmpCredential?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  portRange?: null;
  portList?: {
    uuid?: string,
    name?: string,
    trash?: string
  };
  aliveTests?: string;
  reverseLookupOnly?: string;
  reverseLookupUnify?: string;
  tasks?: TaskModel[];
}


export class TargetModel implements ITarget {
  constructor(
    public id?: string,
    public uuid?: string,
    public name?: string,
    public comment?: string,
    public creationTime?: string,
    public modificationTime?: string,
    public writable?: string,
    public inUse?: string,
    public hosts?: string,
    public excludeHosts?: string,
    public maxHosts?: string,
    public sshCredential?: {
      uuid?: string,
      name?: string,
      port?: string,
      trash?: string
    },
    public smbCredential?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public  esxiCredential?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public snmpCredential?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public portRange?: null,
    public portList?: {
      uuid?: string,
      name?: string,
      trash?: string
    },
    public aliveTests?: string,
    public reverseLookupOnly?: string,
    public reverseLookupUnify?: string,
    public tasks?: TaskModel[],
  ) {
    this.id = id ? id : null;
    this.name = name ? name : null;
    this.uuid = uuid ? uuid : null;
    this.comment = comment ? comment : null;
    this.creationTime = creationTime ? creationTime : null;
    this.modificationTime = modificationTime ? modificationTime : null;
    this.writable = writable ? writable : null;
    this.inUse = inUse ? inUse : null;
    this.hosts = hosts ? hosts : null;
    this.excludeHosts = excludeHosts ? excludeHosts : null;
    this.maxHosts = maxHosts ? maxHosts : null;
    this.sshCredential = sshCredential ? sshCredential : null;
    this.smbCredential = smbCredential ? smbCredential : null;
    this.esxiCredential = esxiCredential ? esxiCredential : null;
    this.snmpCredential = snmpCredential ? snmpCredential : null;
    this.portRange = portRange ? portRange : null;
    this.portList = portList ? portList : null;
    this.aliveTests = aliveTests ? aliveTests : null;
    this.reverseLookupOnly = reverseLookupOnly ? reverseLookupOnly : null;
    this.reverseLookupUnify = reverseLookupUnify ? reverseLookupUnify : null;
    this.tasks = tasks ? tasks : null;
  }
}
