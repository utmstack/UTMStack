export interface IScannerConfigModel {
  uuid: string;
  name: string;
}

export class ScannerConfigModel implements IScannerConfigModel {
  uuid: string;
  name: string;

  constructor(uuid: string, name: string) {
    this.uuid = uuid;
    this.name = name;
  }
}
