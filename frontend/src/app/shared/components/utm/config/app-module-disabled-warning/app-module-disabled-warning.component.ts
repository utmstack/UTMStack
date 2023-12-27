import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../../../../app-module/shared/enum/utm-module.enum';
import {UtmModulesService} from '../../../../../app-module/shared/services/utm-modules.service';
import {UtmRunModeService} from '../../../../services/active-modules/utm-run-mode.service';

@Component({
  selector: 'app-app-module-disabled-warning',
  templateUrl: './app-module-disabled-warning.component.html',
  styleUrls: ['./app-module-disabled-warning.component.css']
})
export class AppModuleDisabledWarningComponent implements OnInit {
  @Input() module: UtmModulesEnum;
  isLite = true;
  isModuleEnable = true;
  modulesEnum = UtmModulesEnum;
  show = true;
  moduleName = '';

  constructor(private moduleService: UtmModulesService, private utmRunModeService: UtmRunModeService) {
  }

  ngOnInit() {
    this.getModuleStatus();
  }

  getModuleStatus() {
    this.moduleService.isActive(this.module).subscribe(response => {
      this.utmRunModeService.isLiteMode().subscribe(rl => {
        this.isLite = rl.body;
        this.isModuleEnable = response.body;
        this.moduleName = this.module;
      });
    });
  }

}
