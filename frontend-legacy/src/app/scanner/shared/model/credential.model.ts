export interface ICredential {
  id?: number;
  uuid?: string;
  owner?: string;
  name?: string;
  comment?: string;
  creationTime?: number;
  modificationTime?: number;
  type?: string;
  allowInsecure?: string;
  authAlgorithm?: string;
  fullType?: string;
  inUse?: string;
  login?: string;
  privacy?: {
    algorithm?: string,
    password?: string
  };
  targets?: {
    uuid: string,
    name: string
  }[];
  writable?: string;
}

export class CredentialModel implements ICredential {
  constructor(
    public id?: number,
    public uuid?: string,
    public owner?: string,
    public name?: string,
    public comment?: string,
    public  creationTime?: number,
    public  modificationTime?: number,
    public  type?: string,
    public allowInsecure?: string,
    public authAlgorithm?: string,
    public  fullType?: string,
    public  inUse?: string,
    public  login?: string,
    public  privacy?: {
      algorithm: string,
      password: string
    },
    public targets?: {
      uuid: string,
      name: string
    }[],
    public writable?: string,
  ) {
    this.id = id ? id : null;
    this.uuid = uuid ? uuid : null;
    this.owner = owner ? owner : null;
    this.name = name ? name : null;
    this.comment = comment ? comment : null;
    this.creationTime = creationTime ? creationTime : null;
    this.modificationTime = modificationTime ? modificationTime : null;
    this.type = type ? type : null;
    this.allowInsecure = allowInsecure ? allowInsecure : null;
    this.authAlgorithm = authAlgorithm ? authAlgorithm : null;
    this.fullType = fullType ? fullType : null;
    this.inUse = inUse ? inUse : null;
    this.login = login ? login : null;
    this.privacy = privacy ? privacy : null;
    this.targets = targets ? targets : null;
    this.writable = writable ? writable : null;
  }
}
