import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {SYSTEM_MODULE_ICONS_PATH} from '../../../../shared/constants/menu_icons.constants';
import {UtmModulesService} from '../../services/utm-modules.service';
import {UtmModuleCheckType} from '../../type/utm-module-check.type';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-checks',
  templateUrl: './app-module-checks.component.html',
  styleUrls: ['./app-module-checks.component.scss']
})
export class AppModuleChecksComponent implements OnInit {
  @Input() module: UtmModuleType;
  @Input() serverId: number;
  @Output() checkResult = new EventEmitter<boolean>();
  checks: UtmModuleCheckType;
  loading = true;
  iconPath = SYSTEM_MODULE_ICONS_PATH;

  constructor(public activeModal: NgbActiveModal,
              private utmModulesService: UtmModulesService) {
  }

  ngOnInit() {
    console.log(this.serverId);
    this.runCheck();
  }

  runCheck() {
    this.utmModulesService.checkConfig({nameShort: this.module.moduleName, serverId: this.serverId}).subscribe(response => {
      this.checks = response.body;
      this.loading = false;
    });
  }

  continue() {
    this.checkResult.emit(this.checks.status === 'OK');
    this.activeModal.close();
  }
}
