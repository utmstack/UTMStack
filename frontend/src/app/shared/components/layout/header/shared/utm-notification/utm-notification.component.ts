import {AfterViewInit, ChangeDetectionStrategy, Component, OnInit, ViewChild} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, filter, map, startWith, switchMap, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../../alert/utm-toast.service';
import {NotificationDTO} from './models/utm-notification.model';
import {NotificationService} from './service/notification.service';
import {NotificationRefreshService} from './service/notification-refresh.service';
import {NgbDropdown} from '@ng-bootstrap/ng-bootstrap';
import {InfiniteScrollDirective, InfiniteScrollEvent} from 'ngx-infinite-scroll';

@Component({
  selector: 'app-utm-notification',
  templateUrl: 'utm-notification.component.html',
  styleUrls: ['utm-notification.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmNotificationComponent implements OnInit, AfterViewInit {
  showNotifications: boolean;
  notifications$: Observable<NotificationDTO[]>;
  unreadNotificationsCount$: Observable<number>;
  loadingMore = false;
  request = {page: 0, size: 5, sort: 'createdAt,DESC' };
  @ViewChild('dropNotification') dropNotification: NgbDropdown;

  constructor(private notificationService: NotificationService,
              private utmToastService: UtmToastService,
              public notificationRefreshService: NotificationRefreshService) {
  }

  ngOnInit(): void {
    this.unreadNotificationsCount$ = this.notificationRefreshService
      .refresh$
      .pipe(
        startWith(() => 0),
        switchMap(() => this.notificationService.getUnreadNotificationCount()));

    this.notifications$ = this.notificationRefreshService.onScrollNotification$
      .pipe(
        filter((onScroll) => !!onScroll),
        switchMap(() => {
          return this.notificationService.getAll(this.request)
            .pipe(
              tap(() => this.loadingMore = false),
              map(response =>  {
                /* this.notifications = [...this.notifications, ...response.body.content];*/
                return response.body.content;
              }),
              catchError(err => {
                this.utmToastService.showError('Failed to fetch notifications',
                  'An error occurred while fetching notifications data.');
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
  }

  viewAllNotifications() {

  }

  markAllAsRead() {

  }

  toggleNotifications() {
    this.showNotifications = true;
  }

  updateRead(notification: NotificationDTO) {
    notification.read = !notification.read;
  }

  onScroll() {
    console.log('onScroll');
    this.request.size += 5;
    this.loadNotifications();
  }

  trackByFn(index: number, notification: NotificationDTO) {
    return notification.id;
  }

  resetRequest() {
    this.request.size = 5 ;
  }
}
