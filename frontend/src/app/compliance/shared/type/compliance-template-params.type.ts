import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';

export class ComplianceTemplateType {
  complianceDescription?: string;
  complianceReportColumns?: UtmFieldType[];
  complianceReportReqBody?: ElasticFilterType[];
  requestParamFilters?: ComplianceTemplateParams[];
  complianceReportResourceUrl?: string;
  complianceRule?: string;
  complianceSolution?: string;
  complianceTpl?: string;
  id?: number;
}

export class ComplianceTemplateParams {
  param?: string;
  type?: string;
  required?: boolean;
  value?: any;
}
