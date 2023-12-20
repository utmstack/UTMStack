import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbDateAdapter, NgbDateNativeAdapter} from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import {setMaxDateToday} from '../../../../util/datepicker-util';

@Component({
  selector: 'app-date-range',
  templateUrl: './date-range.component.html',
  styleUrls: ['./date-range.component.scss'],
  providers: [{provide: NgbDateAdapter, useClass: NgbDateNativeAdapter}]
})
export class DateRangeComponent implements OnInit {
  @Output() dateRange = new EventEmitter<{ timeFrom: string, timeTo: string }>();
  toDate: Date;
  fromDate: Date;
  maxDate = setMaxDateToday();

  constructor(public activeModal: NgbActiveModal) {
  }

  get today(): Date {
    return new Date();
  }

  ngOnInit(): void {
  }

  compareDate(): boolean {
    if (this.fromDate !== undefined && this.toDate !== undefined) {
      if (this.fromDate !== null && this.toDate !== null) {
        return this.fromDate.getTime() - this.toDate.getTime() > 0;
      } else {
        return false;
      }
    }
  }

  emitDate() {
    this.dateRange.emit({
      timeFrom: moment.parseZone(this.fromDate).utc().format(),
      timeTo: moment.parseZone(this.toDate).utc().format()
    });
    this.activeModal.close();
  }
}
