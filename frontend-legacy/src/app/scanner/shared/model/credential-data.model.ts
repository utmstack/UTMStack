export interface ICredentialData {
  id?: number;
  credential?: number;
  type?: string;
  value?: string;

}

export class CredentialDataModel implements ICredentialData {
  constructor(
    public id?: number,
    public credential?: number,
    public type?: string,
    public value?: string
  ) {
    this.id = id ? id : null;
    this.credential = credential ? credential : null;
    this.type = type ? type : null;
    this.value = value ? value : null;
  }
}
