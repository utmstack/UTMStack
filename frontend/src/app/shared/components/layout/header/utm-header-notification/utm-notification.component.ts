import {AfterViewInit, ChangeDetectionStrategy, Component, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {NgbDropdown} from '@ng-bootstrap/ng-bootstrap';
import {Observable, of, Subject} from 'rxjs';
import {catchError, filter, map, startWith, switchMap, takeUntil, tap} from 'rxjs/operators';
import {
  NotificationDTO,
  NotificationStatus
} from '../../../../../app-management/utm-notification/models/utm-notification.model';
import {
  ComponentType,
  NotificationRefreshService
} from '../../../../../app-management/utm-notification/service/notification-refresh.service';
import {NotificationService} from '../../../../../app-management/utm-notification/service/notification.service';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-utm-notification',
  templateUrl: 'utm-notification.component.html',
  styleUrls: ['utm-notification.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmNotificationComponent implements OnInit, AfterViewInit, OnDestroy {
  notifications$: Observable<NotificationDTO[]>;
  unreadNotificationsCount$: Observable<number>;
  loadingMore = false;
  request = {page: 0, size: 5, sort: 'createdAt,DESC', status: NotificationStatus[NotificationStatus.ACTIVE] };
  @ViewChild('dropNotification') dropNotification: NgbDropdown;
  total = 0;
  destroy$ = new Subject<void>();

  constructor(private notificationService: NotificationService,
              private utmToastService: UtmToastService,
              public notificationRefreshService: NotificationRefreshService) {
  }

  ngOnInit(): void {
    this.unreadNotificationsCount$ = this.notificationRefreshService
      .refresh$
      .pipe(
        takeUntil(this.destroy$),
        startWith(() => 0),
        switchMap(() => this.notificationService.getUnreadNotificationCount()));

    this.notifications$ = this.notificationRefreshService.loadData$
      .pipe(
        takeUntil(this.destroy$),
        filter(data =>
          !!data && data.type === ComponentType.NOTIFICATION_LIST && data.value),
        switchMap((value) => {
          return this.notificationService.getAll(this.request)
            .pipe(
              tap(() => this.loadingMore = false),
              map(response =>  response.content),
              tap((data) => this.total = data.length ),
              catchError(err => {
                this.utmToastService.showError('Failed to fetch notifications',
                  'An error occurred while fetching notifications data.');
                this.loadingMore = false;
                return of([]);
              }));
        })
      );
  }

  ngAfterViewInit() {
    this.dropNotification.openChange.subscribe((isOpen: boolean) => {
      if (!isOpen) {
        this.resetRequest();
      }
    });
  }

  loadNotifications() {
    this.loadingMore = true;
    this.notificationRefreshService.loadData();
  }

  viewAllNotifications() {
    this.dropNotification.close();
  }

  markAllAsRead() {
    this.notificationService.markAllAsRead()
      .pipe(
        tap(() => this.notificationRefreshService.loadData()),
        catchError(err => {
          this.utmToastService.showError('Failed to update notifications',
            'An error occurred while updating the notification data.');
          return of([]);
        })
      ).subscribe();
  }

  updateRead(notification: NotificationDTO) {
    this.notificationService.updateNotificationReadStatus(notification.id)
      .pipe(
        tap(() => this.notificationRefreshService.loadData()),
        catchError(err => {
          this.utmToastService.showError('Failed to update notifications',
            'An error occurred while updating the notification data.');
          return of([]);
        })
      ).subscribe();
  }

  onScroll() {
    this.request.size += 5;
    this.loadNotifications();
  }

  trackByFn(index: number, notification: NotificationDTO) {
    return notification.id;
  }

  getSource(notification: NotificationDTO) {
    return notification.source.replace('_', ' ');
  }

  resetRequest() {
    this.request.size = 5;
  }

  setStatus(notification: NotificationDTO) {
    this.notificationService.updateNotificationStatus(notification.id, NotificationStatus[NotificationStatus.HIDDEN])
      .pipe(
        tap(() => this.notificationRefreshService.loadData()),
        catchError(err => {
          this.utmToastService.showError('Failed to remove notifications',
            'An error occurred while removing the notification data.');
          return of([]);
        })
      ).subscribe();
  }

  ngOnDestroy(): void {
   this.destroy$.next();
   this.destroy$.complete();
   this.notificationRefreshService.stopInterval();
  }
}
