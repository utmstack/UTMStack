import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiParseLinks} from 'ng-jhipster';
import {LocalStorageService} from 'ngx-webstorage';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {TaskCreateComponent} from '../../scanner-config/task/task-create/task-create.component';
import {AssetSaveReportComponent} from '../../shared/components/asset-save-report/asset-save-report.component';
import {AssetSeverityHelpComponent} from '../../shared/components/asset-severity-help/asset-severity-help.component';
import {AssetModel} from '../../shared/model/assets/asset.model';
import {AssetsService} from '../shared/services/assets.service';

@Component({
  selector: 'app-assets-host',
  templateUrl: './assets-host.component.html',
  styleUrls: ['./assets-host.component.scss']
})
export class AssetsHostComponent implements OnInit {
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  loading = false;
  routeData: any;
  links: any;
  predicate: any;
  previousPage: any;
  reverse: any;
  assets: AssetModel[] = [];
  view = 'List';
  fields: SortByType[] = [
    {
      fieldName: 'Asset name',
      field: 'name'
    },
    {
      fieldName: 'Discovered',
      field: 'created'
    },
    {
      fieldName: 'Severity',
      field: 'severity'
    },
    {
      fieldName: 'Operative System',
      field: 'os'
    },
    {
      fieldName: 'hostname',
      field: 'Hostname'
    },
    {
      fieldName: 'IP',
      field: 'ip'
    }
  ];
  severityFilter: string;
  hostname: string;
  hostSo: string;
  private sortBy: string;
  private requestParams: any;

  constructor(private assetsService: AssetsService,
              private parseLinks: JhiParseLinks,
              private activatedRoute: ActivatedRoute,
              private modalService: NgbModal,
              private router: Router,
              private localStorage: LocalStorageService) {
  }

  ngOnInit() {
    this.activatedRoute.queryParams.subscribe(param => {
      if (param.hasOwnProperty('severity') || param.hasOwnProperty('host') || param.hasOwnProperty('hostSo')) {
        this.requestParams = {
          page: this.page - 1,
          size: this.itemsPerPage,
          sort: this.sortBy,
          type: 'host'
        };
        if (param.severity) {
          this.severityFilter = param.severity;
        }
        if (param.host) {
          this.hostname = param.host;
        }
        if (param.hostSo) {
          this.hostSo = param.hostSo;
        }
      } else {
        this.loadAssets();
      }
    });
  }

  loadAssets() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      type: 'host'
    };
    this.getAssets();
  }

  getAssets() {
    this.loading = true;
    this.assetsService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  loadPage(page: number) {
    this.page = page;
    if (page !== this.previousPage) {
      this.previousPage = page;
      this.requestParams.page = this.page - 1;
      this.getAssets();
    }
  }

  onSortBy($event: SortEvent) {
    this.sortBy = $event.column + ',' + $event.direction;
    this.requestParams.sort = this.sortBy;
    this.getAssets();
  }

  viewSeverityHelp() {
    const modal = this.modalService.open(AssetSeverityHelpComponent, {centered: true});
  }

  navigateTo(asset: AssetModel) {
    this.router.navigate(['/scanner/assets-discovery/assets-detail'],
      {queryParams: {host: asset.uuid}});
  }

  onFilterHostChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.requestParams.page = 0;
    this.getAssets();
  }

  saveReport() {
    const modal = this.modalService.open(AssetSaveReportComponent, {centered: true});
    modal.componentInstance.type = 'asset';
    modal.componentInstance.filter = this.requestParams;
  }

  private onSuccess(data, headers) {
    this.links = this.parseLinks.parse(headers.get('link'));
    this.totalItems = headers.get('X-Total-Count');
    this.assets = data;
    this.loading = false;
  }

  private onError(error) {
    this.loading = false;
    // this.alertService.error(error.error, error.message, null);
  }

  newScan() {
    const modalTask = this.modalService.open(TaskCreateComponent, {centered: true});
    modalTask.componentInstance.mode = 'simple';
  }
}
