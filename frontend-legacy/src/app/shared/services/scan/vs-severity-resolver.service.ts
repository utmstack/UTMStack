import {Injectable} from '@angular/core';
import {VsSeverityEnum} from '../../../vulnerability-scanner/shared/enums/vs-severity.enum';

@Injectable({
  providedIn: 'root'
})
export class VsSeverityResolverService {

  constructor() {
  }

  public resolveColor(severity: string | null | number): string {
    if (severity !== null) {
      if (typeof severity === 'string') {
        severity = Number.parseFloat(severity);
      }
      if (severity === 0) {
        return '#777777';
      } else if (severity < 0) {
        return '#607D8B';
      } else if (severity >= 0.1 && severity <= 3.9) {
        return '#42A5F5';
      } else if (severity >= 4 && severity <= 6.9) {
        return '#FF9800';
      } else if (severity >= 7) {
        return '#EF5350';
      }
    } else {
      return '#7E57C2';
    }
  }

  public resolveColorByName(severity: string): string {
    switch (severity) {
      case VsSeverityEnum.HIGH:
        return '#EF5350';
      case VsSeverityEnum.MEDIUM:
        return '#FF9800';
      case VsSeverityEnum.LOW:
        return '#42A5F5';
      case VsSeverityEnum.LOG:
        return '#607D8B';
      case VsSeverityEnum.UNKNOWN:
        return '#777777';
    }
  }

  public resolveSeverityLabel(severity: string | null | number): string {
    if (severity !== null) {
      if (typeof severity === 'string') {
        severity = Number.parseFloat(severity);
      }
      if (severity === 0) {
        return VsSeverityEnum.LOG;
      } else if (severity < 0) {
        return VsSeverityEnum.UNKNOWN;
      } else if (severity >= 0.1 && severity <= 3.9) {
        return VsSeverityEnum.LOW;
      } else if (severity >= 4 && severity <= 6.9) {
        return VsSeverityEnum.MEDIUM;
      } else if (severity >= 7) {
        return VsSeverityEnum.HIGH;
      }
    } else {
      return VsSeverityEnum.UNKNOWN;
    }
  }

  public resolveSeverityClass(severity: string | number): string {
    let sev = severity;
    if (typeof severity === 'string') {
      sev = Number.parseFloat(severity);
    }
    if (sev === 0) {
      return 'badge-flat border-slate-300 text-slate-300';
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

  public resolveSeverityFixed(severity: string | null): string {
    return severity !== null ? severity.replace('.0', '') : '';
  }

}
