import {ActiveDirectoryFieldType} from '../types/active-directory-field.type';

export const ACTIVE_DIRECTORY_USER_FIELDS: ActiveDirectoryFieldType[] = [
    {
      attrLDAPName: 'accountExpires',
      attrDisplayName: '',
      aducTab: 'Account',
      aducField: 'Account expires',
      propertySet: 'Account Restrictions',
      staticPropertyMethod: 'AccountExpirationDate',
      hiddenPerms: '',
      mo: 'AccountExpirationDate',
      syntax: 'INTEGER8',
      multiValue: false,
      minRan: null,
      maxRan: null,
      oid: '1.2.840.113556.1.4.159',
      gc: null,
      systemOnIndexed: null,
      anr: null,
      survives: null,
      copied: true,
      nonrepl: null,
      construct: null,
    }
  ]
;
