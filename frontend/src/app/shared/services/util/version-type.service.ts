import {Injectable} from '@angular/core';
import {BehaviorSubject} from 'rxjs';
import {ModalConfirmationComponent} from '../../components/utm/util/modal-confirmation/modal-confirmation.component';
import {NgbModal} from "@ng-bootstrap/ng-bootstrap";

export const EnterpriseFeatures = [
  'AUTH_WITH_PROVIDERS_MODULE',
];


export enum VersionType {
  COMMUNITY = 'COMMUNITY',
  ENTERPRISE = 'ENTERPRISE',
}

@Injectable({
  providedIn: 'root'
})
export class VersionTypeService {
 private versionTypeBehavior = new BehaviorSubject<VersionType>(VersionType.COMMUNITY);
 versionType$ = this.versionTypeBehavior.asObservable();

 constructor(private modalService: NgbModal) {}

 changeVersionType(versionType: VersionType) {
   this.versionTypeBehavior.next(versionType);
 }

 versionType(): VersionType {
   return this.versionTypeBehavior.getValue();
 }

  showVersionInfo() {
    const modalSource = this.modalService.open(ModalConfirmationComponent, { centered: true });

    modalSource.componentInstance.header = 'Enterprise Feature';
    modalSource.componentInstance.message =
      'This feature is available only in the Enterprise edition of the platform. ' +
      'For more information about upgrading or accessing this functionality, please contact our support team at ' +
      '<a href="mailto:support@services.utmstack.com">support@services.utmstack.com</a>.';
    modalSource.componentInstance.confirmBtnText = 'OK';
    modalSource.componentInstance.confirmBtnIcon = 'icon-info';
    modalSource.componentInstance.confirmBtnType = 'default';
    modalSource.componentInstance.hideBtnCancel = true;

    modalSource.result.then(() => {
      // optional callback logic
    });
  }


}
