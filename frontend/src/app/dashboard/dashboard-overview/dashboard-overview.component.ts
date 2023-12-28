import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
// tslint:disable-next-line:max-line-length
import {UtmModulesEnum} from '../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../app-module/shared/services/utm-modules.service';
import {UtmModuleType} from '../../app-module/shared/type/utm-module.type';
import {AccountService} from '../../core/auth/account.service';
import {MenuBehavior} from '../../shared/behaviors/menu.behavior';
import {
  ALERT_GLOBAL_FIELD,
  ALERT_NAME_FIELD,
  ALERT_SEVERITY_FIELD_LABEL,
  ALERT_TIMESTAMP_FIELD
} from '../../shared/constants/alert/alert-field.constant';
import {HIGH_TEXT, LOW_TEXT, MEDIUM_TEXT} from '../../shared/constants/alert/alert-severity.constant';
import {ALERT_ROUTE, LOG_ROUTE} from '../../shared/constants/app-routes.constant';
import {TIME_DASHBOARD_REFRESH} from '../../shared/constants/time-refresh.const';
import {IndexPatternSystemEnumID, IndexPatternSystemEnumName} from '../../shared/enums/index-pattern-system.enum';
import {UtmRunModeService} from '../../shared/services/active-modules/utm-run-mode.service';
import {OverviewAlertDashboardService} from '../../shared/services/charts-overview/overview-alert-dashboard.service';
import {UtmOpenModuleModalService} from '../../shared/services/config/utm-open-module-modal.service';
import {ElasticSearchIndexService} from '../../shared/services/elasticsearch/elasticsearch-index.service';
import {IndexPatternService} from '../../shared/services/elasticsearch/index-pattern.service';
import {LocalFieldService} from '../../shared/services/elasticsearch/local-field.service';
import {ChartSerieValueType} from '../../shared/types/chart-reponse/chart-serie-value.type';

@Component({
  selector: 'app-dashboard-overview',
  templateUrl: './dashboard-overview.component.html',
  styleUrls: ['./dashboard-overview.component.scss']
})
export class DashboardOverviewComponent implements OnInit {
  pdfExport = false;
  refreshTime = TIME_DASHBOARD_REFRESH;
  dailyAlert: ChartSerieValueType[] = [];
  loadingChartDailyAlert = true;
  refreshInterval = 30000;
  ALERT_ROUTE = ALERT_ROUTE;
  LOG_ANALYZER_ROUTE = LOG_ROUTE;
  // params
  paramsAlertSeverity = {};
  paramAlertSeverityCLick = ALERT_SEVERITY_FIELD_LABEL;

  tableTopAlertsParams = {};
  tableTopAlertsParamsClick = ALERT_NAME_FIELD;

  paramsEventByType = {
    // 'global.type.keyword': 'logx',
    '@timestamp': null,
    'dataType.keyword': null,
    patternId: IndexPatternSystemEnumID.LOG,
    indexPattern: IndexPatternSystemEnumName.LOG
  };
  paramEventByTypeCLick = 'dataType.keyword';

  paramsTopEvent = {
    // 'global.type.keyword': 'logx',
    '@timestamp': null,
    'dataType.keyword': 'wineventlog',
    'logx.wineventlog.event_name.keyword': null,
    patternId: IndexPatternSystemEnumID.LOG,
    indexPattern: IndexPatternSystemEnumName.LOG
  };
  paramEvenTopCLick = 'logx.wineventlog.event_name.keyword';

  adActive: boolean;
  vulActive: boolean;
  alertSeverityColorMap: { value: string, color: string }[] = [
    {color: '#42A5F5', value: LOW_TEXT},
    {color: '#FF9800', value: MEDIUM_TEXT},
    {color: '#EF5350', value: HIGH_TEXT}]
  ;


  constructor(private overviewAlertDashboardService: OverviewAlertDashboardService,
              private moduleService: UtmModulesService,
              private utmRunModeService: UtmRunModeService,
              private menuBehavior: MenuBehavior,
              private utmOpenModuleModalService: UtmOpenModuleModalService,
              private localFieldService: LocalFieldService,
              private indexPatternService: IndexPatternService,
              private indexPatternFieldService: ElasticSearchIndexService,
              private accountService: AccountService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.menuBehavior.$menu.next(false);
    window.addEventListener('beforeprint', (event) => {
      this.pdfExport = true;
    });
    window.addEventListener('afterprint', (event) => {
      this.pdfExport = false;
    });
    /**
     * SET ALERT PARAMS TO NAVIGATE
     */
    this.paramsAlertSeverity[ALERT_GLOBAL_FIELD] = 'ALERT';
    this.paramsAlertSeverity[ALERT_TIMESTAMP_FIELD] = null;
    this.paramsAlertSeverity[ALERT_SEVERITY_FIELD_LABEL] = null;

    this.tableTopAlertsParams[ALERT_GLOBAL_FIELD] = 'ALERT';
    this.tableTopAlertsParams[ALERT_TIMESTAMP_FIELD] = null;
    this.tableTopAlertsParams[ALERT_NAME_FIELD] = null;
    /**
     * END
     */

    this.getDailyAlert();

    /**
     * Show activate modules modal on constructor
     * Why?? Because home route may change, but always load in this module
     */
    // this.utmOpenModuleModalService.find(1).subscribe(response => {
    //   if (!response.body.moduleModalShown) {
    //     const modal = this.modalService.open(AppModuleActivateModalComponent, {centered: true});
    //   }
    // });

    setTimeout(() => {
      this.synchronizeFields();
    }, 100000);
  }

  getActiveModuleStatus(modules: UtmModuleType[], moduleEnum: UtmModulesEnum): boolean {
    const index = modules.findIndex(value => value.moduleName === moduleEnum);
    return modules[index].moduleActive;
  }

  getDailyAlert() {
    this.loadingChartDailyAlert = true;
    this.overviewAlertDashboardService.getCardAlertTodayWeek().subscribe(response => {
      this.dailyAlert = response.body;
      this.loadingChartDailyAlert = false;
    });
  }

  exportToPdf() {
    this.pdfExport = true;
    // captureScreen('utmDashboardAlert').then((finish) => {
    //   this.pdfExport = false;
    // });
    setTimeout(() => {
      window.print();
    }, 1000);
  }

  chartEvent($event: string) {
  }

  /**
   * Sync field in local storage from index service
   */
  synchronizeFields() {
    this.accountService.identity(true).then(value => {
      if (value) {
        this.indexPatternService.query({page: 0, size: 2000}).subscribe(responsePatterns => {
          for (const pattern of responsePatterns.body) {
            this.indexPatternFieldService.getElasticIndexField({indexPattern: pattern.pattern})
              .subscribe(responseFields => {
                this.localFieldService.setPatternStoredFields(pattern.pattern, responseFields.body);
              });
          }
        });
      }
    });
  }
}
