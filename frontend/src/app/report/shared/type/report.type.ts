import {UtmModulesEnum} from '../../../app-module/shared/enum/utm-module.enum';
import {ReportTypeEnum} from '../enums/report-type.enum';

export class ReportType {
  creationDate: Date;
  creationUser: string;
  dashboardId: number;
  id: number;
  modificationDate: Date;
  modificationUser: string;
  repDescription: string;
  repName: string;
  reportSectionId: number;
  repShortName: string;
  repType: ReportTypeEnum;
  repUrl: string;
  repModule: UtmModulesEnum;
  repHttpMethod: 'POST' | 'GET';
}
