import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {AlertHistoryType} from '../../../../../shared/types/alert/alert-history.type';
import {AlertUpdateHistoryBehavior} from '../../behavior/alert-update-history.behavior';
import {AlertHistoryActionEnum} from '../../enums/alert-history-action.enum';
import {getID} from '../../util/alert-util-function';
import {AlertHistoryService} from '../alert-history/alert-history.service';

@Component({
  selector: 'app-alert-view-last-change',
  templateUrl: './alert-view-last-change.component.html',
  styleUrls: ['./alert-view-last-change.component.scss']
})
export class AlertViewLastChangeComponent implements OnInit, OnDestroy {
  @Input() action: AlertHistoryActionEnum;
  @Input() alert: any;
  @Output() emptyValue = new EventEmitter<boolean>(false);
  lastChange: AlertHistoryType;
  loadingChange = true;
  changes: string[];
  destroy$: Subject<void> = new Subject<void>();

  constructor(private alertHistoryService: AlertHistoryService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior) {
  }

  ngOnInit(): void {
    this.alertUpdateHistoryBehavior.$refreshHistory
      .pipe(
        takeUntil(this.destroy$),
        filter(value => !!value)
      )
      .subscribe(value => {
      if (value && !!this.action) {
        this.loadingChange = true;
        this.getAlertHistory();
      }
    });
    this.getAlertHistory();
  }

  getAlertHistory() {
    const req = {
      'alertId.equals': getID(this.alert),
      'logAction.equals': this.action,
      page: 0,
      size: 1,
      sort: 'logDate,desc'
    };
    this.alertHistoryService.query(req).subscribe(logs => {
      if (logs.body.length !== 0) {
        this.changes = logs && logs.body && logs.body[0].logMessage.split('[utm-logs-separator]');
      } else {
        this.emptyValue.emit(true);
      }

      this.loadingChange = false;
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.alertUpdateHistoryBehavior.$refreshHistory.next(null);
  }

}
