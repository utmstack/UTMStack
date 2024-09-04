import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {NavBehavior} from '../../../../shared/behaviors/nav.behavior';
import {ModuleChangeStatusBehavior} from '../../behavior/module-change-status.behavior';
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
export class AppModuleActivateButtonComponent implements OnInit, OnDestroy {
  @Input() performPreAction = false;
  @Input() module: UtmModulesEnum;
  @Input() type: string;
  @Input() disabled = false;
  @Input() serverId: number;
  @Output() disableModuleClicked = new EventEmitter<void>();
  loadingDetail = true;
  moduleDetail: UtmModuleType;
  changingStatus: any;
  activatable: boolean;
  destroy$: Subject<void> = new Subject<void>();

  constructor(private utmModulesService: UtmModulesService,
              public modalService: NgbModal,
              private toastService: UtmToastService,
              private moduleRefreshBehavior: ModuleRefreshBehavior,
              private navBehavior: NavBehavior,
              private moduleChangeStatusBehavior: ModuleChangeStatusBehavior) {
  }

  ngOnInit() {
    this.moduleChangeStatusBehavior.moduleStatus$
        .pipe(filter(value => !!value),
              takeUntil(this.destroy$))
        .subscribe( value => {
          this.performPreAction = value.hasPreAction;
          if (value.status != null && this.moduleDetail && (this.moduleDetail.moduleActive && !value.status ||
                this.moduleDetail && !this.moduleDetail.moduleActive && value.status)) {
            this.changeModuleStatus(value.status);
          }
        });
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
          this.changeModuleStatus(statusCheck, true);
        }
      });
    } else {
      const modalDeactivate = this.modalService.open(AppModuleDeactivateComponent, {centered: true});
      modalDeactivate.componentInstance.module = this.moduleDetail;
      modalDeactivate.componentInstance.serverId = this.serverId;
      modalDeactivate.componentInstance.disable.subscribe(statusCheck => {
        this.changeModuleStatus(false, true);
      });
    }
  }

  changeModuleStatus(status: boolean, fromOnclick = false) {
    if (!this.performPreAction || (this.performPreAction && status)) {
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
    } else {
      if (fromOnclick && !status) {
        this.disableModuleClicked.emit();
      }
    }
  }

  ngOnDestroy() {
    this.moduleChangeStatusBehavior.setStatus(null, null);
    this.destroy$.next();
    this.destroy$.complete();
  }

}
