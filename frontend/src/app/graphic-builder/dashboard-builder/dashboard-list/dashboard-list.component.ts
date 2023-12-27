import {HttpResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../shared/chart/types/dashboard/utm-dashboard.type';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortEvent} from '../../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../../shared/types/sort-by.type';
import {copyToClipboard} from '../../../shared/util/copy-to-clipboard-util';
import {cleanVisualizationData} from '../../shared/util/visualization/visualization-cleaner.util';
import {DashboardDeleteComponent} from '../dashboard-delete/dashboard-delete.component';
import {DashboardImportComponent} from '../dashboard-import/dashboard-import.component';
import {UtmDashboardVisualizationService} from '../shared/services/utm-dashboard-visualization.service';
import {UtmDashboardService} from '../shared/services/utm-dashboard.service';
import {buildDashboardUrl} from '../shared/util/get-menu-url';

@Component({
  selector: 'app-dashboard-list',
  templateUrl: './dashboard-list.component.html',
  styleUrls: ['./dashboard-list.component.scss']
})
export class DashboardListComponent implements OnInit {
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modifiedDate'
    }
  ];
  loading = false;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  dashboards: UtmDashboardType[] = [];
  searching = false;
  private requestParams: any;
  private sortBy: SortEvent;
  checkbox: boolean;
  selected: number[] = [];
  exporting = false;

  constructor(private modalService: NgbModal,
              private dashboardService: UtmDashboardService,
              private utmToastService: UtmToastService,
              private spinner: NgxSpinnerService,
              private dashboardVisualizationService: UtmDashboardVisualizationService,
              private router: Router) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
    };
    this.getDashboardList();
  }

  onSortBy($event) {
  }

  editDashboard(dash: UtmDashboardType) {
    this.router.navigate(['/creator/dashboard/builder'], {
      queryParams: {
        mode: 'edit',
        dashboardId: dash.id,
        dashboardName: dash.name.toLowerCase().replace(' ', '_')
      }
    });
  }


  deleteDashboard(dashboard: UtmDashboardType) {
    const modal = this.modalService.open(DashboardDeleteComponent, {centered: true});
    modal.componentInstance.dashboard = dashboard;
    modal.componentInstance.dashboardDeleted.subscribe(deleted => {
      this.getDashboardList();
    });
  }

  loadPage(page: any) {
    this.requestParams.page = page - 1;
    this.getDashboardList();
  }

  getDashboardList() {
    this.dashboardService.query(this.requestParams).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  onFilterDashboardChange($event: any) {
    Object.keys($event).forEach(key => {
      if ($event[key] !== '' && $event[key] !== null) {
        this.requestParams[key] = $event[key];
      } else {
        this.requestParams[key] = undefined;
      }
    });
    this.getDashboardList();
  }

  onSort($event: SortEvent) {
    this.requestParams.sort = $event.column + ',' + $event.direction;
    this.getDashboardList();
  }

  toClipboard(dashboard: UtmDashboardType) {
    copyToClipboard(this.getDashboardUrl(dashboard)).then(() => {
      this.utmToastService.showInfo('In clipboard', 'Dashboard ' + dashboard.name + ' url copied to clipboard');
    });
  }

  getDashboardUrl(dashboard: UtmDashboardType): string {
    return buildDashboardUrl(dashboard);
  }


  onSearchDashboard($event: string) {
    this.searching = true;

    this.requestParams.page = 0;
    this.requestParams['name.contains'] = $event;
    this.getDashboardList();
  }

  viewDashboard(dashboard: UtmDashboardType) {
    this.getDashboardUrl(dashboard);
    this.spinner.show('loadingSpinner');
    this.router.navigate([this.getDashboardUrl(dashboard)]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  toggleCheck() {
    this.checkbox = !this.checkbox;
    if (this.checkbox) {
      this.dashboards.forEach(value => {
        this.addToSelected(value);
      });
    } else {
      this.selected = [];
    }
  }

  private onError(body: any) {
  }

  addToSelected(dashboard: UtmDashboardType) {
    const index = this.selected.findIndex(value => value === dashboard.id);
    if (index === -1) {
      this.selected.push(dashboard.id);
    } else {
      this.selected.splice(index, 1);
    }
  }

  isSelected(dashboard: UtmDashboardType) {
    return this.selected.findIndex(value => value === dashboard.id) !== -1;
  }

  importDashboard() {
    const importModal = this.modalService.open(DashboardImportComponent, {centered: true});
    importModal.componentInstance.dashboardImported.subscribe(() => {
      this.getDashboardList();
    });
  }

  exportDashboard() {
    this.exporting = true;
    this.dashboardVisualizationService.query({'idDashboard.in': this.selected, size: 1000000}).subscribe(response => {
      this.cleanData(response.body).then(visualizations => {
        const uri = 'data:application/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(visualizations));
        const exportFileDefaultName = 'UTMDashboard-' + moment(new Date()).format('YYYY-MM-DD') + '.json';
        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', uri);
        linkElement.setAttribute('download', exportFileDefaultName);
        linkElement.click();
        this.exporting = false;
      });
    });
  }

  cleanData(visualizations: UtmDashboardVisualizationType[]): Promise<UtmDashboardVisualizationType[]> {
    return new Promise<UtmDashboardVisualizationType[]>(resolve => {
      visualizations.forEach(dashVis => {
        if (typeof dashVis.visualization.chartConfig === 'string') {
          dashVis.visualization.chartConfig = JSON.parse(dashVis.visualization.chartConfig);
        }
        cleanVisualizationData(dashVis.visualization).then((visualization) => {
          visualization.chartConfig = JSON.stringify(visualization.chartConfig);
          dashVis.visualization = visualization;
        });
      });
      resolve(visualizations);
    });
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.dashboards = data;
    this.searching = false;
    this.loading = false;
  }
}
