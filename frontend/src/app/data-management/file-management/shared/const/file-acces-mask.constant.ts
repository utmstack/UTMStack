import {AccessMaskEnum} from '../enum/access-mask.enum';
import {FileAccessMaskCodeType} from '../types/file-access-mask-code.type';

export const ACCESS_MASK_CODES: FileAccessMaskCodeType[] = [
  {
    access: 'ReadData (or ListDirectory)',
    hex: AccessMaskEnum.READ_DATA,
    description: 'ReadData - For a file object, the right to read ' +
      'the corresponding file data. For a directory object, the right to read the' +
      ' corresponding directory data.\n' +
      'ListDirectory - For a directory, the right to list the contents of the directory.'
  },
  {
    access: 'WriteData (or AddFile)',
    hex: AccessMaskEnum.WRITE_DATA,
    description: 'WriteData - For a file object, the right to write data to the file. ' +
      'For a directory object, the right to create a file in the directory (FILE_ADD_FILE).\n' +
      'AddFile - For a directory, the right to create a file in the directory..'
  },
  {
    access: 'AppendData (or AddSubdirectory or CreatePipeInstance)',
    hex: AccessMaskEnum.APPEND_DATA,
    description: 'AppendData - For a file object, the right to append data to the file.' +
      ' (For local files, write operations will not overwrite existing data if this flag ' +
      'is specified without FILE_WRITE_DATA.) For a directory object, the right to create a ' +
      'subdirectory (FILE_ADD_SUBDIRECTORY).\n' +
      'AddSubdirectory - For a directory, the right to create a subdirectory.\n' +
      'CreatePipeInstance - For a named pipe, the right to create a pipe.'
  },
  {
    access: 'ReadEA(For registry objects, this is Enumerate sub-keys.)',
    hex: AccessMaskEnum.READ_EA,
    description: 'The right to read extended file attributes.'
  },
  {
    access: 'WriteEA',
    hex: AccessMaskEnum.WRITE_EA,
    description: 'The right to write extended file attributes.'
  },
  {
    access: 'Execute/Traverse',
    hex: AccessMaskEnum.EXECUTE_TRAVERSE,
    description: 'Execute - For a native code file, the right to execute' +
      ' the file. This access right given to scripts may cause the ' +
      'script to be executable, depending on the script interpreter.\n' +
      'Traverse - For a directory, the right to traverse the directory.' +
      ' By default, users are assigned the BYPASS_TRAVERSE_CHECKING  ' +
      'privilege, which ignores the FILE_TRAVERSE  access right. See the' +
      ' remarks in File Security and Access Rights for more information.'
  },
  {
    access: 'Delete child',
    hex: AccessMaskEnum.DELETE_CHILD,
    description: 'For a directory, the right to delete a directory and all the ' +
      'files it contains, including read-only files.'
  },
  {
    access: 'Read attributes',
    hex: AccessMaskEnum.READ_ATTRIBUTES,
    description: 'The right to read file attributes.'
  },
  {
    access: 'Write attributes',
    hex: AccessMaskEnum.WRITE_ATTRIBUTES,
    description: 'The right to write file attributes.'
  },
  {
    access: 'Delete',
    hex: AccessMaskEnum.DELETE,
    description: 'The right to delete the object.'
  },
  {
    access: 'Read control',
    hex: AccessMaskEnum.READ_CONTROL,
    description: 'The right to read the information in the object\'s security' +
      ' descriptor, not including the information' +
      ' in the system access control list (SACL).'
  },
  {
    access: 'Write DAC',
    hex: AccessMaskEnum.WRITE_AC,
    description: 'The right to modify the discretionary access control list' +
      ' (DACL) in the object\'s security descriptor.'
  },
  {
    access: 'Write OWNER',
    hex: AccessMaskEnum.WRITE_OWNER,
    description: 'The right to change the owner in the object\'s security descriptor'
  },
  {
    access: 'Synchronize',
    hex: AccessMaskEnum.SYNCHRONIZE,
    description: 'The right to use the object for synchronization. This enables a thread to ' +
      'wait until the object is in the signaled state. Some object type do not support this access right.'
  },
  {
    access: 'Access SYS_SEC',
    hex: AccessMaskEnum.ACCESS_SYS_SEC,
    description: 'The ACCESS_SYS_SEC access right controls the ability to get or set the SACL' +
      ' in an object\'s security descriptor.'
  }
];
