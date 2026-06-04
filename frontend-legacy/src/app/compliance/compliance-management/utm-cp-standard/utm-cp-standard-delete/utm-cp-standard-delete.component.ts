import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpStandardService} from '../../../shared/services/cp-standard.service';
import {ComplianceStandardType} from '../../../shared/type/compliance-standard.type';

@Component({
  selector: 'app-utm-cp-standard-delete',
  templateUrl: './utm-cp-standard-delete.component.html',
  styleUrls: ['./utm-cp-standard-delete.component.scss']
})
export class UtmCpStandardDeleteComponent implements OnInit {
  @Input() standard: ComplianceStandardType;
  @Output() standardDelete = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private cpStandardService: CpStandardService) {
  }

  ngOnInit() {
  }

  deleteStandard() {
    this.cpStandardService.delete(this.standard.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Standard deleted successfully');
        this.activeModal.close();
        this.standardDelete.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting standard',
          error.error.statusText);
      });
  }
}
