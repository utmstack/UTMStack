export class ActiveDirectoryTreeType {
  name: string;
  id?: string;
  type?: string;
  isAdmin?: boolean;
  objectSid?: string;
  parentId?: string;
  children?: ActiveDirectoryTreeType[];
}
