export interface IScannerModel {
  id?: string;
  uuid?: string;
  owner?: string;
  name?: string;
  comment?: string;
  host?: string;
  port?: string;
  type?: string;
  caPub?: string;
  credentialId?: string;
  creationTime?: string;
  modificationTime?: string;
}

export class ScannerModel implements IScannerModel {
  constructor(
    public  id?: string,
    public uuid?: string,
    public   owner?: string,
    public  name?: string,
    public  comment?: string,
    public   host?: string,
    public   port?: string,
    public   type?: string,
    public  caPub?: string,
    public  credentialId?: string,
    public  creationTime?: string,
    public  modificationTime?: string
  ) {
  }

}
