import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {FileFieldEnum} from '../enum/file-field.enum';


export const FILE_FIELDS: UtmFieldType[] = [
  {
    label: 'Event date',
    field: FileFieldEnum.FILE_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Event name',
    field: FileFieldEnum.FILE_EVENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'User',
    field: FileFieldEnum.FILE_SUBJECT_USER_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object type',
    field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object name',
    field: FileFieldEnum.FILE_OBJECT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Process name',
    field: FileFieldEnum.FILE_PROCESS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Action',
    field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Old ACL',
    field: FileFieldEnum.FILE_OLD_SDDL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'New ACL',
    field: FileFieldEnum.FILE_NEW_SDDL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Accesss',
    field: FileFieldEnum.FILE_ACCESS_LIST_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Handler ID',
    field: FileFieldEnum.FILE_HANDLE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object server',
    field: FileFieldEnum.FILE_OBJECT_SERVER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process ID',
    field: FileFieldEnum.FILE_PROCESS_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Resource attribute',
    field: FileFieldEnum.FILE_RESOURCE_ATT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject logon ID',
    field: FileFieldEnum.FILE_SUBJECT_LOGON_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject user ID',
    field: FileFieldEnum.FILE_SUBJECT_USER_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host architecture',
    field: FileFieldEnum.FILE_HOST_ARCHITECTURE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host ID',
    field: FileFieldEnum.FILE_HOST_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Hostname',
    field: FileFieldEnum.FILE_HOST_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host OS',
    field: FileFieldEnum.FILE_HOST_OS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Build',
    field: FileFieldEnum.FILE_HOTS_OS_BUILD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Family',
    field: FileFieldEnum.FILE_HOST_OS_FAMILY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Platform',
    field: FileFieldEnum.FILE_HOST_OS_PLATFORM_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Version',
    field: FileFieldEnum.FILE_HOST_OS_VERSION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Keywords',
    field: FileFieldEnum.FILE_KEYWORD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OP code',
    field: FileFieldEnum.FILE_OPCODE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Provider GUID',
    field: FileFieldEnum.FILE_PROVIDER_GUID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Share name',
    field: FileFieldEnum.FILE_SHARE_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Share local path',
    field: FileFieldEnum.FILE_SHARE_PATH_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Share IP port',
    field: FileFieldEnum.FILE_SHARE_PATH_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  }
];
export const FILE_COMMON_FIELDS: UtmFieldType[] = [
  {
    label: 'Event date',
    field: FileFieldEnum.FILE_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Object name',
    field: FileFieldEnum.FILE_OBJECT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'User',
    field: FileFieldEnum.FILE_SUBJECT_USER_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object type',
    field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },

  {
    label: 'Process name',
    field: FileFieldEnum.FILE_PROCESS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Action',
    field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Event name',
    field: FileFieldEnum.FILE_EVENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Accesss',
    field: FileFieldEnum.FILE_ACCESS_LIST_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Handler ID',
    field: FileFieldEnum.FILE_HANDLE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object server',
    field: FileFieldEnum.FILE_OBJECT_SERVER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process ID',
    field: FileFieldEnum.FILE_PROCESS_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Resource attribute',
    field: FileFieldEnum.FILE_RESOURCE_ATT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject logon ID',
    field: FileFieldEnum.FILE_SUBJECT_LOGON_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject user ID',
    field: FileFieldEnum.FILE_SUBJECT_USER_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host architecture',
    field: FileFieldEnum.FILE_HOST_ARCHITECTURE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host ID',
    field: FileFieldEnum.FILE_HOST_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Hostname',
    field: FileFieldEnum.FILE_HOST_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host OS',
    field: FileFieldEnum.FILE_HOST_OS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Build',
    field: FileFieldEnum.FILE_HOTS_OS_BUILD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Family',
    field: FileFieldEnum.FILE_HOST_OS_FAMILY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Platform',
    field: FileFieldEnum.FILE_HOST_OS_PLATFORM_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Version',
    field: FileFieldEnum.FILE_HOST_OS_VERSION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Keywords',
    field: FileFieldEnum.FILE_KEYWORD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OP code',
    field: FileFieldEnum.FILE_OPCODE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Provider GUID',
    field: FileFieldEnum.FILE_PROVIDER_GUID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
];
export const FILE_SHARED_FIELDS: UtmFieldType[] = [
  {
    label: 'Event date',
    field: FileFieldEnum.FILE_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Hostname',
    field: FileFieldEnum.FILE_HOST_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Share name',
    field: FileFieldEnum.FILE_SHARE_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Share local path',
    field: FileFieldEnum.FILE_SHARE_PATH_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Share IP port',
    field: FileFieldEnum.FILE_SHARE_PATH_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'User',
    field: FileFieldEnum.FILE_SUBJECT_USER_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Action',
    field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object type',
    field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object name',
    field: FileFieldEnum.FILE_OBJECT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process name',
    field: FileFieldEnum.FILE_PROCESS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Event name',
    field: FileFieldEnum.FILE_EVENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Accesss',
    field: FileFieldEnum.FILE_ACCESS_LIST_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Handler ID',
    field: FileFieldEnum.FILE_HANDLE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object server',
    field: FileFieldEnum.FILE_OBJECT_SERVER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process ID',
    field: FileFieldEnum.FILE_PROCESS_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Resource attribute',
    field: FileFieldEnum.FILE_RESOURCE_ATT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject logon ID',
    field: FileFieldEnum.FILE_SUBJECT_LOGON_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject user ID',
    field: FileFieldEnum.FILE_SUBJECT_USER_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host architecture',
    field: FileFieldEnum.FILE_HOST_ARCHITECTURE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host ID',
    field: FileFieldEnum.FILE_HOST_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },

  {
    label: 'Host OS',
    field: FileFieldEnum.FILE_HOST_OS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Build',
    field: FileFieldEnum.FILE_HOTS_OS_BUILD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Family',
    field: FileFieldEnum.FILE_HOST_OS_FAMILY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Platform',
    field: FileFieldEnum.FILE_HOST_OS_PLATFORM_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Version',
    field: FileFieldEnum.FILE_HOST_OS_VERSION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Keywords',
    field: FileFieldEnum.FILE_KEYWORD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OP code',
    field: FileFieldEnum.FILE_OPCODE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Provider GUID',
    field: FileFieldEnum.FILE_PROVIDER_GUID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
];
export const FILE_PERMISSION_FIELDS: UtmFieldType[] = [
  {
    label: 'Event date',
    field: FileFieldEnum.FILE_TIMESTAMP_FIELD,
    type: ElasticDataTypesEnum.DATE,
    visible: true,
  },
  {
    label: 'Hostname',
    field: FileFieldEnum.FILE_HOST_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'User',
    field: FileFieldEnum.FILE_SUBJECT_USER_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object type',
    field: FileFieldEnum.FILE_OBJECT_TYPE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Object name',
    field: FileFieldEnum.FILE_OBJECT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Action',
    field: FileFieldEnum.FILE_ACCESS_MASK_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Old ACL',
    field: FileFieldEnum.FILE_OLD_SDDL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'New ACL',
    field: FileFieldEnum.FILE_NEW_SDDL_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Process name',
    field: FileFieldEnum.FILE_PROCESS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Event name',
    field: FileFieldEnum.FILE_EVENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Accesss',
    field: FileFieldEnum.FILE_ACCESS_LIST_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Handler ID',
    field: FileFieldEnum.FILE_HANDLE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object server',
    field: FileFieldEnum.FILE_OBJECT_SERVER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process ID',
    field: FileFieldEnum.FILE_PROCESS_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Resource attribute',
    field: FileFieldEnum.FILE_RESOURCE_ATT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject logon ID',
    field: FileFieldEnum.FILE_SUBJECT_LOGON_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject user ID',
    field: FileFieldEnum.FILE_SUBJECT_USER_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host architecture',
    field: FileFieldEnum.FILE_HOST_ARCHITECTURE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host ID',
    field: FileFieldEnum.FILE_HOST_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },

  {
    label: 'Host OS',
    field: FileFieldEnum.FILE_HOST_OS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Build',
    field: FileFieldEnum.FILE_HOTS_OS_BUILD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Family',
    field: FileFieldEnum.FILE_HOST_OS_FAMILY_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Platform',
    field: FileFieldEnum.FILE_HOST_OS_PLATFORM_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OS Version',
    field: FileFieldEnum.FILE_HOST_OS_VERSION_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Keywords',
    field: FileFieldEnum.FILE_KEYWORD_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'OP code',
    field: FileFieldEnum.FILE_OPCODE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Provider GUID',
    field: FileFieldEnum.FILE_PROVIDER_GUID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },

  {
    label: 'Share IP port',
    field: FileFieldEnum.FILE_SHARE_PATH_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  }
];


export const FILE_FILTER_FIELDS: UtmFieldType[] = [

  {
    label: 'User',
    field: FileFieldEnum.FILE_SUBJECT_USER_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Event name',
    field: FileFieldEnum.FILE_EVENT_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process name',
    field: FileFieldEnum.FILE_PROCESS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Handler ID',
    field: FileFieldEnum.FILE_HANDLE_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Object server',
    field: FileFieldEnum.FILE_OBJECT_SERVER_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Process ID',
    field: FileFieldEnum.FILE_PROCESS_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },

  {
    label: 'Resource attribute',
    field: FileFieldEnum.FILE_RESOURCE_ATT_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject domain name',
    field: FileFieldEnum.FILE_SUBJECT_DOMAIN_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject logon ID',
    field: FileFieldEnum.FILE_SUBJECT_LOGON_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Subject user ID',
    field: FileFieldEnum.FILE_SUBJECT_USER_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host architecture',
    field: FileFieldEnum.FILE_HOST_ARCHITECTURE_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Host ID',
    field: FileFieldEnum.FILE_HOST_ID_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  },
  {
    label: 'Hostname',
    field: FileFieldEnum.FILE_HOST_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: true,
  },
  {
    label: 'Host OS',
    field: FileFieldEnum.FILE_HOST_OS_NAME_FIELD,
    type: ElasticDataTypesEnum.STRING,
    visible: false,
  }
];

export const PERMISSION_FILE_EVENT_ID_NUMBER = 4670;
export const ALL_FILE_EVENT_ID_NUMBER = [4670, 5145, 5140, 5142, 5143, 5144, 5168, 4656, 4658, 4659, 4660, 4663, 4664, 4985, 5051];
export const SHARE_FILE_EVENT_ID_NUMBER = [5145, 5140, 5142, 5143, 5144, 5168];
export const DELETED_FILE_EVENT_ID_NUMBER = [4663];
export const CREATED_FILE_EVENT_ID_NUMBER = 4663;
export const FILE_OBJECT_TYPE_VALUE = ['File', 'Folder'];

// NETWORK SHARE FIELDS
