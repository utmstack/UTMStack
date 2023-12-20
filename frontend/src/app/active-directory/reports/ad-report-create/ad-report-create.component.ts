import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ElasticTimeEnum} from '../../../shared/enums/elastic-time.enum';
import {ElasticFilterCommonType} from '../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {InputClassResolve} from '../../../shared/util/input-class-resolve';
import {resolveRangeByTime} from '../../../shared/util/resolve-date';
import {TIME_RANGE_MILLISECONDS} from '../../shared/const/time-range-milliseconds.const';
import {ActiveDirectoryTreeType} from '../../shared/types/active-directory-tree.type';
import {AdReportTypeEnum} from '../../tracker/shared/enums/ad-report-type.enum';
import {AdTrackerType} from '../../tracker/shared/type/ad-tracker.type';
import {AdReportService} from '../shared/services/ad-report.service';
import {AdReportType} from '../shared/type/ad-report.type';

@Component({
  selector: 'app-ad-report-create',
  templateUrl: './ad-report-create.component.html',
  styleUrls: ['./ad-report-create.component.scss']
})
export class AdReportCreateComponent implements OnInit {
  @Input() data: ActiveDirectoryTreeType[];
  reportType: { key: string, value: string }[];
  limitRange = [50, 100, 200, 500, 1000, 2500];
  range = TIME_RANGE_MILLISECONDS;
  saving = false;
  generateReport = false;
  formAdReport: FormGroup;
  adReportTypeEnum = AdReportTypeEnum;
  private filterTime: TimeFilterType = resolveRangeByTime('day');
  defaultTime: ElasticFilterCommonType = {last: 1, time: ElasticTimeEnum.DAY, label: 'last 1 day'};

  constructor(public activeModal: NgbActiveModal,
              private toastService: UtmToastService,
              private adReportService: AdReportService,
              public inputClass: InputClassResolve,
              private fb: FormBuilder
  ) {
    this.reportType = [
      {key: this.adReportTypeEnum.ACTIVITY, value: 'Events'},
      {key: this.adReportTypeEnum.DETAIL, value: 'Detail'}
    ];
  }

  ngOnInit() {
    this.initFormReport();
  }

  initFormReport() {
    this.formAdReport = this.fb.group({
      name: [''],
      description: [''],
      type: ['', Validators.required],
      schedule: [''],
      scheduleRange: [86400000],
      limit: [50, Validators.required],
      objects: [''],
      emails: [''],
      startTime: [''],
    });
  }

  saveReport() {
    this.saving = true;
    const report: AdReportType = {
      description: this.formAdReport.get('description').value,
      emails: this.formAdReport.get('emails').value,
      limit: this.formAdReport.get('limit').value,
      name: this.formAdReport.get('name').value,
      nextExecution: this.calcNextExecution(),
      objectsType: this.formAdReport.get('objects').value,
      schedule: this.formAdReport.get('scheduleRange').value,
      type: this.formAdReport.get('type').value,
    };
    this.adReportService.create(report).subscribe((rep) => {
        this.saving = false;
        this.activeModal.close();
        this.toastService.showSuccessBottom(
          'Report ' + report.name + ' has benn saved to your active directory reports');
      }, error1 => {
        this.toastService.showError('Error saving report ',
          'Error creating report, check your network');
      }
    );
  }

  viewReport() {
    let report: AdReportType;
    if (this.formAdReport.get('type').value === this.adReportTypeEnum.ACTIVITY) {
      report = {
        limit: this.formAdReport.get('limit').value,
        objectsType: this.formAdReport.get('objects').value,
        type: this.formAdReport.get('type').value,
        from: this.filterTime.timeFrom,
        to: this.filterTime.timeTo,
      };
    } else {
      report = {
        objectsType: this.formAdReport.get('objects').value,
        type: this.formAdReport.get('type').value,
      };
    }
    this.generateReport = true;
    this.adReportService.exportToPdf(report).subscribe((dat) => {
      this.generateReport = false;
      const data = new Blob([dat], {type: 'application/pdf'});
      if (data.size > 0) {
        this.activeModal.close();
        const fileURL = window.URL.createObjectURL(data);
        window.open(fileURL);
      } else {
        this.toastService.showWarning('No data', 'No data found for this report');
      }
    });
  }

  calcNextExecution(): Date {
    const milliseconds = this.formAdReport.get('scheduleRange').value;
    const hour: { hour: 5, minute: 10, second: 0 } = this.formAdReport.get('startTime').value;
    const date = new Date();
    date.setHours(hour.hour);
    date.setMinutes(hour.minute);
    date.setSeconds(hour.second);
    return new Date(date.getTime() + milliseconds);
  }

  timeFilterChange($event: TimeFilterType) {
    this.filterTime = $event;
  }

  onTableChange($event: AdTrackerType[]) {
    const objects: {
      'objectName': string,
      'objectSid': string
    }[] = [];
    for (const obj of $event) {
      objects.push({
        objectName: obj.objectName,
        objectSid: obj.objectId
      });
    }
    this.formAdReport.get('objects').setValue(objects);
  }

  onEmailChange($event: string) {
    this.formAdReport.get('emails').setValue($event);
  }

  setSchedule($event: boolean) {
    this.formAdReport.get('schedule').setValue($event);
    this.setFormValidity($event);
  }

  setFormValidity(activevalidators: boolean) {
    if (activevalidators) {
      this.formAdReport.get('name').setValidators(Validators.required);
      this.formAdReport.get('schedule').setValidators(Validators.required);
      this.formAdReport.get('scheduleRange').setValidators(Validators.required);
      this.formAdReport.get('objects').setValidators(Validators.required);
      this.formAdReport.get('emails').setValidators(Validators.required);
      this.formAdReport.get('startTime').setValidators(Validators.required);
    } else {
      this.formAdReport.get('name').setValidators(null);
      this.formAdReport.get('schedule').setValidators(null);
      this.formAdReport.get('scheduleRange').setValidators(null);
      this.formAdReport.get('objects').setValidators(null);
      this.formAdReport.get('emails').setValidators(null);
      this.formAdReport.get('startTime').setValidators(null);
    }

    this.formAdReport.get('name').updateValueAndValidity();
    this.formAdReport.get('schedule').updateValueAndValidity();
    this.formAdReport.get('scheduleRange').updateValueAndValidity();
    this.formAdReport.get('objects').updateValueAndValidity();
    this.formAdReport.get('emails').updateValueAndValidity();
    this.formAdReport.get('startTime').updateValueAndValidity();
    this.formAdReport.updateValueAndValidity();
  }
}
