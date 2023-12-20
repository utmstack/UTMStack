import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupService} from '../../shared/services/utm-module-group.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';


@Component({
  selector: 'app-int-aws-groups',
  templateUrl: './int-aws-groups.component.html',
  styleUrls: ['./int-aws-groups.component.scss']
})
export class IntAwsGroupsComponent implements OnInit {
  awsGroup: { name: string, module: string };
  @Input() integrationId: number;
  @Input() confKey: string;
  loadingConf: boolean;
  saving = false;
  moduleConfig: UtmModuleGroupConfType[];
  typing = false;
  timer: any;

  constructor(private utmModuleGroupConfService: UtmModuleGroupConfService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.loadConfig();
  }

  loadConfig() {
    // this.loadingConf = true;
    // this.utmModuleGroupConfService.loadConfig(this.confKey, this.integrationId).subscribe(value => {
    //   this.loadingConf = false;
    //   this.moduleConfig = value[0];
    //   this.awsGroup = this.moduleConfig.confValue ? JSON.parse(this.moduleConfig.confValue) : {name: '', module: ''};
    // });
  }

  save($event: any) {
    this.typing = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      if ((this.awsGroup.name !== '' && this.awsGroup.name !== null && this.awsGroup.module !== '' && this.awsGroup.module !== null)) {
        this.saveConfig(JSON.stringify(this.awsGroup));
      }
      this.typing = false;
    }, 1500);
  }

  saveConfig(value: any) {
    // this.saving = true;
    // this.utmIntConfigManageService.saveConfig(this.moduleConfig, value).subscribe(saved => {
    //   if (saved) {
    //     this.toastService.showSuccessBottom('Configuration saved successfully');
    //     this.saving = false;
    //   } else {
    //     this.saving = false;
    //     this.toastService.showError('Error', 'Error saving configuration, go to application logs for more details');
    //   }
    // });
  }

}
