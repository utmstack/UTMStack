import {Component, ElementRef, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {MenuBehavior} from '../../../../../../behaviors/menu.behavior';
import {NewAlertBehavior} from '../../../../../../behaviors/new-alert.behavior';
import {stringParamToQueryParams} from '../../../../../../util/query-params-to-filter.util';
import {AlertOpenStatusService} from '../../../../../../webflux/alert-open-status.service';
import {Subject} from "rxjs";
import {takeUntil} from "rxjs/operators";

@Component({
  selector: 'app-utm-notification-alert',
  templateUrl: './utm-notification-alert.component.html',
  styleUrls: ['./utm-notification-alert.component.css']
})
export class UtmNotificationAlertComponent implements OnInit, OnDestroy {
  @ViewChild('iconBell') iconBell: ElementRef;
  totalAlerts = 10;
  prevTotal = 5;
  timeoutAlert: number;
  intervalAlert: number;
  destroy$: Subject<boolean> = new Subject<boolean>();

  constructor(private toast: UtmToastService,
              private newAlertBehavior: NewAlertBehavior,
              private alertOpenStatusService: AlertOpenStatusService,
              private spinner: NgxSpinnerService,
              private menuBehavior: MenuBehavior,
              public router: Router) {
  }

  ngOnInit() {
    this.timeoutAlert = setTimeout(() => {
      this.newAlertBehavior.$alertChange.next(0);
      this.getTotalAlertsOpen();
      this.intervalAlert = setInterval(() => {
        this.getTotalAlertsOpen();
      }, 30000);
    }, 30000);

  }

  ngOnDestroy() {
    this.destroy$.next(true);
    clearInterval(this.timeoutAlert);
    clearInterval(this.intervalAlert);
  }

  setCount(count) {
    // const el = this.iconBell.nativeElement;
    // el.setAttribute('data-count', count > 20 ? '20+' : count);
  }

  navigateToAlerts() {
    this.spinner.show('loadingSpinner');
    stringParamToQueryParams('alertType=ALERT').then(queryParams => {
      // Reload component on navigate
      this.router.routeReuseStrategy.shouldReuseRoute = () => false;
      this.router.onSameUrlNavigation = 'reload';
      this.router.navigate(['/data/alert/view'],
        {queryParams}).then(() => {
        this.menuBehavior.$menu.next(false);
        this.spinner.hide('loadingSpinner');
      });
    });
  }

  private getTotalAlertsOpen() {
    this.alertOpenStatusService.getOpenAlert()
      .pipe(takeUntil(this.destroy$))
      .subscribe(response => {
      this.totalAlerts = response.body;
      this.newAlertBehavior.$alertChange.next(1);
      if (this.prevTotal < this.totalAlerts) {
        if (!this.router.url.includes('/dashboard/export')) {
          this.toast.showWarning('There are ' + (this.totalAlerts - this.prevTotal) +
            ' pending alerts to manage', 'New alerts');
        }
        this.prevTotal = this.totalAlerts;
        this.newAlertBehavior.$alertChange.next(this.prevTotal);
        this.setCount(this.totalAlerts);
      }
    });
  }

}
