export interface INtvModel {
  oid?: string;
  type?: string;
  name?: string;
  family?: string;
  cvssBase?: string;
  cve?: string;
  bid?: string;
  xref?: string;
  tags?: string;
}

export class NtvModel implements INtvModel {
  constructor(public oid?: string,
              public  type?: string,
              public  name?: string,
              public family?: string,
              public cvssBase?: string,
              public cve?: string,
              public bid?: string,
              public xref?: string,
              public tags?: string) {
  }
}
