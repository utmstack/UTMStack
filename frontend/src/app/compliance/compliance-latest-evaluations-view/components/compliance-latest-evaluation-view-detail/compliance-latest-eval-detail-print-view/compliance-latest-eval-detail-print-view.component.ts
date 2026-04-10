import {HttpErrorResponse} from '@angular/common/http';
import {Component, OnInit} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {EMPTY, Observable} from 'rxjs';
import {catchError} from 'rxjs/operators';
import {UtmToastService} from '../../../../../shared/alert/utm-toast.service';
import {TimezoneFormatService} from '../../../../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../../../../shared/types/date-pipe-default-options';
import {ComplianceStatusExtendedEnum} from '../../../../shared/enums/compliance-status.enum';
import {CpControlConfigService} from '../../../../shared/services/cp-control-config.service';
import {ComplianceControlLatestEvaluationType} from '../../../../shared/type/compliance-control-latest-evaluation.type';

@Component({
  selector: 'app-compliance-detail-print-view',
  templateUrl: './compliance-latest-eval-detail-print-view.component.html',
  styleUrls: ['./compliance-latest-eval-detail-print-view.component.scss']
})
export class ComplianceLatestEvalDetailPrintViewComponent implements OnInit {
  reportId: number;
  control: ComplianceControlLatestEvaluationType;
  dateFormat$: Observable<DatePipeDefaultOptions>;

  protected readonly ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;

  constructor(private activatedRoute: ActivatedRoute,
              private cpControlConfigService: CpControlConfigService,
              private timezoneFormatService: TimezoneFormatService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
    this.activatedRoute.params.subscribe(params => {
      this.reportId = params.id;
      this.getTemplate();
    });
  }

  getTemplate() {
    this.cpControlConfigService.find(this.reportId)
      .pipe(
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError(
            'Error',
            'Unable to retrieve the report template. Please try again or contact support.'
          );
          return EMPTY;
        })
      )
      .subscribe(response => {
        this.control = response.body;
      });
  }

}
