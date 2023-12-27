import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../../../shared/behaviors/nav.behavior';
import {ModuleRefreshBehavior} from '../../behavior/module-refresh.behavior';
import {UtmModulesEnum} from '../../enum/utm-module.enum';
import {UtmModulesService} from '../../services/utm-modules.service';
import {UtmModuleType} from '../../type/utm-module.type';
import {AppModuleChecksComponent} from '../app-module-checks/app-module-checks.component';
import {AppModuleDeactivateComponent} from '../app-module-deactivate/app-module-deactivate.component';

@Component({
  selector: 'app-app-module-activate-button',
  templateUrl: './app-module-activate-button.component.html',
  styleUrls: ['./app-module-activate-button.component.css']
})
export class AppModuleActivateButtonComponent implements OnInit {
  @Input() module: UtmModulesEnum;
  @Input() type: string;
  @Input() disabled = false;
  @Input() serverId: number;
  loadingDetail = true;
  moduleDetail: UtmModuleType;
  changingStatus: any;
  activatable: boolean;

  constructor(private utmModulesService: UtmModulesService,
              public modalService: NgbModal,
              private toastService: UtmToastService,
              private moduleRefreshBehavior: ModuleRefreshBehavior,
              private navBehavior: NavBehavior) {
  }

  ngOnInit() {
    this.getModuleDetail(this.module);
  }

  getModuleDetail(module: UtmModulesEnum) {
    this.utmModulesService.getModulesDetails({nameShort: module, serverId: this.serverId})
      .subscribe(response => {
        this.activatable = response.body.activatable;
        this.moduleDetail = response.body;
        this.loadingDetail = false;
      });
  }

  enableModule() {
    if (!this.moduleDetail.moduleActive) {
      const modalChecks = this.modalService.open(AppModuleChecksComponent,
        {centered: true});
      modalChecks.componentInstance.module = this.moduleDetail;
      modalChecks.componentInstance.serverId = this.serverId;
      modalChecks.componentInstance.checkResult.subscribe(statusCheck => {
        if (statusCheck) {
          this.changeModuleStatus(statusCheck);
        }
      });
    } else {
      const modalDeactivate = this.modalService.open(AppModuleDeactivateComponent, {centered: true});
      modalDeactivate.componentInstance.module = this.moduleDetail;
      modalDeactivate.componentInstance.serverId = this.serverId;
      modalDeactivate.componentInstance.disable.subscribe(statusCheck => {
        this.changeModuleStatus(false);
      });
    }
  }

  changeModuleStatus(status: boolean) {
    this.changingStatus = true;
    this.utmModulesService.setModuleActivate(status, this.moduleDetail.moduleName, this.serverId)
      .subscribe(response => {
        this.moduleDetail.moduleActive = response.moduleActive;
        this.changingStatus = false;
        this.navBehavior.$nav.next(true);
        this.moduleRefreshBehavior.$moduleChange.next(true);
        this.toastService.showSuccessBottom('Module ' + this.moduleDetail.moduleName +
          ' has been ' + (this.moduleDetail.moduleActive ? 'enabled' : 'disabled') + ' successfully');
      });
  }

}
