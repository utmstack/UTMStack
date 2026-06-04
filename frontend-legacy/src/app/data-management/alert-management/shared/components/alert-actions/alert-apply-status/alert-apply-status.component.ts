import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
// tslint:disable-next-line:import-spacing max-line-length
import {
  CLOSED,
  CLOSED_ICON,
  IGNORED,
  IGNORED_ICON,
  OPEN_ICON,
  REVIEW,
  REVIEW_ICON
} from '../../../../../../shared/constants/alert/alert-status.constant';
import {EventDataTypeEnum} from '../../../enums/event-data-type.enum';
import {AlertManagementService} from '../../../services/alert-management.service';
import {AlertCompleteComponent} from '../../alert-complete/alert-complete.component';

@Component({
  selector: 'app-alert-apply-status',
  templateUrl: './alert-apply-status.component.html',
  styleUrls: ['./alert-apply-status.component.scss']
})
export class AlertApplyStatusComponent implements OnInit {
  @Input() statusFilter: number;
  @Input() alertsIDs: string[];
  @Input() dataType: EventDataTypeEnum;
  eventDataTypeEnum = EventDataTypeEnum;
  @Output() statusChange = new EventEmitter<number>();
  status: number;
  review = REVIEW;
  ignored = IGNORED;
  closed = CLOSED;
  openIcon = OPEN_ICON;
  reviewIcon = REVIEW_ICON;
  ignoredIcon = IGNORED_ICON;
  closedIcon = CLOSED_ICON;

  constructor(private modalService: NgbModal,
              private alertManagementService: AlertManagementService) {
  }

  ngOnInit() {
  }

  applyStatus() {
    if (this.status === IGNORED || this.status === CLOSED) {
      const modalRef = this.modalService.open(AlertCompleteComponent, {centered: true});
      modalRef.componentInstance.alertsIDs = this.alertsIDs;
      modalRef.componentInstance.canCreateRule = false;
      modalRef.componentInstance.status = this.status;
      modalRef.componentInstance.statusChange.subscribe((statusChange) => {
        this.statusChange.emit(this.status);
      });
    } else {
      this.markMultiple(this.status).then(finis => {
        this.statusChange.emit(this.status);
      });
    }
  }

  markMultiple(status): Promise<any> {
    return new Promise((resolve, reject) => {
      this.alertManagementService.updateAlertStatus(this.alertsIDs, status).subscribe(al => {
        resolve('finish');
      });
    });
  }
}
