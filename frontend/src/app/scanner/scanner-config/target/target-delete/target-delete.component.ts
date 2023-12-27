import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {TargetModel} from '../../../shared/model/target.model';
import {TargetService} from '../shared/services/target.service';

@Component({
  selector: 'app-target-delete',
  templateUrl: './target-delete.component.html',
  styleUrls: ['./target-delete.component.scss']
})
export class TargetDeleteComponent implements OnInit {
  @Input() target: TargetModel;
  @Output() targetDeleted = new EventEmitter<string>();

  constructor(
    public activeModal: NgbActiveModal,
    private targetService: TargetService,
    private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteTarget() {
    this.targetService.delete(this.target.uuid)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Target deleted successfully');
        this.activeModal.close();
        this.targetDeleted.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting target',
          error.error.statusText);
      });
  }
}
