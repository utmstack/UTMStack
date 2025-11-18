import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {
  ModalConfirmationComponent
} from '../../../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {
  EnterpriseFeatures,
  VersionType,
  VersionTypeService
} from '../../../../shared/services/util/version-type.service';
import {UtmModulesEnum} from '../../enum/utm-module.enum';
import {UtmModuleType} from '../../type/utm-module.type';

@Component({
  selector: 'app-app-module-card',
  templateUrl: './app-module-card.component.html',
  styleUrls: ['./app-module-card.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AppModuleCardComponent implements OnInit, OnDestroy {

  constructor(private versionTypeService: VersionTypeService,
              private modalService: NgbModal) {
  }
  @Input() module: UtmModuleType;
  @Output() showModuleIntegration = new EventEmitter<UtmModuleType>();
  versionType = VersionType;
  version: VersionType;
  modules = UtmModulesEnum;
  destroy$: Subject<void> = new Subject<void>();
  ModulesEnterprise = EnterpriseFeatures;

  ngOnInit() {
    this.versionTypeService.versionType$
    .pipe(takeUntil(this.destroy$))
      .subscribe(versionType => this.version = versionType);
  }

  showIntegration() {
    this.showModuleIntegration.emit(this.module);
  }

  showMessage() {
    const modalSource = this.modalService.open(ModalConfirmationComponent, {centered: true});

    modalSource.componentInstance.header = 'Enterprise integration';
    modalSource.componentInstance.message = 'This integration is only available in the Enterprise version of the platform. ' +
      'If you are interested in accessing this feature or need more information, please contact our support team at ' +
      '<a href="mailto:support@utmstack.com">support@services.utmstack.com</a>.';
    modalSource.componentInstance.confirmBtnText = 'Accept';
    modalSource.componentInstance.confirmBtnIcon = 'icon-cog3';
    modalSource.componentInstance.confirmBtnType = 'default';
    modalSource.componentInstance.hideBtnCancel = true;
    modalSource.result.then(() => {

    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
