import {ScannerSeverityEnum} from '../enums/scanner-severity.enum';

export const STATUS_TASK = ['Delete Requested', 'Done', 'New', 'Requested',
  'Running', 'Stop Requested', 'Stopped', 'Internal Error'];
export const SEVERITY_VALUE = [
  ScannerSeverityEnum.UNKNOWN,
  ScannerSeverityEnum.LOG,
  ScannerSeverityEnum.LOW,
  ScannerSeverityEnum.MEDIUM,
  ScannerSeverityEnum.HIGH,
  ScannerSeverityEnum.ALL
];
