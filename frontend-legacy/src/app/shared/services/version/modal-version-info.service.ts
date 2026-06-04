import {Injectable} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {ModalConfirmationComponent} from '../../components/utm/util/modal-confirmation/modal-confirmation.component';

@Injectable({
  providedIn: 'root'
})
export class ModalVersionInfoService {

  constructor(private modalService: NgbModal) {}

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
