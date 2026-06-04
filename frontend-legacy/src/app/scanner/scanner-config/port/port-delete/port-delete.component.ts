import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {PortModel} from '../../../shared/model/port.model';
import {PortService} from '../shared/services/port.service';

@Component({
  selector: 'app-port-delete',
  templateUrl: './port-delete.component.html',
  styleUrls: ['./port-delete.component.scss']
})
export class PortDeleteComponent implements OnInit {
  @Input() port: PortModel;
  @Output() portDeleted = new EventEmitter<string>();

  constructor(
    public activeModal: NgbActiveModal,
    private portService: PortService,
    private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  delete() {
    this.portService.delete(this.port.uuid)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Credential deleted successfully');
        this.activeModal.close();
        this.portDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting port',
          'Error deleting port, please check your network and try again');
      });
  }

}
