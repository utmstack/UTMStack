import {Component, Input, OnInit} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';


@Component({
  selector: 'app-int-google-projects',
  templateUrl: './int-google-projects.component.html',
  styleUrls: ['./int-google-projects.component.css']
})
export class IntGoogleProjectsComponent implements OnInit {
  googleForm: FormGroup;
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
    return this.googleForm.controls.group as FormArray;
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
      const groupGoogle = JSON.parse(this.integrationConfig.confValue);
      for (const group of groupGoogle) {
        this.addGoogleGroup(group);
      }
    } else {
      this.addGoogleGroup();
    }
  }

  saveConfig() {
    this.saving = true;
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
    this.googleForm = this.fb.group({
      group: this.fb.array([])
    });
  }

  addGoogleGroup(group?: any) {
    this.group.push(this.fb.group({
      projectId: [group ? group.projectId : '', [Validators.required]],
      topic: [group ? group.topic : '', [Validators.required]],
      subscription: [group ? group.subscription : '', [Validators.required]],
      jsonKey: [group ? group.jsonKey : '', [Validators.required]],
    }));
  }

  deleteGoogleGroup(index) {
    this.group.removeAt(index);
    if (!isNaN(index - 1)) {
      this.group.at(index - 1).get('subBucket').setValue(null);
    }
  }
}
