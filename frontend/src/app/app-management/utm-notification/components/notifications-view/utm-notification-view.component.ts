import {AfterViewInit, ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {Observable, Subject} from 'rxjs';
import {filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {TimeFilterComponent} from '../../../../shared/components/utm/filters/time-filter/time-filter.component';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {resolveRangeByTime} from '../../../../shared/util/resolve-date';
import {NotificationDTO, UtmNotification} from '../../models/utm-notification.model';
import {ComponentType, NotificationRefreshService} from '../../service/notification-refresh.service';
import {NotificationService} from '../../service/notification.service';
import {AppLogTypeEnum} from "../../../app-logs/shared/enum/app-log-type.enum";

@Component({
  selector: 'app-utm-notification-view',
  templateUrl: 'utm-notification-view.component.html',
  styleUrls: ['utm-notification-view.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmNotificationViewComponent implements OnInit, AfterViewInit, OnDestroy {
  eventDate: TimeFilterType = resolveRangeByTime('week');
  notifications$: Observable<NotificationDTO[]>;
  itemsPerPage = ITEMS_PER_PAGE;
  loadingMore = false;
  request = {page: 0, size: 25, sort: 'createdAt,DESC' };
  destroy$ = new Subject<void>();
  loading: any;
  @ViewChild('timeFilter') timeFilter: TimeFilterComponent;
  page: UtmNotification = {
    last: false,
    totalElements: 0,
    totalPages: 0,
    first: true,
    size: 25,
    number: 0,
    sort: {
      sorted: false,
      unsorted: true,
      empty: true
    },
    numberOfElements: 0,
    empty: true
  };
  pageT = 0;
  constructor(private notificationService: NotificationService,
              private utmToastService: UtmToastService,
              public notificationRefreshService: NotificationRefreshService) {
  }

  ngOnInit(): void {
    this.notifications$ = this.notificationRefreshService.loadData$
      .pipe(takeUntil(this.destroy$),
        filter(data =>
          !!data && data.type === ComponentType.NOTIFICATION_VIEW && data.value),
        switchMap(() => {
          return this.load();
        }));
  }

  ngAfterViewInit() {
    this.timeFilter.timeFilterChange
      .pipe(takeUntil(this.destroy$))
      .subscribe((time) => console.log(time));
  }

  load() {
    return  this.notificationService.getAll(this.request)
      .pipe(
        tap((response) =>  this.page = {...response, number: response.number + 1}),
        map(response => response.content));
  }

  trackByFn(index: number, notification: NotificationDTO) {
    return notification.id;
  }

  getSource(notification: NotificationDTO) {
    return notification.source.replace('_', ' ');
  }

  loadPage(page: number) {
    this.page.number = page;
    this.request = {...this.request, page: this.page.number - 1};
    this.notificationRefreshService.loadData(ComponentType.NOTIFICATION_VIEW);
  }

  onItemsPerPageChange(size: number) {
    this.page.size = size;
    this.request = {...this.request, size};
    this.notificationRefreshService.loadData(ComponentType.NOTIFICATION_VIEW);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.notificationRefreshService.loadData(null);
  }
}
