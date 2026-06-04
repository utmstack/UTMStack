import {ActiveDirectoryUserSource} from './active-directory-user-source';

export class ActiveDirectoryUsers {
  id: string;
  sid: string;
  name: string;
  source: ActiveDirectoryUserSource;
  createdDate: string;
  modifiedDate: string;
}
