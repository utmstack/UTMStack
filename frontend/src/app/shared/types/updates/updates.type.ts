export class UpdatesType {
  version: string;
  tag: string;
  notes: string;
  id?: number;
}

export class VersionInfo {
  'display-ribbon-on-profiles': string;
  git: any;
  build: {
    artifact: string,
    name: string,
    time: Date,
    version: string,
    group: string
  };
  activeProfiles: string[];
}


