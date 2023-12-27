export interface ITargetCredentialDataModel {
  credentialId?: number;
  id?: number;
  port?: number;
  targetId?: number;
  type?: string;
}

export class TargetCredentialDataModel implements ITargetCredentialDataModel {
  constructor(
    public credentialId?: number,
    public id?: number,
    public port?: number,
    public targetId?: number,
    public type?: string
  ) {
    this.id = id ? id : null;
    this.credentialId = credentialId ? credentialId : null;
    this.type = type ? type : null;
    this.port = port ? port : null;
    this.targetId = targetId ? targetId : null;
  }
}
