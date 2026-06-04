import {AfterContentChecked, ChangeDetectorRef, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbCalendar} from '@ng-bootstrap/ng-bootstrap';
import {ScheduleService} from '../../../../scanner/scanner-config/schedule/shared/services/schedule.service';
import {ScheduleModel} from '../../../../scanner/shared/model/schedule.model';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {timePeriodValidator} from '../../../../shared/util/schedule-time-period-validator';

@Component({
  selector: 'app-active-directory-schedule-create',
  templateUrl: './ad-schedule-create.component.html',
  styleUrls: ['./ad-schedule-create.component.scss']
})
export class AdScheduleCreateComponent implements OnInit, AfterContentChecked {
  @Input() schedule: ScheduleModel;
  @Output() scheduleCreated = new EventEmitter<string>();
  date: { year: number; month: number };
  timezones = [];
  duration = ['hour', 'day', 'week', 'month'];
  formSchedule: FormGroup;
  creating = false;

  constructor(public activeModal: NgbActiveModal,
              private calendar: NgbCalendar,
              private fb: FormBuilder,
              private scheduleService: ScheduleService,
              private utmToastService: UtmToastService,
              public inputClass: InputClassResolve,
              private cdref: ChangeDetectorRef) {
  }

  ngOnInit() {
    this.date = this.calendar.getToday();
    this.initFormSchedule();
    if (this.schedule) {
      this.setFormSchedule();
    }
  }

  ngAfterContentChecked() {
    this.cdref.detectChanges();
  }

  initFormSchedule() {
    this.formSchedule = this.fb.group({
      name: ['', Validators.required],
      comment: [''],
      startDate: ['', Validators.required],
      startTime: ['', Validators.required],
      period: ['', Validators.required],
      periodTime: ['', Validators.required],
      duration: ['', Validators.required],
      durationTime: ['', Validators.required],
      time: []
    }, {validator: timePeriodValidator});
  }

  setFormSchedule() {
    this.formSchedule.get('name').setValue(this.schedule.name);
    this.formSchedule.get('comment').setValue(this.schedule.comment);
    this.formSchedule.get('startDate').setValue(this.setDate());
    this.formSchedule.get('startTime').setValue(this.setStartTime());
    this.formSchedule.get('period').setValue(this.schedule.simplePeriod.value);
    this.formSchedule.get('periodTime').setValue(this.schedule.simplePeriod.unit);
    this.formSchedule.get('duration').setValue(this.schedule.simpleDuration.value);
    this.formSchedule.get('durationTime').setValue(this.schedule.simpleDuration.unit);
  }

  setDate() {
    const firsTime = new Date(this.schedule.firstTime);
    return {year: firsTime.getFullYear(), month: firsTime.getMonth() - 1, day: firsTime.getDay()};
  }

  setStartTime() {
    const firsTime = new Date(this.schedule.firstTime);
    return {hour: firsTime.getHours(), minute: firsTime.getMinutes()};
  }

  createSchedule() {
    this.creating = true;
    this.scheduleService.create(this.buildObjectRequest()).subscribe(scheduleCreated => {
      this.utmToastService.showSuccessBottom('Schedule created successfully');
      this.scheduleCreated.emit(scheduleCreated.body.id);
      this.activeModal.close();
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating schedule',
        'Error creating schedule, check your network');
    });
  }

  editSchedule() {
    this.creating = true;
    this.scheduleService.update(this.buildObjectRequest()).subscribe(scheduleCreated => {
      this.utmToastService.showSuccessBottom('Schedule edited successfully');
      this.scheduleCreated.emit(scheduleCreated.body.id);
      this.activeModal.close();
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error edited schedule',
        'Error creating schedule, check your network');
    });
  }

  buildObjectRequest() {
    const d: { year: number, month: number, day: number } = this.formSchedule.get('startDate').value;
    const schedule = {
      comment: this.formSchedule.get('comment').value,
      name: this.formSchedule.get('name').value,
      firstTime: {
        dayOfMonth: d.day,
        month: d.month,
        year: d.year,
        hour: this.formSchedule.get('startTime').value.hour,
        minute: this.formSchedule.get('startTime').value.minute
      },
      period: {
        unit: this.formSchedule.get('periodTime').value,
        value: this.formSchedule.get('period').value,
      },
      duration: {
        unit: this.formSchedule.get('durationTime').value,
        value: this.formSchedule.get('duration').value,
      },
      // timezone: 'string'
    };
    return schedule;
  }

  datepickerToday(): { year: number, month: number, day: number } {
    const today = new Date();
    return {year: today.getFullYear(), month: today.getMonth() - 1, day: today.getDay()};
  }

}
