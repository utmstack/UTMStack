import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {CpStandardSectionService} from '../../services/cp-standard-section.service';
import {ComplianceStandardSectionType} from '../../type/compliance-standard-section.type';

@Component({
  selector: 'app-utm-cp-standard-section-create',
  templateUrl: './utm-cp-standard-section-create.component.html',
  styleUrls: ['./utm-cp-standard-section-create.component.scss']
})
export class UtmCpStandardSectionCreateComponent implements OnInit {
  @Input() standardSection: ComplianceStandardSectionType;
  @Input() standardId: number;
  @Output() standardSectionSaved = new EventEmitter<ComplianceStandardSectionType>();
  creating = false;
  formStandardSection: FormGroup;

  constructor(public activeModal: NgbActiveModal,
              public modalService: NgbModal,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              private cpStandardSectionService: CpStandardSectionService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormStandard();
    if (this.standardSection) {
      this.formStandardSection.patchValue(this.standardSection);
    } else {
      this.formStandardSection.get('standardId').setValue(this.standardId);
    }
  }

  initFormStandard() {
    this.formStandardSection = this.fb.group({
      id: [],
      standardId: ['', Validators.required],
      standardSectionDescription: ['', Validators.required],
      standardSectionName: ['', Validators.required]
    });
  }

  createStandard() {
    this.creating = true;
    if (this.standardSection) {
      this.editStandardSection();
    } else {
      this.saveStandardSection();
    }
  }

  private editStandardSection() {
    this.cpStandardSectionService.update(this.formStandardSection.value).subscribe(response => {
      this.utmToastService.showSuccessBottom('Standard section updated successfully');
      this.activeModal.close();
      this.standardSectionSaved.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error updating standard section');
    });
  }

  private saveStandardSection() {
    this.cpStandardSectionService.create(this.formStandardSection.value).subscribe((response) => {
      this.utmToastService.showSuccessBottom('Standard section created successfully');
      this.activeModal.close();
      this.standardSectionSaved.emit(response.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error creating standard section');
    });
  }

}
