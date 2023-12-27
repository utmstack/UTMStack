import {Component, Input, OnInit} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';


@Component({
  selector: 'app-int-ldap-groups',
  templateUrl: './int-ldap-groups.component.html',
  styleUrls: ['./int-ldap-groups.component.scss']
})
export class IntLdapGroupsComponent implements OnInit {
  ldapForm: FormGroup;
  @Input() integrationId: number;
  @Input() confKey: string;
  loadingConf: boolean;
  saving = false;
  integrationConfig: UtmModuleGroupConfType;
  typing = false;
  timer: any;

  constructor(private utmIntConfigManageService: UtmModuleGroupConfService,
              private toastService: UtmToastService,
              private fb: FormBuilder) {
  }

  get group() {
    return this.ldapForm.controls.group as FormArray;
  }

  ngOnInit() {
    this.loadConfig();
    this.initForm();
  }

  loadConfig() {
    // this.loadingConf = true;
    // this.utmIntConfigManageService.loadConfig(this.confKey, this.integrationId).subscribe(value => {
    //   this.loadingConf = false;
    //   this.integrationConfig = value[0];
    //   this.setFormValue();
    // });
  }

  setFormValue() {
    if (this.integrationConfig.confValue) {
      const groupLdap = JSON.parse(this.integrationConfig.confValue);
      for (const group of groupLdap) {
        this.addLdapGroup(group);
      }
    } else {
      this.addLdapGroup();
    }
  }

  // save(value: any) {
  //   this.typing = true;
  //   if (this.timer) {
  //     clearTimeout(this.timer);
  //   }
  //   this.timer = setTimeout(() => {
  //     if (this.ldapForm.valid) {
  //       this.saveConfig(JSON.stringify(value));
  //     }
  //     this.typing = false;
  //   }, 1500);
  // }
  //
  // saveAll() {
  //   if (this.ldapForm.valid) {
  //     for (const group of this.group.value) {
  //       console.log(group);
  //       this.saveConfig(JSON.stringify(group));
  //     }
  //   }
  // }

  saveConfig() {
    // this.saving = true;
    // console.log(this.integrationConfig);
    // this.utmIntConfigManageService.saveConfig(this.integrationConfig, JSON.stringify(this.group.value)).subscribe(saved => {
    //   if (saved) {
    //     this.toastService.showSuccessBottom('Configuration saved successfully');
    //     this.saving = false;
    //   } else {
    //     this.saving = false;
    //     this.toastService.showError('Error', 'Error saving configuration, go to application logs for more details');
    //   }
    // });
  }

  initForm() {
    this.ldapForm = this.fb.group({
      group: this.fb.array([])
    });
  }

  addLdapGroup(group?: any) {
    this.group.push(this.fb.group({
      host: [group ? group.host : '', [Validators.required]],
      user_dn: [group ? group.user_dn : '', [Validators.required]],
      password: [group ? group.password : '', [Validators.required]],
      search_base: [group ? group.search_base : '', [Validators.required]],
    }));
  }

  deleteLdapGroup(index) {
    this.group.removeAt(index);
    if (!isNaN(index - 1)) {
      this.group.at(index - 1).get('subBucket').setValue(null);
    }
  }

}
