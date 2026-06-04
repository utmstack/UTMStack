export class AdReportType {
  description?: string;
  emails?: string;
  id?: number;
  limit?: number;
  name?: string;
  nextExecution?: Date;
  schedule?: 0;
  type?: string;
  user?: string;
  from?: Date | string;
  to?: Date | string;
  objectsType: {
    'objectName': string,
    'objectSid': string
  }[];
}
