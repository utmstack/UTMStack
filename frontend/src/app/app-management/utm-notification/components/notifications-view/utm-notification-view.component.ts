import {AfterViewInit, ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {TimeFilterComponent} from '../../../../shared/components/utm/filters/time-filter/time-filter.component';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {TimeFilterType} from '../../../../shared/types/time-filter.type';
import {resolveRangeByTime} from '../../../../shared/util/resolve-date';
import {
  NotificationDTO,
  NotificationRequest,
  NotificationSource,
  UtmNotification
} from '../../models/utm-notification.model';
import {ComponentType, NotificationRefreshService} from '../../service/notification-refresh.service';
import {NotificationService} from '../../service/notification.service';
import {AppLogTypeEnum} from '../../../app-logs/shared/enum/app-log-type.enum';

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
  request: NotificationRequest = {
    page: 0,
    size: 25,
    sort: 'createdAt,DESC',
    from: this.eventDate.timeFrom.slice(0, 19),
    to: this.eventDate.timeTo.slice(0, 19)};
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
  sources = Object.keys(NotificationSource).filter(n => isNaN(Number(n)));
  types = Object.keys(AppLogTypeEnum).filter(n => isNaN(Number(n)));
  constructor(private notificationService: NotificationService,
              private utmToastService: UtmToastService,
              public notificationRefreshService: NotificationRefreshService) {
  }

  ngOnInit(): void {
    this.notifications$ = this.notificationRefreshService.loadData$
      .pipe(takeUntil(this.destroy$),
        filter(data =>
          !!data && data.type === ComponentType.NOTIFICATION_VIEW && data.value),
        switchMap(() => this.load()));
  }

  ngAfterViewInit() {
    this.timeFilter.timeFilterChange
      .pipe(takeUntil(this.destroy$))
      .subscribe((time) => {
        this.request = {
          ...this.request,
          from: time.timeFrom.slice(0, 19),
          to: time.timeTo.slice(0, 19)
        };
        this.notificationRefreshService.loadData(ComponentType.NOTIFICATION_VIEW);
      });
  }

  load() {
    return  this.notificationService.getAll(this.request)
      .pipe(
        tap((response) =>  this.page = {...response, number: response.number + 1}),
        map(response => response.content),
        catchError(err => {
          this.utmToastService.showError('Failed to fetch notifications',
            'An error occurred while fetching notifications data.');
          this.loadingMore = false;
          return of(null);
        }));
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

  onSelectSourceChange(source: string) {
    this.request = {
      ...this.request,
      source
    };
    this.notificationRefreshService.loadData(ComponentType.NOTIFICATION_VIEW);
  }

  onSelectTypeChange(type: string) {
    this.request = {
      ...this.request,
      type
    };
    this.notificationRefreshService.loadData(ComponentType.NOTIFICATION_VIEW);
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.notificationRefreshService.loadData(null);
  }
}
