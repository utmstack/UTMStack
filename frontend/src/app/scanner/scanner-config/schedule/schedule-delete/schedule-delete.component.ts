import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';

import {ScheduleModel} from '../../../shared/model/schedule.model';
import {ScheduleService} from '../shared/services/schedule.service';

@Component({
  selector: 'app-schedule-delete',
  templateUrl: './schedule-delete.component.html',
  styleUrls: ['./schedule-delete.component.scss']
})
export class ScheduleDeleteComponent implements OnInit {
  @Input() schedule: ScheduleModel;
  @Output() scheduleDeleted = new EventEmitter<string>();

  constructor(
    public activeModal: NgbActiveModal,
    private scheduleService: ScheduleService,
    private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteSchedule() {
    this.scheduleService.delete(this.schedule.uuid)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Schedule deleted successfully');
        this.activeModal.close();
        this.scheduleDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting schedule',
          'Error deleting Schedule, please check your network and try again');
      });
  }

}
