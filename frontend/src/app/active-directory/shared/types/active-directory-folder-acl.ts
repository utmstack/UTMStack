import {ActiveDirectoryFolderAclAccess} from './active-directory-folder-acl-access';

export class ActiveDirectoryFolderAcl {
  owner?: string;
  folder?: string;
  access?: ActiveDirectoryFolderAclAccess[];
}
