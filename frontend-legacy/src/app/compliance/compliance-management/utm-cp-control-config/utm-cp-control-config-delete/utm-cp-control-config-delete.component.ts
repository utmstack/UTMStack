import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpControlConfigService} from '../../../shared/services/cp-control-config.service';
import {ComplianceControlType} from '../../../shared/type/compliance-control.type';

@Component({
  selector: 'app-utm-cp-control-config-delete',
  templateUrl: './utm-cp-control-config-delete.component.html',
  styleUrls: ['./utm-cp-control-config-delete.component.scss']
})
export class UtmCpControlConfigDeleteComponent implements OnInit {
  @Input() control: ComplianceControlType;
  @Output() controlDelete = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private cpControlConfigService: CpControlConfigService) {
  }

  ngOnInit() {

  }

  deleteControl() {
    this.cpControlConfigService.delete(this.control.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Control deleted successfully');
        this.activeModal.close();
        this.controlDelete.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting control',
          error.error.statusText);
      });
  }
}
