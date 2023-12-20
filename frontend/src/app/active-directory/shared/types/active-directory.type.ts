import {ActiveDirectoryFolderAcl} from './active-directory-folder-acl';
import {ActiveDirectoryIpType} from './active-directory-ip.type';
import {ActiveDirectoryLocalGroupType} from './active-directory-local-group.type';
import {ActiveDirectoryLocalUserType} from './active-directory-local-user.type';
import {ActiveDirectoryUserComputer} from './active-directory-user-computer';
import {AdUserAclType} from './ad-user-acl.type';

export class ActiveDirectoryType {
  accountExpires?: string;
  admin?: boolean;
  adminCount?: boolean;
  badPasswordTime?: string;
  cn?: string;
  disabled?: boolean;
  displayName?: string;
  distinguishedName?: string;
  id?: string;
  ips?: ActiveDirectoryIpType[];
  lastLogon?: string;
  lastLogonTimestamp?: string;
  localFolders?: ActiveDirectoryFolderAcl[];
  localGroups?: ActiveDirectoryLocalGroupType[];
  localUsers?: ActiveDirectoryLocalUserType[];
  locked?: boolean;
  member?: string[];
  memberOf?: string[];
  memberOfClean?: string [];
  objectCategory?: string;
  objectClass?: string [];
  objectSid?: string;
  realLastLogon?: string;
  sAMAccountName?: string;
  userACLs?: AdUserAclType[];
  userAccountControl?: number;
  whenCreated?: string;
  computerInformation?: ActiveDirectoryUserComputer[];
  type?: 'user' | 'computer' | 'group' | 'object';
  operatingSystem?: string;
}
