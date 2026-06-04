import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {ComplianceReportType} from './compliance-report.type';
import {ComplianceScheduleFilterType} from "./compliance-schedule-filter.type";

export class ComplianceScheduleType {
  id?: number;
  scheduleString: string;
  urlWithParams?: string;
  filters?: string;
  filterDef: ComplianceScheduleFilterType[];
  compliance?: ComplianceReportType;
  complianceId: number;
}
