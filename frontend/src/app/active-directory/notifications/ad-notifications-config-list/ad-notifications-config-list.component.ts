import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ITEMS_PER_PAGE} from '../../../shared/constants/pagination.constants';
import {SortByType} from '../../../shared/types/sort-by.type';
import {AdNotificationsConfigCreateComponent} from '../ad-notifications-config-create/ad-notifications-config-create.component';

@Component({
  selector: 'app-ad-notifications-config-list',
  templateUrl: './ad-notifications-config-list.component.html',
  styleUrls: ['./ad-notifications-config-list.component.scss']
})
export class AdNotificationsConfigListComponent implements OnInit {
  notifications: any[] = [];
  fields: SortByType[] = [
    {
      fieldName: 'Name',
      field: 'name'
    },
    {
      fieldName: 'Last modification',
      field: 'modificationTime'
    }
  ];
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = ITEMS_PER_PAGE;
  search: any;
  sortBy: string;

  constructor(private modalService: NgbModal,
              private router: Router) {
  }

  ngOnInit() {
    this.getNotificationList();
  }


  onSortBy($event) {
  }

  editNotification(track) {
    // this.router.navigate(['/creator/builder/chart-builder'],
    //   {
    //     queryParams: {
    //       type: vis.eventType,
    //       pattern: vis.pattern.pattern,
    //       patternId: vis.pattern.id,
    //       chart: vis.chartType,
    //       mode: 'edit',
    //       visualizationId: vis.id
    //     }
    //   }).then(() => {
    // });
  }


  deleteNotifications(vis) {
    // const modal = this.modalService.open(VisualizationDeleteComponent, {centered: true});
    // modal.componentInstance.visualization = vis;
    // modal.componentInstance.visualizationDeleted.subscribe(deleted => {
    //   // this.loadTasks();
    //   this.getNotificationList();
    // });
  }

  loadPage(page: any) {
  }


  getNotificationList() {
    // const req = {
    //   page: this.page - 1,
    //   size: this.itemsPerPage,
    //   'pattern.contains': this.search,
    //   sort: this.sortBy
    // };
    // this.visualizationService.query(req).subscribe(
    //   (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
    //   (res: HttpResponse<any>) => this.onError(res.body)
    // );
  }

  newConfig() {
    const modalNotifyConfig = this.modalService.open(AdNotificationsConfigCreateComponent, {centered: true});
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.notifications = data;
    this.loading = false;
  }

  private onError(error) {
    // this.alertService.error(error.error, error.message, null);
  }
}
