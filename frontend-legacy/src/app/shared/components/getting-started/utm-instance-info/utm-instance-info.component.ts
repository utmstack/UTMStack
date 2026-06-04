import {Component, Input, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {UtmConfigParamsService} from '../../../services/config/utm-config-params.service';
import {LocationService, SelectOption} from '../../../services/location.service';
import {ConfigDataTypeEnum, SectionConfigParamType} from '../../../types/configuration/section-config-param.type';

@Component({
  selector: 'app-utm-instance-info',
  templateUrl: './utm-instance-info.component.html',
  styleUrls: ['./utm-instance-info.component.css']
})
export class UtmInstanceInfoComponent implements OnInit {
  @Input() formConfigs: SectionConfigParamType[] | null = null;
  dynamicForm: FormGroup;
  selectOptions: { [key: string]: SelectOption[] } = {};
  isLoading: { [key: string]: boolean } = {};
  ConfigDataTypeEnum = ConfigDataTypeEnum;


  constructor(private fb: FormBuilder,
              private locationService: LocationService,
              private utmConfigParamsService: UtmConfigParamsService,
              private toastService: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    if (this.formConfigs && this.formConfigs.length > 0) {
      const controlsConfig: any = {};
      this.formConfigs.forEach(config => {
        const validators = [];
        if (config.confParamRequired) {
          validators.push(Validators.required);
        }
        if (config.confParamDatatype === 'email' && config.confParamRegexp) {
          validators.push(Validators.pattern(config.confParamRegexp));
        }
        if (config.confParamDatatype === ConfigDataTypeEnum.CountryList) {
          this.loadSelectOptions(this.getName(config.confParamShort));
        }
        controlsConfig[this.getName(config.confParamShort)] = [ config.confParamValue ? config.confParamValue : '', validators];
      });
      this.dynamicForm = this.fb.group(controlsConfig);
    } else {
      this.dynamicForm = this.fb.group({});
    }
  }

  loadSelectOptions(controlName: string): void {
    this.isLoading[controlName] = true;

    this.locationService.getCountries().subscribe(
      (response) => {
        if (response.body) {
          this.selectOptions[controlName] = response.body;
        } else {
          this.selectOptions[controlName] = [];
        }
        this.isLoading[controlName] = false;
      },
      (error) => {
        this.selectOptions[controlName] = [];
        this.isLoading[controlName] = false;
      }
    );
  }

  getSelectOptions(controlName: string): SelectOption[] {
    return this.selectOptions[controlName] || [];
  }

  getName(name: string): string {
    return name.split('.').pop()!;
  }

  onSubmit() {
    const config = this.formConfigs.map(config => ({
      ...config,
      confParamValue: this.dynamicForm.get(this.getName(config.confParamShort)).value
    }));
    this.utmConfigParamsService.update(config)
      .subscribe(
      () => {
        this.modalService.dismissAll('success');
        this.toastService.showSuccess('Configuration updated successfully');
      },
      () => this.toastService.showInfo('Error', 'Failed to update configuration'));
  }
}
