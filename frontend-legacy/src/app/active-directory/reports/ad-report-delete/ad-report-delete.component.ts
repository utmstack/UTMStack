import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {AdReportService} from '../shared/services/ad-report.service';
import {AdReportType} from '../shared/type/ad-report.type';

@Component({
  selector: 'app-ad-report-delete',
  templateUrl: './ad-report-delete.component.html',
  styleUrls: ['./ad-report-delete.component.scss']
})
export class AdReportDeleteComponent implements OnInit {
  @Input() adReport: AdReportType;
  @Output() adReportDeleted = new EventEmitter<string>();

  constructor(public activeModal: NgbActiveModal,
              private adReportService: AdReportService,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteReport() {
    this.adReportService.delete(this.adReport.id).subscribe(() => {
      this.adReportDeleted.emit('adReport deleted');
      this.utmToastService.showSuccessBottom('Report deleted successfully');
      this.activeModal.close();
    }, error1 => {
      this.utmToastService.showError('Error deleting report',
        'Error deleting adReport, check your network');
    });
  }
}
