import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {RestartApiBehavior} from '../../../../behaviors/restart-api.behavior';
import {UtmConfigParamsService} from '../../../../services/config/utm-config-params.service';
import {SectionConfigParamType} from '../../../../types/configuration/section-config-param.type';


@Component({
  selector: 'app-config-params',
  templateUrl: './app-config-params.component.html',
  styleUrls: ['./app-config-params.scss']
})
export class AppConfigParamsComponent implements OnInit {
  @Input() sectionConfigId: number;
  @Output() validateConfig = new EventEmitter<boolean>();
  @Output() configChange = new EventEmitter<SectionConfigParamType>();
  configs: SectionConfigParamType[] = [];
  loading = true;
  timer: any;
  typing: any;
  saving = false;

  constructor(private utmConfigParamsService: UtmConfigParamsService,
              private restartApiBehavior: RestartApiBehavior,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.getConfigurations();
  }


  getConfigurations() {
    this.loading = true;
    this.utmConfigParamsService.query({
      page: 0,
      size: 10000,
      'sectionId.equals': this.sectionConfigId,
      sort: 'id,asc'
    })
      .subscribe(response => {
        this.loading = false;
        this.configs = response.body;
        this.validateConfig.emit(this.checkConfigValid());
      }, error => {
        this.toastService.showError('Error', 'Error getting application configurations');
      });
  }

  saveSectionConfig(value: any, conf: SectionConfigParamType) {
    conf.confParamValue = value;
    this.configChange.emit(conf);
  }


  checkConfigValid(): boolean {
    let valid = true;
    for (const conf of this.configs) {
      if (conf.confParamRequired) {
        const validateConf = this.checkConfigValue(conf.confParamValue);
        if (!validateConf) {
          valid = validateConf;
        }
      }
    }
    return valid;
  }

  checkConfigValue(config: string): boolean {
    return config !== null && config !== '' && config !== undefined;
  }

  save($event: any, conf: SectionConfigParamType) {
    this.typing = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      this.saveSectionConfig($event, conf);
      this.typing = false;
    }, 1000);
  }
}
