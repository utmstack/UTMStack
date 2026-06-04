import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {LocalStorageService} from 'ngx-webstorage';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {AssetSeverityHelpComponent} from '../../shared/components/asset-severity-help/asset-severity-help.component';
// tslint:disable-next-line:max-line-length
import {ScannerExportVulnerabilitiesComponent} from '../../shared/components/scanner-export-vulnerabilities/scanner-export-vulnerabilities.component';
import {AssetDetailModel} from '../../shared/model/assets/detail/asset-detail.model';
import {TaskModel} from '../../shared/model/task.model';
import {PieSeverityClassDef} from '../shared/chart/pie-severity-class.def';
import {WordCloudDef} from '../shared/chart/word-cloud.def';
import {AssetHostDetailDashboardService} from '../shared/services/asset-host-detail-dashboard.service';
import {AssetHostDetailService} from '../shared/services/asset-host-detail.service';
// @ts-ignore
require('echarts-wordcloud');

@Component({
  selector: 'app-assets-host-detail',
  templateUrl: './task-result.component.html',
  styleUrls: ['./taks-result.component.scss']
})
export class TaskResultComponent implements OnInit {
  details: AssetDetailModel[];
  task: TaskModel;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  loading = false;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  fields: SortByType[] = [
    {
      fieldName: 'Vulnerability',
      field: 'vulnerability'
    },
    {
      fieldName: 'Severity',
      field: 'severity'
    },
    {
      fieldName: 'Last modification',
      field: 'modified'
    },
    {
      fieldName: 'Severity',
      field: 'severity'
    },
    {
      fieldName: 'Min QOD',
      field: 'min_qod'
    },
    {
      fieldName: 'Location',
      field: 'location'
    }
  ];
  pieOption: any;
  wordCloudOption: any;
  barOption: any;
  loadingPieOption = true;
  loadingWordCloudOption = true;
  loadingBarOption = true;
  private sortBy: string;
  private requestParams: any;
  viewDetail: AssetDetailModel;

  constructor(private routeParams: ActivatedRoute,
              private assetHostDetailService: AssetHostDetailService,
              private localStorage: LocalStorageService,
              private modalService: NgbModal,
              private assetHostDetailDashboardService: AssetHostDetailDashboardService,
              private pieSeverityClassDef: PieSeverityClassDef,
              private wordCloudDef: WordCloudDef,
              private router: Router) {
  }

  ngOnInit() {
    this.sortBy = 'severity,desc';
    this.task = this.localStorage.retrieve('TASK_RESULT');
    this.routeParams.queryParams.subscribe(param => {
      this.loadTaskResult();
      this.loadCharts();
    });

  }

  loadCharts() {
    this.getResultsBySeverityClass();
    this.getResultsWordCloud();
  }

  getResultsBySeverityClass() {
    this.loadingPieOption = true;
    this.loadingBarOption = true;
    this.assetHostDetailDashboardService.resultsBySeverityClass({
      'reportId.equals': this.task.lastReport.report.uuid,
      'minQod.greaterThan': 70,
      'taskId.equals': this.task.uuid
    }).subscribe(value => {
      this.loadingPieOption = false;
      this.loadingBarOption = false;
      if (value.body[0].groups) {
        this.pieOption = this.pieSeverityClassDef.buildChartBySeverityClass(value.body[0]);
        this.barOption = this.pieSeverityClassDef.buildChartResultsByCVSS(value.body[0]);
      } else {
        this.pieOption = null;
        this.barOption = null;
      }
    });
  }

  getResultsWordCloud() {
    this.loadingWordCloudOption = true;
    this.assetHostDetailDashboardService.resultsVulnerabilityWordCloud({
      'reportId.equals': this.task.lastReport.report.uuid,
      'minQod.greaterThan': 70,
      'taskId.equals': this.task.uuid
    }).subscribe(value => {
      this.loadingWordCloudOption = false;
      if (value.body[0].groups) {
        this.wordCloudOption = this.wordCloudDef.buildChartWordCloud(value.body[0]);
      } else {
        this.wordCloudOption = null;
      }
    });
  }

  loadTaskResult() {
    this.loading = true;
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      details: true,
      'minQod.greaterThan': 70,
      'taskId.equals': this.task.uuid,
      'reportId.equals': this.task.lastReport.report.uuid
    };
    this.getTaskResult();
  }

  getTaskResult() {
    this.assetHostDetailService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.requestParams.sort = this.sortBy;
    this.loadTaskResult();
  }

  viewSeverityHelp() {
    const modal = this.modalService.open(AssetSeverityHelpComponent, {centered: true});
  }

  loadPage(page: number) {
    this.page = page;
    this.loadTaskResult();
  }

  saveReport() {
    const modal = this.modalService.open(ScannerExportVulnerabilitiesComponent, {centered: true});
    modal.componentInstance.reportId = this.task.lastReport.report.uuid;
  }

  viewAssetDetail(asset: string) {
    this.router.navigate(['/scanner/assets-discovery/assets-detail'],
      {queryParams: {ip: asset}});
  }

  private onSuccess(data, headers) {
    // this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.details = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
