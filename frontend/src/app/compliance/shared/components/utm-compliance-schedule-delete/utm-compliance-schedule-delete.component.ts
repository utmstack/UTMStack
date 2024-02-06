import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../../../shared/behaviors/nav.behavior';
import {ComplianceScheduleService} from '../../services/compliance-schedule.service';


@Component({
  selector: 'app-dashboard-delete',
  templateUrl: './utm-compliance-schedule-delete.component.html',
  styleUrls: ['./utm-compliance-schedule-delete.component.scss']
})
export class UtmComplianceScheduleDeleteComponent implements OnInit {
  @Input() complianceSchedule: any;
  @Output() complianceScheduleDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private complianceScheduleService: ComplianceScheduleService,
              private utmToastService: UtmToastService,
              private navBehavior: NavBehavior) {
  }

  ngOnInit() {
  }

  deleteDashboard() {
    this.complianceScheduleService.delete(this.complianceSchedule.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Schedule Compliance deleted successfully');
        this.activeModal.close();
        this.navBehavior.$nav.next(true);
        this.complianceScheduleDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting schedule compliance',
          'Error deleting dashboard, please check your network and try again');
      });
  }
}
