import {Injectable} from '@angular/core';
import {NgbModal, NgbModalRef} from '@ng-bootstrap/ng-bootstrap';

@Injectable({
  providedIn: 'root'
})
export class ModalService {
  modalConfig: any;
  private isOpen = false;

  constructor(private modalService: NgbModal) {
  }

  open(component: any, data?: any, windowClass?: string): NgbModalRef {
    this.modalConfig = {
      backdrop: false,
      keyboard: true,
      centered: true
    };
    let modalRef;
    modalRef = this.modalService.open(component, this.modalConfig);

    if (data !== null || data !== undefined) {
      modalRef.componentInstance.data = data;
    }
    modalRef.result.then(
      result => {
        this.isOpen = false;
      },
      reason => {
        this.isOpen = false;
      }
    );
    return modalRef;
  }

  close() {
    this.modalService.dismissAll('close');
  }
}
