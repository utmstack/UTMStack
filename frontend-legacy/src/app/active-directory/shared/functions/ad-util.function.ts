import {ActiveDirectoryTreeType} from '../types/active-directory-tree.type';

export function resolveType(objectClass: string[]): string {
  if (objectClass) {
    if (objectClass.includes('group')) {
      return 'GROUP';
    } else if (objectClass.includes('person') && objectClass.includes('computer')) {
      return 'COMPUTER';
    } else {
      return 'USER';
    }
  } else {
    return 'NONE';
  }
}

export function getTreeIcon(item: ActiveDirectoryTreeType): string {
  switch (item.type) {
    case 'COMPUTER':
      return 'icon-display';
    case 'GROUP':
      return 'icon-users2';
    case 'NONE':
      return 'icon-ungroup';
    default:
      return item.isAdmin ? 'icon-user-tie' : 'icon-user';
  }
}

export function resolveDistinguishedName(distinguishedName: string) {
// "CN=Invitado,CN=Users,DC=team,DC=cu"
  return distinguishedName.substring(3, distinguishedName.indexOf(','));
}

export function resolveMembersOf(members: string[]) {
  if (members) {
    const arrMembers: string[] = [];
    for (const member of members) {
      if (member.includes('CN=')) {
        const group = String(member).split('\'').join('')
          .substring(3, member.indexOf(',CN'))
          .split('=').join('');
        const index = arrMembers.findIndex(value => value === group);
        if (index === -1) {
          arrMembers.push(group);
        }
      }
    }
    return arrMembers;
  } else {
    return [];
  }
}
