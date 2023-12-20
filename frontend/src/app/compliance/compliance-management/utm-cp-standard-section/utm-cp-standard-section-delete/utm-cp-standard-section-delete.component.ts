import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpStandardSectionService} from '../../../shared/services/cp-standard-section.service';
import {ComplianceStandardSectionType} from '../../../shared/type/compliance-standard-section.type';

@Component({
  selector: 'app-utm-cp-standard-section-delete',
  templateUrl: './utm-cp-standard-section-delete.component.html',
  styleUrls: ['./utm-cp-standard-section-delete.component.scss']
})
export class UtmCpStandardSectionDeleteComponent implements OnInit {
  @Input() standardSection: ComplianceStandardSectionType;
  @Output() standardSectionDelete = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private cpStandardSectionService: CpStandardSectionService) {
  }

  ngOnInit() {
  }

  deleteStandard() {
    this.cpStandardSectionService.delete(this.standardSection.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Standard section deleted successfully');
        this.activeModal.close();
        this.standardSectionDelete.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting standard section',
          error.error.statusText);
      });
  }

}
