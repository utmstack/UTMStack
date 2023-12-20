import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {ComplianceRequestTypeEnum} from '../enums/compliance-request-type.enum';
import {ComplianceTypeEnum} from '../enums/compliance-type.enum';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';
import {ComplianceTemplateParams} from './compliance-template-params.type';

export class ComplianceReportType {
  columns?: UtmFieldType[];
  configReportDataOrigin?: number;
  configReportEditable?: boolean;
  configReportExportCsvUrl?: string;
  configReportFilterByTime?: boolean;
  configReportPageable?: true;
  configReportRequestType?: ComplianceRequestTypeEnum;
  configReportResourceUrl?: string;
  configSolution?: string;
  id?: number;
  requestBodyFilters?: ElasticFilterType[];
  requestParamFilters?: ComplianceTemplateParams[];
  section?: ComplianceStandardSectionType;
  standardSectionId?: number;
  dashboardId?: number;
  configType?: ComplianceTypeEnum;
  configUrl?: string;
  associatedDashboard?: UtmDashboardType;
}
