import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../alert/utm-toast.service';
import {RestartApiBehavior} from '../../../../behaviors/restart-api.behavior';
import {
  DATE_FORMATS,
  DATE_SETTING_FORMAT_SHORT,
  DATE_SETTING_TIMEZONE_SHORT,
  TIMEZONES
} from '../../../../constants/date-timezone-date.const';
import {UtmConfigParamsService} from '../../../../services/config/utm-config-params.service';
import {ConfigDataTypeEnum, SectionConfigParamType} from '../../../../types/configuration/section-config-param.type';
import {ApplicationConfigSectionEnum, SectionConfigType} from '../../../../types/configuration/section-config.type';
import {AppConfigDeleteConfirmComponent} from '../app-config-delete-confirm/app-config-delete-confirm.component';


@Component({
  selector: 'app-config-sections',
  templateUrl: './app-config-sections.component.html',
  styleUrls: ['./app-config-sections.component.css']
})
export class AppConfigSectionsComponent implements OnInit, OnDestroy {

  @Input() section: SectionConfigType;
  @Output() validConfigSection = new EventEmitter<boolean>();
  @Input() allowDeleteSection = false;
  @Output() changesApplied = new EventEmitter<boolean>();
  @Output() deleteSection = new EventEmitter<number>();
  configs: SectionConfigParamType[] = [];
  saving: any;
  configToSave: SectionConfigParamType[] = [];
  loading = true;
  timer: any;
  typing: any;
  configDataTypeEnum = ConfigDataTypeEnum;
  timezones = TIMEZONES;
  dateFormats = DATE_FORMATS;
  isCheckedEmailConfig = false;
  sectionType = ApplicationConfigSectionEnum;

  constructor(private utmConfigParamsService: UtmConfigParamsService,
              private modalService: NgbModal,
              private restartApiBehavior: RestartApiBehavior,
              private toastService: UtmToastService) {
  }


  ngOnInit() {
    this.getConfigurations();
    this.changesApplied.emit(true);
  }

  ngOnDestroy(): void {
    if (this.configToSave.length > 0) {
    }
  }

  getConfigurations() {
    this.loading = true;
    this.utmConfigParamsService.query({
      page: 0,
      size: 10000,
      'sectionId.equals': this.section.id,
      sort: 'id,asc'
    })
      .subscribe(response => {
        this.loading = false;
        this.configs = response.body;
        this.validConfigSection.emit(this.checkConfigValid());
      }, error => {
        this.toastService.showError('Error', 'Error getting application configurations');
      });
  }

  saveConfig() {
    this.checkedEmailConfig(false);
    this.saving = true;
    if (this.checkConfigValid()) {
      this.utmConfigParamsService.update(this.configToSave).subscribe(response => {

          if (this.detectRequiredRestart()) {
            this.restartApiBehavior.$restartSignal.next(true);
          }
          this.saving = false;

          this.validConfigSection.emit(this.checkConfigValid());
          this.changesApplied.emit(true);

          this.toastService.showSuccessBottom('Configuration saved successfully');
          this.refreshTimeSettings();
          this.configToSave = [];
        },
        error => {
          this.saving = false;
          if (error.status === 412) {
            this.toastService.showError('Error', 'Before activating Multi-Factor Authentication (MFA), ' +
              'please ensure your email settings are configured correctly');
          } else {
            this.toastService.showError('Error', 'Error saving configuration, go to application logs for more details');
          }
        });
    } else {
      this.toastService.showInfo('Configuration not valid', 'There are required configurations that you need set before save');
    }
  }

  detectRequiredRestart(): boolean {
    return this.configToSave.findIndex(value => value.confParamRestartRequired === true) > -1;
  }

  saveSectionConfig(value: any, conf: SectionConfigParamType) {
    conf.confParamValue = value;
    const indexConfig = this.configToSave.findIndex(val => val.id === conf.id);
    if (indexConfig !== -1) {
      this.configToSave[indexConfig] = conf;
    } else {
      this.configToSave.push(conf);
    }
    this.validConfigSection.emit(this.checkConfigValid());
    this.changesApplied.emit(false);
  }

  checkConfigValid(): boolean {
    let valid = true;
    for (const conf of this.configs) {
      if (conf.confParamRequired || (conf.confParamRegexp && conf.confParamDatatype === ConfigDataTypeEnum.EmailList)) {
        const validateConf = this.checkConfigValue(conf);
        if (!validateConf) {
          valid = validateConf;
        }
      }
    }
    return valid;
  }

  checkConfigValue(config: SectionConfigParamType): boolean {
    switch (config.confParamDatatype) {
      case ConfigDataTypeEnum.EmailList:
        return this.isValid(config);

      default:
        return config.confParamValue !== null && config.confParamValue !== '' && config.confParamValue !== undefined;
    }
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

  deleteConfig() {
    const modal = this.modalService.open(AppConfigDeleteConfirmComponent, {centered: true});
    modal.componentInstance.section = this.section;
    modal.componentInstance.acceptDelete.subscribe(() => {
      this.deleteSection.emit(this.section.id);
    });
  }

  refreshTimeSettings() {
    const index = this.configToSave
      .findIndex(value => value.confParamShort === DATE_SETTING_TIMEZONE_SHORT
        || value.confParamShort === DATE_SETTING_FORMAT_SHORT);
    if (index !== -1) {
      window.location.reload();
      // const timezone = this.getConfigToSaveValue(DATE_SETTING_TIMEZONE_SHORT);
      // const dateFormat = this.getConfigToSaveValue(DATE_SETTING_FORMAT_SHORT);
      // const format: DatePipeDefaultOptions = {
      //   dateFormat: dateFormat ? dateFormat : DEFAULT_DATE_SETTING_DATE,
      //   timezone: timezone ? timezone : DEFAULT_DATE_SETTING_TIMEZONE
      // };
      // this.timezoneFormatService.setDateFormatSubject(format);
    }
  }

  isValid(conf: SectionConfigParamType) {
    return new RegExp(conf.confParamRegexp).test(conf.confParamValue);
  }

  getConfigToSaveValue(shortName: string): string {
    const index = this.configToSave
      .findIndex(value => value.confParamShort === shortName);
    if (index !== -1) {
      return this.configToSave[index].confParamValue;
    } else {
      return null;
    }
  }

  checkedEmailConfig(event: boolean){
    this.isCheckedEmailConfig = event;
  }

  isPasswordSet() {
    const conf = this.configToSave.find(conf => conf.confParamShort === 'utmstack.mail.password');
    return this.configToSave.length > 0 &&  conf && conf.confParamValue !== '';
  }

}
