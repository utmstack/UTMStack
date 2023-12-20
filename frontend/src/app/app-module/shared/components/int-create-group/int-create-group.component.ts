import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {UtmModuleGroupService} from '../../services/utm-module-group.service';
import {UtmModuleGroupType} from '../../type/utm-module-group.type';

@Component({
  selector: 'app-int-create-group',
  templateUrl: './int-create-group.component.html',
  styleUrls: ['./int-create-group.component.css']
})
export class IntCreateGroupComponent implements OnInit {
  @Input() group: UtmModuleGroupType;
  @Input() moduleId: number;
  @Output() groupChange = new EventEmitter<UtmModuleGroupType>();
  formGroupConfig: FormGroup;
  adding = false;

  constructor(private fb: FormBuilder,
              public activeModal: NgbActiveModal,
              private groupService: UtmModuleGroupService,
              private toast: UtmToastService,
              public inputClass: InputClassResolve) {
  }

  ngOnInit() {
    this.formGroupConfig = this.fb.group({
      groupDescription: [''],
      groupName: ['', Validators.required],
      moduleId: [this.moduleId, [Validators.required]],
      id: []
    });
    if (this.group) {
      this.formGroupConfig.patchValue(this.group);
    }
  }

  saveGroup() {
    this.adding = true;
    if (this.group) {
      this.editGroup();
    } else {
      this.createGroup();
    }
  }

  createGroup() {
    this.groupService.create(this.getFormValue()).subscribe(response => {
      this.emitGroup(response.body);
    }, () => {
      this.error('creating');
    });
  }

  editGroup() {
    this.groupService.update(this.formGroupConfig.value).subscribe(response => {
      this.emitGroup(response.body);
    }, () => {
      this.error('editing');
    });
  }

  emitGroup(group: UtmModuleGroupType) {
    this.toast.showSuccessBottom('Tenant group saved successfully');
    this.groupChange.emit(group);
    this.activeModal.close();
  }

  getFormValue() {
    return {
      description: this.formGroupConfig.get('groupDescription').value,
      moduleId: this.moduleId,
      name: this.formGroupConfig.get('groupName').value
    };
  }

  error(type: 'editing' | 'creating') {
    this.toast.showError('Error ' + type + ' tenant', 'Error ' + type + ' ,tenant group configuration');
  }

}
