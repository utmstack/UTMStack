import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {ReportTypeEnum} from '../../enums/report-type.enum';
import {ReportService} from '../../service/report.service';
import {SectionReportService} from '../../service/section-report.service';
import {ReportSectionType} from '../../type/report-section.type';
import {ReportType} from '../../type/report.type';

@Component({
  selector: 'app-utm-report-create',
  templateUrl: './utm-report-create.component.html',
  styleUrls: ['./utm-report-create.component.scss']
})
export class UtmReportCreateComponent implements OnInit {
  @Input() report: ReportType;
  @Output() reportCreated = new EventEmitter<string>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  formReport: FormGroup;
  reportSections: ReportSectionType[];


  constructor(private sectionReportService: SectionReportService,
              private reportService: ReportService,
              public activeModal: NgbActiveModal,
              private fb: FormBuilder,
              private utmToastService: UtmToastService,
              public inputClass: InputClassResolve,
              public modalService: NgbModal) {
  }

  ngOnInit() {
    this.getReportSections();
    this.initFormReport();
    if (this.report) {
      this.formReport.patchValue(this.report);
    }
  }

  getReportSections() {
    this.sectionReportService.query({size: 500, 'repSecSystem.equals': false}).subscribe(response => {
      this.reportSections = response.body;
    });
  }

  initFormReport() {
    this.formReport = this.fb.group({
      dashboardId: [],
      id: [],
      repDescription: ['', [Validators.required, Validators.maxLength(512)]],
      repName: ['',  [Validators.required, Validators.maxLength(255)]],
      reportSectionId: [null, Validators.required],
      repType: [ReportTypeEnum.TEMPLATE]
    });
  }

  backStep() {
    this.step -= 1;
    this.stepCompleted.pop();
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }


  createCompliance() {
    this.creating = true;

    if (this.report) {
      this.reportService.update(this.formReport.value).subscribe(() => {
        this.utmToastService.showSuccessBottom('Report edited successfully');
        this.activeModal.close();
        this.reportCreated.emit('edited');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error editing report');
      });
    } else {
      this.reportService.create(this.formReport.value).subscribe(() => {
        this.utmToastService.showSuccessBottom('Report  created successfully');
        this.activeModal.close();
        this.reportCreated.emit('created');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error creating report');
      });
    }
  }

  onDashboardSelected($event: number) {
    this.formReport.get('dashboardId').setValue($event);
  }
}
