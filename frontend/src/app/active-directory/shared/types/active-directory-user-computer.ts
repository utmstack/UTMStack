import {ActiveDirectoryFolderAcl} from './active-directory-folder-acl';
import {ActiveDirectoryLocalGroupType} from './active-directory-local-group.type';

export class ActiveDirectoryUserComputer {
  objectSid?: string;
  name?: string;
  localGroups?: ActiveDirectoryLocalGroupType[];
  localFolders?: ActiveDirectoryFolderAcl[];
}

