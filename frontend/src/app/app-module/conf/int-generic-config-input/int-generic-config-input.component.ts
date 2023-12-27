import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmModuleGroupConfService} from '../../shared/services/utm-module-group-conf.service';
import {UtmModuleGroupConfType} from '../../shared/type/utm-module-group-conf.type';


@Component({
  selector: 'app-int-generic-config-input',
  templateUrl: './int-generic-config-input.component.html',
  styleUrls: ['./int-generic-config-input.component.scss']
})
export class IntGenericConfigInputComponent implements OnInit {
  @Input() groupId: number;
  @Input() confKey: string;
  integrationConfig: UtmModuleGroupConfType;
  loadingConf: boolean;
  saving = false;
  typing = false;
  timer: any;
  init = 0;

  constructor(private utmModuleGroupConfService: UtmModuleGroupConfService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.loadConfig();
  }

  loadConfig() {
    this.loadingConf = true;
    this.utmModuleGroupConfService.query({groupId: this.groupId, confKey: this.confKey}).subscribe(value => {
      this.loadingConf = false;
      this.integrationConfig = value[0];
    });
  }

  save($event: any) {
    this.typing = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      this.saveConfig($event.target.value);
      this.typing = false;
    }, 1500);
  }

  saveConfig(value: any) {
    this.saving = true;
    this.utmModuleGroupConfService.saveConfig(value).subscribe(saved => {
      if (saved) {
        this.toastService.showSuccessBottom('Configuration saved successfully');
        this.saving = false;
      } else {
        this.saving = false;
        this.toastService.showError('Error', 'Error saving configuration, go to application logs for more details');
      }
    });
  }
}
