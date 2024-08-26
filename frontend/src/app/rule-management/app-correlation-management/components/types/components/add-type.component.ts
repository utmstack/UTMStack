import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {DataType} from '../../../../models/rule.model';
import {DataTypeService} from '../../../../services/data-type.service';

@Component({
  selector: 'app-add-type',
  templateUrl: './add-type.component.html',
  styleUrls: ['./add-type.component.scss']
})
export class AddTypeComponent implements OnInit {

  @Input() type: DataType;
  dataTypeForm: FormGroup;
  loading = false;
  mode: 'ADD' | 'EDIT';

  constructor(public activeModal: NgbActiveModal,
              private dataTypeService: DataTypeService,
              private formBuilder: FormBuilder,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.mode = this.type ? 'EDIT' : 'ADD';
    this.initForm();
  }

  get dataType() {
    return this.dataTypeForm.get('dataType');
  }

  get dataTypeName() {
    return this.dataTypeForm.get('dataTypeName');
  }

  get dataTypeDescription() {
    return this.dataTypeForm.get('dataTypeDescription');
  }


  initForm() {
    this.dataTypeForm = this.formBuilder.group({
      dataType: [ this.type ? this.type.dataType : '', [Validators.required, Validators.maxLength(250)]],
      dataTypeName: [ this.type ? this.type.dataTypeName : '', [Validators.required, Validators.maxLength(250)]],
      dataTypeDescription: [this.type ? this.type.dataTypeDescription : '']
    });
  }

  onSubmit() {
    if (this.dataTypeForm.valid) {
      this.loading = true;
      const formType: DataType = {
        ...this.dataTypeForm.value,
        included: false
      };
      const typeToSave = this.mode === 'ADD' ? formType : {
        ...this.type,
        ...formType
      };
      this.dataTypeService
        .saveDataType(this.mode, typeToSave)
           .subscribe({
             next: response => {
               console.log('Data type saved successfully', response);
               this.dataTypeForm.reset();
               this.loading = false;
               this.utmToastService.showSuccessBottom(this.mode === 'ADD'
                       ? 'Data type saved successfully' : 'Data type edited successfully');
               this.activeModal.close(true);
             },
             error: err => {
               this.loading = false;
               this.utmToastService.showError('Error', this.mode === 'ADD'
                   ? 'Error saving rule' : 'Error editing rule');
               console.error('Error saving rule:', err.message);
             }
           });
    }
  }

}
