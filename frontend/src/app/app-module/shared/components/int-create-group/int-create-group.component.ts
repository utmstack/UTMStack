import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {UtmModuleGroupService} from '../../services/utm-module-group.service';
import {UtmModuleGroupType} from '../../type/utm-module-group.type';
import {GroupTypeEnum} from "../../enum/group-type.enum";
import {UtmModuleCollectorService} from "../../services/utm-module-collector.service";
import {UtmModulesEnum} from "../../enum/utm-module.enum";
import {UtmModuleCollectorType} from "../../type/utm-module-collector.type";

@Component({
  selector: 'app-int-create-group',
  templateUrl: './int-create-group.component.html',
  styleUrls: ['./int-create-group.component.css']
})
export class IntCreateGroupComponent implements OnInit {
  @Input() group: UtmModuleGroupType;
  @Input() moduleId: number;
  @Input() collectors: any[];
  @Input() groupType: number = GroupTypeEnum.TENANT;
  @Output() groupChange = new EventEmitter<UtmModuleGroupType>();
  formGroupConfig: FormGroup;
  adding = false;
  GroupTypeEnum = GroupTypeEnum;

  constructor(private fb: FormBuilder,
              public activeModal: NgbActiveModal,
              private groupService: UtmModuleGroupService,
              private toast: UtmToastService,
              public inputClass: InputClassResolve,
              private collectorService: UtmModuleCollectorService) {
  }

  ngOnInit() {
    this.formGroupConfig = this.fb.group({
      collector: this.groupType === GroupTypeEnum.COLLECTOR ? ['', Validators.required] : [],
      groupDescription: [''],
      groupName: this.groupType === GroupTypeEnum.TENANT ? ['', Validators.required] : [],
      moduleId: [this.moduleId, [Validators.required]],
      id: []
    });
    if (this.group) {
      const groupToEdit = this.groupType === GroupTypeEnum.TENANT ? this.group :
        {
          ...this.group,
          collector: parseInt(this.group.collector, 10)
        };
      this.formGroupConfig.patchValue(groupToEdit);
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
    this.toast.showSuccessBottom(`${this.groupType === GroupTypeEnum.TENANT ? 'Tenant' : 'Collector'} ` +
    `group saved successfully`);
    this.groupChange.emit(group);
    this.activeModal.close();
  }

  getFormValue() {
    const collector = this.collectors.find(c => c.id ===
        this.formGroupConfig.get('collector').value);

    return {
      description: this.formGroupConfig.get('groupDescription').value,
      moduleId: this.moduleId,
      name: this.groupType === GroupTypeEnum.TENANT ? this.formGroupConfig.get('groupName').value :
          this.collectorService.generateUniqueName(collector.collector, collector.groups),
      collector: this.formGroupConfig.get('collector').value ? this.formGroupConfig.get('collector').value : null
    };
  }

  error(type: 'editing' | 'creating') {
    this.toast.showError('Error ' + type + ' tenant', 'Error ' + type + ' ,tenant group configuration');
  }

  getHeader() {
    return ` ${this.getAction()}  ${this.getGroupType()}`;
  }

  getGroupType() {
    return this.groupType === GroupTypeEnum.TENANT ? 'tenant' : 'collector';
  }

  getAction() {
    return this.group ? 'Edit' : 'Add';
  }

}
