import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpReportsService} from '../../../shared/services/cp-reports.service';
import {ComplianceReportType} from '../../../shared/type/compliance-report.type';

@Component({
  selector: 'app-utm-cp-report-delete',
  templateUrl: './utm-cp-report-delete.component.html',
  styleUrls: ['./utm-cp-report-delete.component.scss']
})
export class UtmCpReportDeleteComponent implements OnInit {
  @Input() report: ComplianceReportType;
  @Output() reportDelete = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private cpReportsService: CpReportsService) {
  }

  ngOnInit() {
  }

  deleteStandard() {
    this.cpReportsService.delete(this.report.id)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Report deleted successfully');
        this.activeModal.close();
        this.reportDelete.emit('deleted');
      }, (error) => {
        this.utmToastService.showError('Error deleting report',
          error.error.statusText);
      });
  }


}
