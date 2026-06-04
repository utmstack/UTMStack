import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {CpReportsService} from '../../shared/services/cp-reports.service';
import {ComplianceReportType} from '../../shared/type/compliance-report.type';

@Component({
  selector: 'app-utm-cp-import',
  templateUrl: './utm-cp-import.component.html',
  styleUrls: ['./utm-cp-import.component.scss']
})
export class UtmCpImportComponent implements OnInit {
  @Output() reportsImported = new EventEmitter<string>();
  complianceReports: ComplianceReportType[] = [];
  step = 1;
  stepCompleted: number[] = [];
  importing = false;
  imported = 0;
  override = false;

  constructor(private utmToast: UtmToastService,
              public activeModal: NgbActiveModal,
              private complianceReportService: CpReportsService) {
  }

  ngOnInit() {
  }

  onFileImportLoad($event) {
    this.complianceReports = $event;
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  backStep() {
    this.stepCompleted.pop();
    this.step -= 1;
  }

  import() {
    const interval = setInterval(() => {
      this.imported += 1;
    }, 2000);
    this.importing = true;
    this.complianceReportService.import({
      override: this.override,
      reports: this.complianceReports
    }).subscribe(value => {
      this.importing = false;
      clearInterval(interval);
      this.utmToast.showSuccessBottom('Compliance imported successfully');
      this.reportsImported.emit('success');
      this.activeModal.close();
    }, error => {
      clearInterval(interval);
      this.importing = false;
      this.utmToast.showError('Error', 'Error importing compliance reports');
    });
  }

}
