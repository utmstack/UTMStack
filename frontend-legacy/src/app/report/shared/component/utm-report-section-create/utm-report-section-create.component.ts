import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {SectionReportService} from '../../service/section-report.service';
import {ReportSectionType} from '../../type/report-section.type';

@Component({
  selector: 'app-utm-report-section-create',
  templateUrl: './utm-report-section-create.component.html',
  styleUrls: ['./utm-report-section-create.component.scss']
})
export class UtmReportSectionCreateComponent implements OnInit {
  @Input() section: ReportSectionType;
  @Output() sectionSaved = new EventEmitter<ReportSectionType>();
  creating = false;
  formReportSection: FormGroup;

  constructor(public activeModal: NgbActiveModal,
              public modalService: NgbModal,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              private sectionReportService: SectionReportService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormStandard();
    if (this.section) {
      this.formReportSection.patchValue(this.section);
    }
  }

  initFormStandard() {
    this.formReportSection = this.fb.group({
      id: [],
      repSecDescription: ['', Validators.required],
      repSecName: ['', Validators.required],
      repSecSystem: [false]
    });
  }

  createStandard() {
    this.creating = true;
    if (this.section) {
      this.editStandard();
    } else {
      this.saveStandard();
    }
  }

  private editStandard() {
    this.sectionReportService.update(this.formReportSection.value).subscribe((standard) => {
      this.utmToastService.showSuccessBottom('Report section updated successfully');
      this.activeModal.close();
      this.sectionSaved.emit(standard.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error updating report section ');
    });
  }

  private saveStandard() {
    this.sectionReportService.create(this.formReportSection.value).subscribe(standard => {
      this.utmToastService.showSuccessBottom('Report section  created successfully');
      this.activeModal.close();
      this.sectionSaved.emit(standard.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error creating report section ');
    });
  }

}
