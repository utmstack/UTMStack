import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmDashboardService} from '../../../../graphic-builder/dashboard-builder/shared/services/utm-dashboard.service';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {UtmDashboardType} from '../../../../shared/chart/types/dashboard/utm-dashboard.type';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';

@Component({
  selector: 'app-utm-dashboard-select',
  templateUrl: './utm-dashboard-select.component.html',
  styleUrls: ['./utm-dashboard-select.component.scss']
})
export class UtmDashboardSelectComponent implements OnInit {
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = 10;
  dashboards: UtmDashboardType[] = [];
  searching = false;
  @Input() idDashboard: number;
  @Output() dashboardSelected = new EventEmitter<number>();
  private requestParams: any;
  private sortBy: SortEvent;
  dashboard: UtmDashboardType;

  constructor(private modalService: NgbModal,
              private dashboardService: UtmDashboardService,
              private utmToastService: UtmToastService,
              private spinner: NgxSpinnerService,
              private router: Router) {
  }

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      'name.contains': null
    };
    this.getDashboardList();
    if (this.idDashboard) {
      this.getSelectedDashboard(this.idDashboard);
    }
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

  getSelectedDashboard(id: number) {
    this.dashboardService.find(id).subscribe(response => {
      this.dashboard = response.body;
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

  onSearchDashboard($event: string) {
    this.searching = true;
    this.requestParams['name.contains'] = $event;
    this.getDashboardList();
  }

  selectDashboard(dashboard: UtmDashboardType) {
    this.dashboard = dashboard;
    this.idDashboard = dashboard.id;
    this.dashboardSelected.emit(dashboard.id);
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.dashboards = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(body: any) {
  }
}
