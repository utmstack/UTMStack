export interface IHostDetailModel {
  name?: string;
  value?: string;
  source?: {
    id?: string;
    type?: string;
  };
}

export class HostDetailModel implements IHostDetailModel {
  name?: string;
  value?: string;
  source?: {
    id?: string;
    type?: string;
  };

  constructor(name: string, value: string, source: { id?: string; type?: string }) {
  }
}

