import {Component, Input, OnInit} from '@angular/core';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ExportPdfService} from '../../../../shared/services/util/export-pdf.service';
import {TimezoneFormatService} from '../../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../../shared/types/date-pipe-default-options';
import {ComplianceStatusExtendedEnum} from '../../../shared/enums/compliance-status.enum';
import {ComplianceControlLatestEvaluationType} from '../../../shared/type/compliance-control-latest-evaluation.type';

@Component({
  selector: 'app-compliance-latest-evaluation-view-detail',
  templateUrl: './compliance-latest-evaluation-view-detail.component.html',
  styleUrls: ['./compliance-latest-evaluation-view-detail.component.css']
})
export class ComplianceLatestEvaluationViewDetailComponent implements OnInit {
  @Input() control: ComplianceControlLatestEvaluationType;
  dateFormat$: Observable<DatePipeDefaultOptions>;
  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;

  constructor(private timezoneFormatService: TimezoneFormatService,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
  }

  exportToPdf() {
    this.spinner.show('buildPrintPDF');
    const url = '/compliance/evaluation-detail-print-view/' + this.control.id;
    const fileName = this.control.controlName.replace(/ /g, '_');
    this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.exportPdfService.handlePdfResponse(response));
    }, error => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.toastService.showError('Error', 'An error occurred while creating a PDF.'));
    });
  }
}
