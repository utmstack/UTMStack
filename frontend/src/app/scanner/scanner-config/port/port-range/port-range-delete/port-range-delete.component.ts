import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {PortRangeModel} from '../../../../shared/model/port-range.model';
import {PortRangeService} from '../shared/services/port-range.service';

@Component({
  selector: 'app-port-range-delete',
  templateUrl: './port-range-delete.component.html',
  styleUrls: ['./port-range-delete.component.scss']
})
export class PortRangeDeleteComponent implements OnInit {
  @Input() portRange: PortRangeModel;
  @Output() portRangeDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private portRangeService: PortRangeService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  delete() {
    this.portRangeService.delete(this.portRange.uuid).subscribe(portCreated => {
      this.portRangeDeleted.emit('success');
      this.activeModal.dismiss();
      this.utmToastService.showSuccessBottom('Port range created successfully');
    }, error1 => {
      this.utmToastService.showError('Error creating port range',
        error1.error.statusText);
    });
  }

}
