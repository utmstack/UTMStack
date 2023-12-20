import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
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
export class AlertViewLastChangeComponent implements OnInit {
  @Input() action: AlertHistoryActionEnum;
  @Input() alert: any;
  @Output() emptyValue = new EventEmitter<boolean>(false);
  lastChange: AlertHistoryType;
  loadingChange = true;

  constructor(private alertHistoryService: AlertHistoryService,
              private alertUpdateHistoryBehavior: AlertUpdateHistoryBehavior) {
  }

  ngOnInit(): void {
    this.alertUpdateHistoryBehavior.$refreshHistory.subscribe(value => {
      if (value) {
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
      this.alertUpdateHistoryBehavior.$refreshHistory.next(null);
      if (logs.body.length !== 0) {
        this.lastChange = logs.body[0];
      } else {
        this.emptyValue.emit(true);
      }

      this.loadingChange = false;
    });
  }

}
