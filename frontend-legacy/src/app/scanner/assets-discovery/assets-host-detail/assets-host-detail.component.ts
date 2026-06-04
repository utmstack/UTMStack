import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {LocalStorageService} from 'ngx-webstorage';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {ChartTypeEnum} from '../../../shared/enums/chart-type.enum';
import {SortByType} from '../../../shared/types/sort-by.type';
import {AssetSaveReportComponent} from '../../shared/components/asset-save-report/asset-save-report.component';
import {AssetSeverityHelpComponent} from '../../shared/components/asset-severity-help/asset-severity-help.component';
import {AssetModel} from '../../shared/model/assets/asset.model';
import {AssetDetailModel} from '../../shared/model/assets/detail/asset-detail.model';
import {PieSeverityClassDef} from '../shared/chart/pie-severity-class.def';
import {WordCloudDef} from '../shared/chart/word-cloud.def';
import {AssetHostDetailDashboardService} from '../shared/services/asset-host-detail-dashboard.service';
import {AssetHostDetailService} from '../shared/services/asset-host-detail.service';
import {AssetsService} from '../shared/services/assets.service';
// @ts-ignore
require('echarts-wordcloud');

@Component({
  selector: 'app-assets-host-detail',
  templateUrl: './assets-host-detail.component.html',
  styleUrls: ['./assets-host-detail.component.scss']
})
export class AssetsHostDetailComponent implements OnInit {
  details: AssetDetailModel[];
  asset: AssetModel;
  chartTypeEnum = ChartTypeEnum;
  totalItems: any;
  page = 1;
  itemsPerPage = 5;
  loading = true;
  routeData: any;
  links: any;
  viewDetail: AssetDetailModel;
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
  loadingPieOption = true;
  loadingWordCloudOption = true;
  private sortBy: string;
  private requestParams: any;

  constructor(private routeParams: ActivatedRoute,
              private assetHostDetailService: AssetHostDetailService,
              private localStorage: LocalStorageService,
              private modalService: NgbModal,
              private assetHostDetailDashboardService: AssetHostDetailDashboardService,
              private pieSeverityClassDef: PieSeverityClassDef,
              private wordCloudDef: WordCloudDef,
              private assetService: AssetsService) {
  }

  ngOnInit() {
    this.sortBy = 'severity,desc';
    this.routeParams.queryParams.subscribe(param => {
      this.assetService.query({'hostIp.equals': param.ip, type: 'host'}).subscribe(value => {
        this.asset = value.body[0];
        this.loadAssetDetail();
        this.loadCharts();
      });
    });

  }

  loadCharts() {
    this.getResultsBySeverityClass();
    this.getResultsWordCloud();
  }

  getResultsBySeverityClass() {
    this.loadingPieOption = true;
    this.assetHostDetailDashboardService.resultsBySeverityClass(
      {'host.equals': this.asset.name, 'minQod.greaterThan': 69}).subscribe(value => {
      this.loadingPieOption = false;
      if (value.body[0].groups) {
        this.pieOption = this.pieSeverityClassDef.buildChartBySeverityClass(value.body[0]);
      } else {
        this.pieOption = null;
      }
    });
  }

  getResultsWordCloud() {
    this.loadingWordCloudOption = true;
    this.assetHostDetailDashboardService.resultsVulnerabilityWordCloud({
      'host.equals': this.asset.name, 'minQod.greaterThan': 69,
    }).subscribe(value => {
      this.loadingWordCloudOption = false;
      if (value.body[0].groups) {
        this.wordCloudOption = this.wordCloudDef.buildChartWordCloud(value.body[0]);
      } else {
        this.wordCloudOption = null;
      }
    });
  }

  loadAssetDetail() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      details: true,
      'minQod.greaterThan': 69,
      'host.equals': this.asset.name
    };
    this.getAssetDetail();
  }

  getAssetDetail() {
    this.loading = true;
    this.assetHostDetailService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.requestParams.sort = this.sortBy;
    this.loadAssetDetail();
  }

  viewSeverityHelp() {
    const modal = this.modalService.open(AssetSeverityHelpComponent, {centered: true});
  }

  loadPage(page: number) {
    this.page = page;
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.loadAssetDetail();
    }
  }

  saveReport() {
    const modal = this.modalService.open(AssetSaveReportComponent, {centered: true});
    modal.componentInstance.type = 'detail';
    modal.componentInstance.filter = {
      host: this.asset.name,
      'minQod.greaterThan': 69
    };
  }

  private onSuccess(data, headers) {
    // this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.details = data;
    this.loading = false;
  }

  // viewDetail(detail: AssetDetailModel) {
  //   const modal = this.modalService.open(VulnerabilityDetailComponent, {centered: true, size: 'lg'});
  //   modal.componentInstance.detail = detail;
  // }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
