import {ScannerSeverityEnum} from '../enums/scanner-severity.enum';

export function resolveColor(severity: string) {
  const sev = Number.parseFloat(severity);
  if (sev === 0) {
    return '#4e342e';
  } else if (sev < 0) {
    return '#343a40';
  } else if (sev >= 0.1 && sev <= 3.9) {
    return '#00838F';
  } else if (sev >= 4 && sev <= 6.9) {
    return '#D84315';
  } else if (sev >= 7) {
    return '#C62828';
  }
}

export function resolveSeverityLabel(severity: string) {
  const sev = Number.parseFloat(severity);
  if (sev === 0) {
    return ScannerSeverityEnum.LOG;
  } else if (sev < 0) {
    return ScannerSeverityEnum.UNKNOWN;
  } else if (sev >= 0.1 && sev <= 3.9) {
    return ScannerSeverityEnum.LOW;
  } else if (sev >= 4 && sev <= 6.9) {
    return ScannerSeverityEnum.MEDIUM;
  } else if (sev >= 7) {
    return ScannerSeverityEnum.HIGH;
  }
}

export function resolveSeverityClass(severity: string) {
  const sev = Number.parseFloat(severity);
  if (sev === 0) {
    return 'badge-flat border-brown text-brown-800';
  } else if (sev < 0) {
    return 'badge-flat border-dark text-dark-800';
  } else if (sev >= 0.1 && sev <= 3.9) {
    return 'badge-flat border-info-800 text-info-800';
  } else if (sev >= 4 && sev <= 6.9) {
    return 'badge-flat border-warning-800 text-warning-800';
  } else if (sev >= 7) {
    return 'badge-flat border-danger-800 text-danger-800';
  }
}
