import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {CpStandardService} from '../../services/cp-standard.service';
import {ComplianceStandardType} from '../../type/compliance-standard.type';

@Component({
  selector: 'app-utm-cp-standard-create',
  templateUrl: './utm-cp-standard-create.component.html',
  styleUrls: ['./utm-cp-standard-create.component.scss']
})
export class UtmCpStandardCreateComponent implements OnInit {
  @Input() standard: ComplianceStandardType;
  @Output() standardSaved = new EventEmitter<ComplianceStandardType>();
  creating = false;
  formStandard: FormGroup;

  constructor(public activeModal: NgbActiveModal,
              public modalService: NgbModal,
              public inputClass: InputClassResolve,
              private utmToastService: UtmToastService,
              private cpStandardService: CpStandardService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.initFormStandard();
    if (this.standard) {
      this.formStandard.patchValue(this.standard);
    }
  }

  initFormStandard() {
    this.formStandard = this.fb.group({
      id: [],
      standardDescription: ['', Validators.required],
      standardName: ['', Validators.required]
    });
  }

  createStandard() {
    this.creating = true;
    if (this.standard) {
      this.editStandard();
    } else {
      this.saveStandard();
    }
  }

  private editStandard() {
    this.cpStandardService.update(this.formStandard.value).subscribe((standard) => {
      this.utmToastService.showSuccessBottom('Standard updated successfully');
      this.activeModal.close();
      this.standardSaved.emit(standard.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error updating standard');
    });
  }

  private saveStandard() {
    this.cpStandardService.create(this.formStandard.value).subscribe(standard => {
      this.utmToastService.showSuccessBottom('Standard created successfully');
      this.activeModal.close();
      this.standardSaved.emit(standard.body);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error creating standard');
    });
  }
}
