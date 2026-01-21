import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {ComplianceStrategyEnum} from '../enums/compliance-strategy.enum';
import {UtmComplianceQueryConfigType} from './compliance-query-config.type';
import {ComplianceStandardSectionType} from './compliance-standard-section.type';


export class ComplianceReportConfigType {
  id?: number;
  section?: ComplianceStandardSectionType;
  standardSectionId?: number;
  configReportName?: string;
  configSolution?: string;
  configRemediation?: string;
  strategy?: ComplianceStrategyEnum;
  queriesConfigs?: UtmComplianceQueryConfigType[];

  columns?: UtmFieldType[];
  dashboardId?: number;
  associatedDashboard?: UtmDashboardType;
  selected?: boolean;
  visualization?: any;
  status?: string;
  dashboard?: UtmDashboardVisualizationType[];
}
